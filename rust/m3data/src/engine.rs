//! Grow a simulation from `inputs.json` and append each time step.
//!
//! Path shells come from [`m3path::PathGrowth`]. Each event places those points
//! at its center, the same way `EventDb.increaseMaxNodeTime` does. Event nodes
//! copy the path-node mask at the moment they are created and do not follow
//! later path-node updates. A `From` link on an event node is the earlier event
//! node, not the path node.

use std::collections::{BTreeMap, HashMap};

use m3path::{
    get_connection_state, restored_link, ConnectionState, ConnectionsState, LivePathNode,
    PathGrowth, NB_CONNECTIONS,
};
use m3point::{NodeHandle, Point, PointPackData, ORIGIN};

use crate::error::{Error, Result};
use crate::inputs::{resolve_events, validate_inputs, ResolvedEvent, SimulationInputs};
use crate::rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow};
use crate::store::SimulationStore;

pub struct RunStats {
    pub resumed: bool,
    pub max_time_steps: i32,
    pub time: i32,
    pub events: usize,
    pub nodes: usize,
    pub steps_computed: i32,
}

struct CtxState {
    persisted_max_dist: i32,
    growth: PathGrowth,
}

struct EventState {
    id: i32,
    path_ctx_id: i32,
    creation_time: i32,
    color: i32,
    center: Point,
    max_node_time: i32,
    persisted_max_node_time: i32,
}

struct World {
    contexts: Vec<CtxState>,
    events: Vec<EventState>,
    point_ids: HashMap<(i32, i32, i32), i64>,
    next_point_id: i64,
    next_path_node_id: i64,
    next_node_id: i64,
    event_node_ids: HashMap<(i32, i64), i64>,
    fresh_points: Vec<PointRow>,
    fresh_nodes: Vec<NodeRow>,
}

struct Draft {
    event_id: i32,
    path_node_id: i64,
    trio_id: i32,
    d: i32,
    mask: i32,
    abs: Point,
    links: [Option<i64>; 3],
}

pub async fn run_simulation(
    store: &mut SimulationStore,
    inputs: &SimulationInputs,
) -> Result<RunStats> {
    validate_inputs(inputs)?;
    let pack = PointPackData::shared();
    let resumed = !store.nodes().is_empty();
    if !resumed {
        if !store.points().is_empty()
            || !store.path_nodes().is_empty()
            || !store.events().is_empty()
            || !store.path_contexts().is_empty()
        {
            return Err(Error::new(
                "simulation has saved rows but no nodes; refusing to seed over a partial write",
            ));
        }
        let specs = resolve_events(inputs)?;
        if specs.is_empty() {
            return Err(Error::new("inputs.json does not create any events"));
        }
        let mut world = seed(&specs)?;
        world.persist(store).await?;
    }
    let mut world = load(store)?;
    let mut steps = 0;
    while world.grow_one(pack, inputs.max_time_steps)? {
        world.persist(store).await?;
        steps += 1;
    }
    if world.needs_persist() {
        world.persist(store).await?;
    }
    let time = store
        .nodes()
        .iter()
        .map(|node| node.creation_time)
        .max()
        .unwrap_or(0);
    Ok(RunStats {
        resumed,
        max_time_steps: inputs.max_time_steps,
        time,
        events: store.events().len(),
        nodes: store.nodes().len(),
        steps_computed: steps,
    })
}

fn seed(specs: &[ResolvedEvent]) -> Result<World> {
    let pack = PointPackData::shared();
    let mut world = World {
        contexts: Vec::new(),
        events: Vec::new(),
        point_ids: HashMap::new(),
        next_point_id: 1,
        next_path_node_id: 1,
        next_node_id: 1,
        event_node_ids: HashMap::new(),
        fresh_points: Vec::new(),
        fresh_nodes: Vec::new(),
    };
    world.intern_new_sorted(&[ORIGIN]);

    let mut ctx_for: HashMap<(i32, i32), i32> = HashMap::new();
    let mut next_ctx = 1_i32;
    for spec in specs {
        let key = (spec.growth_ctx_id, spec.growth_offset);
        if ctx_for.contains_key(&key) {
            continue;
        }
        let id = next_ctx;
        next_ctx += 1;
        let mut growth = PathGrowth::new_root(pack, id, spec.growth_ctx_id, spec.growth_offset)?;
        growth.nodes[0].id = world.next_path_node_id;
        world.next_path_node_id += 1;
        ctx_for.insert(key, id);
        world.contexts.push(CtxState {
            persisted_max_dist: -1,
            growth,
        });
    }

    let mut next_event = 1_i32;
    for spec in specs {
        let ctx_id = ctx_for[&(spec.growth_ctx_id, spec.growth_offset)];
        let event_id = next_event;
        next_event += 1;
        world.intern_new_sorted(&[spec.point]);
        let point_id = world.point_id(spec.point);
        let root = world
            .contexts
            .iter()
            .find(|ctx| ctx.growth.id == ctx_id)
            .expect("context just inserted")
            .growth
            .nodes
            .iter()
            .find(|node| node.d == 0)
            .expect("root");
        let root_id = root.id;
        let trio_id = i32::from(root.connections.trio_id);
        let mask = i32::from(root.connections.connection_mask);
        let node_id = world.next_node_id;
        world.next_node_id += 1;
        world.event_node_ids.insert((event_id, root_id), node_id);
        world.fresh_nodes.push(NodeRow {
            id: node_id,
            event_id,
            path_node_id: root_id,
            trio_id,
            point_id,
            d: 0,
            creation_time: spec.creation_time,
            connection_mask: mask,
            node1: None,
            node2: None,
            node3: None,
        });
        world.events.push(EventState {
            id: event_id,
            path_ctx_id: ctx_id,
            creation_time: spec.creation_time,
            color: i32::from(spec.color as u8),
            center: spec.point,
            max_node_time: spec.creation_time,
            persisted_max_node_time: i32::MIN,
        });
    }
    Ok(world)
}

fn load(store: &SimulationStore) -> Result<World> {
    let mut point_ids = HashMap::new();
    let mut id_to_point = HashMap::new();
    for row in store.points() {
        point_ids.insert(row.xyz(), row.id);
        id_to_point.insert(row.id, row.point());
    }

    let mut contexts = Vec::new();
    for row in store.path_contexts() {
        let mut nodes = Vec::new();
        for path in store
            .path_nodes()
            .iter()
            .filter(|node| node.path_ctx_id == row.id)
        {
            let point = id_to_point.get(&path.point_id).copied().ok_or_else(|| {
                Error::new(format!(
                    "path node {} points at missing point {}",
                    path.id, path.point_id
                ))
            })?;
            let builder =
                NodeHandle::from_path_builder_index(path.path_builders_id, path.path_builder_idx)?;
            let mask = path.connection_mask as u16;
            let stored = [path.path_node1, path.path_node2, path.path_node3];
            let mut link_ids = [0_i64; NB_CONNECTIONS];
            for conn in 0..NB_CONNECTIONS {
                let state = get_connection_state(mask, conn as i32);
                link_ids[conn] = restored_link(state, stored[conn]);
            }
            nodes.push(LivePathNode {
                id: path.id,
                point,
                d: path.d,
                builder,
                connections: ConnectionsState {
                    trio_id: path.trio_id as u8,
                    connection_mask: mask,
                    link_ids,
                },
                link_at: [None; NB_CONNECTIONS],
                dirty: false,
            });
        }
        let mut growth =
            PathGrowth::from_stored(row.id, row.growth_ctx_id, row.growth_offset, nodes)?;
        let id_to_index: HashMap<i64, usize> = growth
            .nodes
            .iter()
            .enumerate()
            .map(|(index, node)| (node.id, index))
            .collect();
        for node in &mut growth.nodes {
            for conn in 0..NB_CONNECTIONS {
                let link = node.connections.link_ids[conn];
                if link > 0 {
                    node.link_at[conn] = id_to_index.get(&link).copied();
                }
            }
        }
        contexts.push(CtxState {
            persisted_max_dist: row.max_dist,
            growth,
        });
    }

    let mut event_node_ids = HashMap::new();
    for node in store.nodes() {
        event_node_ids.insert((node.event_id, node.path_node_id), node.id);
    }

    let mut events = Vec::new();
    for row in store.events() {
        let root = store
            .nodes()
            .iter()
            .find(|node| node.event_id == row.id && node.d == 0)
            .ok_or_else(|| Error::new(format!("event {} has no root node", row.id)))?;
        let center = id_to_point.get(&root.point_id).copied().ok_or_else(|| {
            Error::new(format!(
                "event {} root points at missing point {}",
                row.id, root.point_id
            ))
        })?;
        let max_from_nodes = store
            .nodes()
            .iter()
            .filter(|node| node.event_id == row.id)
            .map(|node| node.creation_time)
            .max()
            .unwrap_or(row.creation_time);
        events.push(EventState {
            id: row.id,
            path_ctx_id: row.path_ctx_id,
            creation_time: row.creation_time,
            color: row.color,
            center,
            max_node_time: max_from_nodes,
            persisted_max_node_time: row.max_node_time,
        });
    }

    let next_point_id = store.points().iter().map(|row| row.id).max().unwrap_or(0) + 1;
    let next_path_node_id = store
        .path_nodes()
        .iter()
        .map(|row| row.id)
        .max()
        .unwrap_or(0)
        + 1;
    let next_node_id = store.nodes().iter().map(|row| row.id).max().unwrap_or(0) + 1;

    Ok(World {
        contexts,
        events,
        point_ids,
        next_point_id,
        next_path_node_id,
        next_node_id,
        event_node_ids,
        fresh_points: Vec::new(),
        fresh_nodes: Vec::new(),
    })
}

impl World {
    fn grow_one(&mut self, pack: &PointPackData, max_time_steps: i32) -> Result<bool> {
        let next_time = self
            .events
            .iter()
            .filter(|event| event.max_node_time < max_time_steps)
            .map(|event| event.max_node_time + 1)
            .filter(|time| *time <= max_time_steps)
            .min();
        let Some(next_time) = next_time else {
            return Ok(false);
        };
        let due: Vec<usize> = self
            .events
            .iter()
            .enumerate()
            .filter(|(_, event)| {
                event.creation_time < next_time && event.max_node_time + 1 == next_time
            })
            .map(|(index, _)| index)
            .collect();
        if due.is_empty() {
            return Err(Error::new(format!(
                "time {next_time} is still below max_time_steps but no event can grow"
            )));
        }

        let mut need: BTreeMap<i32, i32> = BTreeMap::new();
        for index in &due {
            let event = &self.events[*index];
            let distance = next_time - event.creation_time;
            need.entry(event.path_ctx_id)
                .and_modify(|current| *current = (*current).max(distance))
                .or_insert(distance);
        }
        for ctx in &mut self.contexts {
            if let Some(distance) = need.get(&ctx.growth.id) {
                ctx.growth.grow_to(pack, *distance)?;
            }
        }

        let mut fresh = Vec::new();
        for (ctx_index, ctx) in self.contexts.iter().enumerate() {
            for (node_index, node) in ctx.growth.nodes.iter().enumerate() {
                if node.id <= 0 {
                    fresh.push((ctx.growth.id, ctx_index, node_index, node.point));
                }
            }
        }
        fresh.sort_by_key(|(ctx_id, _, _, point)| (*ctx_id, *point));
        for (_, ctx_index, node_index, _) in fresh {
            let id = self.next_path_node_id;
            self.next_path_node_id += 1;
            let node = &mut self.contexts[ctx_index].growth.nodes[node_index];
            node.id = id;
            node.dirty = true;
        }
        for ctx in &mut self.contexts {
            ctx.growth.backfill_next_links();
        }

        let mut drafts = Vec::new();
        for event_index in &due {
            let event = &self.events[*event_index];
            let distance = next_time - event.creation_time;
            let ctx = self
                .contexts
                .iter()
                .find(|ctx| ctx.growth.id == event.path_ctx_id)
                .ok_or_else(|| Error::new(format!("missing path context {}", event.path_ctx_id)))?;
            for node in ctx.growth.nodes.iter().filter(|node| node.d == distance) {
                drafts.push(Draft {
                    event_id: event.id,
                    path_node_id: node.id,
                    trio_id: i32::from(node.connections.trio_id),
                    d: node.d,
                    mask: i32::from(node.connections.connection_mask),
                    abs: event.center.add(node.point),
                    links: event_links(node, event.id, &self.event_node_ids)?,
                });
            }
        }
        drafts.sort_by_key(|draft| (draft.event_id, draft.path_node_id));

        let mut coords = Vec::new();
        for ctx in &self.contexts {
            for node in &ctx.growth.nodes {
                if node.dirty {
                    coords.push(node.point);
                }
            }
        }
        for draft in &drafts {
            coords.push(draft.abs);
        }
        self.intern_new_sorted(&coords);

        for draft in drafts {
            let id = self.next_node_id;
            self.next_node_id += 1;
            self.event_node_ids
                .insert((draft.event_id, draft.path_node_id), id);
            let point_id = self.point_id(draft.abs);
            self.fresh_nodes.push(NodeRow {
                id,
                event_id: draft.event_id,
                path_node_id: draft.path_node_id,
                trio_id: draft.trio_id,
                point_id,
                d: draft.d,
                creation_time: next_time,
                connection_mask: draft.mask,
                node1: draft.links[0],
                node2: draft.links[1],
                node3: draft.links[2],
            });
        }
        for index in due {
            self.events[index].max_node_time = next_time;
        }
        Ok(true)
    }

    fn needs_persist(&self) -> bool {
        if !self.fresh_points.is_empty() || !self.fresh_nodes.is_empty() {
            return true;
        }
        if self.contexts.iter().any(|ctx| {
            ctx.growth.max_dist() != ctx.persisted_max_dist
                || ctx.growth.nodes.iter().any(|node| node.dirty)
        }) {
            return true;
        }
        self.events
            .iter()
            .any(|event| event.max_node_time != event.persisted_max_node_time)
    }

    async fn persist(&mut self, store: &mut SimulationStore) -> Result<()> {
        let coords: Vec<Point> = self
            .contexts
            .iter()
            .flat_map(|ctx| {
                ctx.growth
                    .nodes
                    .iter()
                    .filter(|node| node.dirty)
                    .map(|node| node.point)
            })
            .collect();
        self.intern_new_sorted(&coords);

        if !self.fresh_points.is_empty() {
            store.append_points(&self.fresh_points).await?;
            self.fresh_points.clear();
        }

        let ctx_rows: Vec<PathContextRow> = self
            .contexts
            .iter()
            .filter(|ctx| ctx.growth.max_dist() != ctx.persisted_max_dist)
            .map(|ctx| PathContextRow {
                id: ctx.growth.id,
                growth_ctx_id: ctx.growth.growth_ctx_id,
                growth_offset: ctx.growth.growth_offset,
                path_builders_id: ctx
                    .growth
                    .nodes
                    .iter()
                    .find(|node| node.d == 0)
                    .map(|node| node.builder.cube_id()),
                max_dist: ctx.growth.max_dist(),
            })
            .collect();
        store.append_path_contexts(&ctx_rows).await?;
        for ctx in &mut self.contexts {
            ctx.persisted_max_dist = ctx.growth.max_dist();
        }

        let mut path_rows = Vec::new();
        for ctx in &self.contexts {
            for node in &ctx.growth.nodes {
                if !node.dirty {
                    continue;
                }
                let point_id = self.point_id(node.point);
                let links = (0..NB_CONNECTIONS)
                    .map(|conn| {
                        m3path::stored_link(
                            node.connections.connection_state(conn as i32),
                            node.connections.link_ids[conn],
                        )
                    })
                    .collect::<Vec<_>>();
                path_rows.push(PathNodeRow {
                    id: node.id,
                    path_ctx_id: ctx.growth.id,
                    path_builders_id: node.builder.cube_id(),
                    path_builder_idx: node.builder.path_builder_index(),
                    trio_id: i32::from(node.connections.trio_id),
                    point_id,
                    d: node.d,
                    connection_mask: i32::from(node.connections.connection_mask),
                    path_node1: links[0],
                    path_node2: links[1],
                    path_node3: links[2],
                });
            }
        }
        store.append_path_nodes(&path_rows).await?;
        for ctx in &mut self.contexts {
            for node in &mut ctx.growth.nodes {
                node.dirty = false;
            }
        }

        if !self.fresh_nodes.is_empty() {
            store.append_nodes(&self.fresh_nodes).await?;
            self.fresh_nodes.clear();
        }

        let event_rows: Vec<EventRow> = self
            .events
            .iter()
            .filter(|event| event.max_node_time != event.persisted_max_node_time)
            .map(|event| EventRow {
                id: event.id,
                path_ctx_id: event.path_ctx_id,
                creation_time: event.creation_time,
                color: event.color,
                end_time: event.creation_time,
                max_node_time: event.max_node_time,
            })
            .collect();
        store.append_events(&event_rows).await?;
        for event in &mut self.events {
            event.persisted_max_node_time = event.max_node_time;
        }
        Ok(())
    }

    fn intern_new_sorted(&mut self, points: &[Point]) {
        let mut fresh: Vec<(i32, i32, i32)> = points
            .iter()
            .map(|point| (point.0[0], point.0[1], point.0[2]))
            .filter(|key| !self.point_ids.contains_key(key))
            .collect();
        fresh.sort();
        fresh.dedup();
        for key in fresh {
            let id = self.next_point_id;
            self.next_point_id += 1;
            self.point_ids.insert(key, id);
            self.fresh_points.push(PointRow {
                id,
                x: key.0,
                y: key.1,
                z: key.2,
            });
        }
    }

    fn point_id(&self, point: Point) -> i64 {
        self.point_ids[&(point.0[0], point.0[1], point.0[2])]
    }
}

fn event_links(
    node: &LivePathNode,
    event_id: i32,
    event_node_ids: &HashMap<(i32, i64), i64>,
) -> Result<[Option<i64>; 3]> {
    let mut links = [None; 3];
    for conn in 0..NB_CONNECTIONS {
        match node.connections.connection_state(conn as i32) {
            // Go stores NULL for these. Next is intentionally not linked on event nodes.
            ConnectionState::NotSet | ConnectionState::Blocked | ConnectionState::Next => {}
            ConnectionState::From => {
                let path_from = node.connections.link_ids[conn];
                let Some(id) = event_node_ids.get(&(event_id, path_from)).copied() else {
                    return Err(Error::new(format!(
                        "event {event_id} from-link to path node {path_from} has no event node"
                    )));
                };
                links[conn] = Some(id);
            }
        }
    }
    Ok(links)
}
