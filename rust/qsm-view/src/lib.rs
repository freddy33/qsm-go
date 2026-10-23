//! Map a saved simulation onto the markers, links, and axes `qsm-view` draws.
//!
//! The Bevy window is the `qsm-view` binary (`--features view`). These types
//! stay free of a window so the mapping can be tested with `cargo test`.

use std::collections::{HashMap, HashSet};
use std::path::Path;

use m3data::{latest_run, load_inputs, EventRow, NodeRow, PointRow, SimulationStore};
use m3path::{get_connection_state, ConnectionState, NB_CONNECTIONS};

/// Go `AxeExtraLength` in `ui/m3gl/m3gl.go`. Axes run to `max_coord` plus this.
pub const AXE_EXTRA: i32 = 3;
/// Lattice lines are drawn on multiples of this, through the origin planes.
pub const GRID_STEP: i32 = 3;
/// Sphere radius for a node that is not an event root.
pub const NODE_RADIUS: f32 = 0.2;
/// Event roots (`d == 0`) are larger than the nodes that grow from them.
pub const ORIGIN_RADIUS: f32 = 0.5;
/// Selected spheres are drawn at this multiple of their radius.
pub const SELECTED_SCALE: f32 = 1.45;
/// Shade for a node that was not created at the time currently shown.
/// Matches `m3gl::DEFAULT_OLD_EVENT_DIMMER`.
pub const INACTIVE_DIMMER: f32 = 0.3;
/// Red, green, blue, and yellow. A bit set in the mask is drawn.
pub const ALL_EVENT_COLORS: u8 = 1 | 2 | 4 | 8;

/// One event node, already joined to its point and event color.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SceneNode {
    pub id: i64,
    pub position: [i32; 3],
    pub color: u8,
    pub creation_time: i32,
    /// True for the root of an event (`nodes.d == 0`).
    pub origin: bool,
    /// Path distance from the event root. Roots are `0`.
    pub d: i32,
    /// `events.id` of the event this node belongs to.
    pub event_id: i32,
    /// `events.path_ctx_id`, the path this event grows on.
    pub path_ctx_id: i32,
    /// Earlier event-node ids stored in a `From` link on this node.
    pub from: Vec<i64>,
}

/// What the window is currently willing to draw.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ViewFilter {
    pub time: i32,
    pub color_mask: u8,
    pub show_lines: bool,
    pub show_axes: bool,
    /// Draw only nodes created at `time`, plus the selected node's from-chain.
    pub latest_only: bool,
    /// Event-node id chosen by a click, if any.
    pub selected: Option<i64>,
    /// Bumped whenever a field above changes, so the window can rebuild.
    pub revision: u64,
}

impl ViewFilter {
    pub fn at_time(time: i32) -> Self {
        Self {
            time,
            color_mask: ALL_EVENT_COLORS,
            show_lines: true,
            show_axes: true,
            latest_only: false,
            selected: None,
            revision: 1,
        }
    }
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Marker {
    pub node_id: i64,
    pub position: [f32; 3],
    pub color: u8,
    pub origin: bool,
    /// `creation_time` equals the filter time. Inactive nodes are drawn darker.
    pub active: bool,
    pub selected: bool,
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Segment {
    pub a: [f32; 3],
    pub b: [f32; 3],
    /// `1` / `2` / `4` for the X / Y / Z axis, `0` for a lattice line, or the
    /// earlier node's event color for a connection.
    pub color: u8,
}

#[derive(Clone, Debug, PartialEq, Default)]
pub struct Frame {
    pub markers: Vec<Marker>,
    pub segments: Vec<Segment>,
    pub axes: Vec<Segment>,
}

impl Frame {
    pub fn node_count(&self) -> usize {
        self.markers.len()
    }

    pub fn origin_count(&self) -> usize {
        self.markers.iter().filter(|m| m.origin).count()
    }

    pub fn line_count(&self) -> usize {
        self.segments.len()
    }
}

/// sRGB triples for event colors `1` red, `2` green, `4` blue, `8` yellow.
pub fn color_rgb(color: u8) -> [f32; 3] {
    match color {
        1 => [0.86, 0.16, 0.16],
        2 => [0.16, 0.71, 0.24],
        4 => [0.22, 0.38, 0.90],
        8 => [0.90, 0.78, 0.16],
        _ => [0.45, 0.47, 0.50],
    }
}

/// Event color, dimmed when the node was not created at the displayed time.
pub fn display_rgb(color: u8, active: bool) -> [f32; 3] {
    let [r, g, b] = color_rgb(color);
    if active {
        [r, g, b]
    } else {
        [
            r * INACTIVE_DIMMER,
            g * INACTIVE_DIMMER,
            b * INACTIVE_DIMMER,
        ]
    }
}

/// [`display_rgb`], pulled toward white when the node is selected.
pub fn shade_rgb(color: u8, active: bool, selected: bool) -> [f32; 3] {
    let [r, g, b] = display_rgb(color, active);
    if selected {
        const MIX: f32 = 0.45;
        [
            r + (1.0 - r) * MIX,
            g + (1.0 - g) * MIX,
            b + (1.0 - b) * MIX,
        ]
    } else {
        [r, g, b]
    }
}

/// Grayscale latitude bands for event-root spheres. The material's base color
/// tints this, so roots read as a different surface from the flat node spheres.
pub fn event_texture_rgba(size: u32) -> Vec<u8> {
    let size = size.max(1);
    let mut rgba = vec![0u8; (size * size * 4) as usize];
    for y in 0..size {
        let v = y as f32 / size as f32;
        let pole = v < 0.12 || v > 0.88;
        let band = ((v * 8.0) as u32).is_multiple_of(2);
        let value = if pole {
            255
        } else if band {
            230
        } else {
            40
        };
        for x in 0..size {
            let i = ((y * size + x) * 4) as usize;
            rgba[i] = value;
            rgba[i + 1] = value;
            rgba[i + 2] = value;
            rgba[i + 3] = 255;
        }
    }
    rgba
}

/// Join stored rows into drawable nodes.
///
/// A node with no point or no event is skipped. Only `From` links are kept;
/// `Next`, `Blocked`, and `NotSet` are not connections to draw. Each `From`
/// link is stored on the later node and points at the earlier one.
pub fn scene_from_rows(
    points: &[PointRow],
    events: &[EventRow],
    nodes: &[NodeRow],
) -> Vec<SceneNode> {
    let points: HashMap<i64, [i32; 3]> = points.iter().map(|p| (p.id, [p.x, p.y, p.z])).collect();
    let events: HashMap<i32, &EventRow> = events.iter().map(|e| (e.id, e)).collect();
    let mut scene = Vec::with_capacity(nodes.len());
    for node in nodes {
        let Some(&position) = points.get(&node.point_id) else {
            continue;
        };
        let Some(event) = events.get(&node.event_id) else {
            continue;
        };
        let links = [node.node1, node.node2, node.node3];
        let mut from = Vec::new();
        for conn in 0..NB_CONNECTIONS {
            if get_connection_state(node.connection_mask as u16, conn as i32)
                != ConnectionState::From
            {
                continue;
            }
            if let Some(id) = links[conn] {
                from.push(id);
            }
        }
        scene.push(SceneNode {
            id: node.id,
            position,
            color: event.color as u8,
            creation_time: node.creation_time,
            origin: node.d == 0,
            d: node.d,
            event_id: node.event_id,
            path_ctx_id: event.path_ctx_id,
            from,
        });
    }
    scene
}

/// The selected node, then every node reached by following `From` links.
///
/// Order is breadth-first. Each id is recorded once. A missing id is still the
/// first entry, so a stale selection does not panic.
pub fn path_back_to_event(nodes: &[SceneNode], selected: i64) -> Vec<i64> {
    let by_id: HashMap<i64, &SceneNode> = nodes.iter().map(|n| (n.id, n)).collect();
    let mut ordered = Vec::new();
    let mut seen = HashSet::new();
    let mut queue = std::collections::VecDeque::new();
    queue.push_back(selected);
    while let Some(id) = queue.pop_front() {
        if !seen.insert(id) {
            continue;
        }
        ordered.push(id);
        if let Some(node) = by_id.get(&id) {
            for &earlier in &node.from {
                queue.push_back(earlier);
            }
        }
    }
    ordered
}

/// Markers whose color bit is enabled and whose time passes the filter.
///
/// Normally that is `creation_time <= filter.time`. With `latest_only`, it is
/// `creation_time == filter.time`, plus every node on the selected node's
/// from-chain back to its event root. A connection is drawn once, and only
/// when both ends are visible. Turning a color off hides every node of that
/// color, including event roots and a selected path of that color. Axes run
/// from `-(max_coord + 3)` to `+(max_coord + 3)`. Lattice lines sit on the
/// three planes through the origin, every [`GRID_STEP`] units, out to
/// `max_coord`.
pub fn build_frame(nodes: &[SceneNode], filter: &ViewFilter, max_coord: i32) -> Frame {
    let by_id: HashMap<i64, &SceneNode> = nodes.iter().map(|n| (n.id, n)).collect();
    let chain: HashSet<i64> = if filter.latest_only {
        filter
            .selected
            .map(|id| path_back_to_event(nodes, id).into_iter().collect())
            .unwrap_or_default()
    } else {
        HashSet::new()
    };
    let visible = |node: &SceneNode| {
        if node.color & filter.color_mask == 0 {
            return false;
        }
        let in_time = if filter.latest_only {
            node.creation_time == filter.time
        } else {
            node.creation_time <= filter.time
        };
        in_time || chain.contains(&node.id)
    };
    let mut markers = Vec::new();
    let mut segments = Vec::new();
    let mut seen = HashSet::new();
    for node in nodes {
        if !visible(node) {
            continue;
        }
        markers.push(Marker {
            node_id: node.id,
            position: pos(node.position),
            color: node.color,
            origin: node.origin,
            active: node.creation_time == filter.time,
            selected: filter.selected == Some(node.id),
        });
        if !filter.show_lines {
            continue;
        }
        for &earlier_id in &node.from {
            let Some(earlier) = by_id.get(&earlier_id) else {
                continue;
            };
            if !visible(earlier) {
                continue;
            }
            let key = if node.id < earlier_id {
                (node.id, earlier_id)
            } else {
                (earlier_id, node.id)
            };
            if !seen.insert(key) {
                continue;
            }
            segments.push(Segment {
                a: pos(earlier.position),
                b: pos(node.position),
                color: earlier.color,
            });
        }
    }
    let axes = if filter.show_axes {
        lattice(max_coord)
    } else {
        Vec::new()
    };
    Frame {
        markers,
        segments,
        axes,
    }
}

fn pos(p: [i32; 3]) -> [f32; 3] {
    [p[0] as f32, p[1] as f32, p[2] as f32]
}

fn lattice(max_coord: i32) -> Vec<Segment> {
    let len = (max_coord + AXE_EXTRA).max(1) as f32;
    let mut axes = vec![
        Segment {
            a: [-len, 0.0, 0.0],
            b: [len, 0.0, 0.0],
            color: 1,
        },
        Segment {
            a: [0.0, -len, 0.0],
            b: [0.0, len, 0.0],
            color: 2,
        },
        Segment {
            a: [0.0, 0.0, -len],
            b: [0.0, 0.0, len],
            color: 4,
        },
    ];
    let span = max_coord as f32;
    let n = max_coord / GRID_STEP;
    for i in -n..=n {
        if i == 0 {
            continue;
        }
        let c = (i * GRID_STEP) as f32;
        // XY, XZ, and YZ planes. The zero line of each plane is the axis.
        axes.push(Segment {
            a: [-span, c, 0.0],
            b: [span, c, 0.0],
            color: 0,
        });
        axes.push(Segment {
            a: [c, -span, 0.0],
            b: [c, span, 0.0],
            color: 0,
        });
        axes.push(Segment {
            a: [-span, 0.0, c],
            b: [span, 0.0, c],
            color: 0,
        });
        axes.push(Segment {
            a: [c, 0.0, -span],
            b: [c, 0.0, span],
            color: 0,
        });
        axes.push(Segment {
            a: [0.0, -span, c],
            b: [0.0, span, c],
            color: 0,
        });
        axes.push(Segment {
            a: [0.0, c, -span],
            b: [0.0, c, span],
            color: 0,
        });
    }
    axes
}

/// Camera aim point and distance that fit every loaded node.
pub fn camera_fit(nodes: &[SceneNode]) -> ([f32; 3], f32) {
    let Some(first) = nodes.first() else {
        return ([0.0, 0.0, 0.0], 40.0);
    };
    let mut min = first.position;
    let mut max = first.position;
    for node in nodes.iter().skip(1) {
        for axis in 0..3 {
            min[axis] = min[axis].min(node.position[axis]);
            max[axis] = max[axis].max(node.position[axis]);
        }
    }
    let center = [
        (min[0] + max[0]) as f32 * 0.5,
        (min[1] + max[1]) as f32 * 0.5,
        (min[2] + max[2]) as f32 * 0.5,
    ];
    let extent = (max[0] - min[0])
        .max(max[1] - min[1])
        .max(max[2] - min[2])
        .max(1) as f32;
    (center, extent * 1.6 + 8.0)
}

/// A simulation folder opened for display.
#[derive(Clone, Debug)]
pub struct LoadedSim {
    pub name: String,
    pub max_coord: i32,
    pub max_time: i32,
    pub nodes: Vec<SceneNode>,
    /// File name of the newest `run-*.json`, when one exists.
    pub run_name: Option<String>,
}

/// Read `inputs.json`, the latest run file, and the Iceberg tables under
/// `data_root/<name>`.
///
/// Does not create a folder. An unknown name, or a folder with no event nodes,
/// is an error: grow the simulation with `qsm <name>` first.
pub fn load_simulation(data_root: &Path, name: &str) -> Result<LoadedSim, String> {
    if !valid_name(name) {
        return Err(format!("invalid simulation name `{name}`"));
    }
    let root = data_root.join(name);
    if !root.is_dir() {
        return Err(format!(
            "no simulation folder {}; run `qsm {name}` first",
            root.display()
        ));
    }
    let inputs_path = root.join("inputs.json");
    let inputs = if inputs_path.is_file() {
        Some(load_inputs(&inputs_path).map_err(|err| err.to_string())?)
    } else {
        None
    };
    let run = latest_run(&root).map_err(|err| err.to_string())?;
    if inputs.is_none() && run.is_none() {
        return Err(format!(
            "{} has no inputs.json or run file; run `qsm {name}` first",
            root.display()
        ));
    }
    let max_coord = inputs
        .as_ref()
        .map(|doc| doc.space.max_coord)
        .or_else(|| run.as_ref().map(|(_, record)| record.space.max_coord))
        .unwrap_or(18);
    let run_name = run
        .as_ref()
        .and_then(|(path, _)| path.file_name().map(|n| n.to_string_lossy().into_owned()));

    let runtime = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .map_err(|err| err.to_string())?;
    let store = runtime
        .block_on(SimulationStore::open(data_root, name))
        .map_err(|err| err.to_string())?;
    drop(runtime);

    if store.nodes().is_empty() {
        return Err(format!("{name} has no event nodes; run `qsm {name}` first"));
    }
    let nodes = scene_from_rows(store.points(), store.events(), store.nodes());
    if nodes.is_empty() {
        return Err(format!(
            "{name} has event nodes that do not join to a point and an event"
        ));
    }
    let max_time = nodes.iter().map(|n| n.creation_time).max().unwrap_or(0);
    Ok(LoadedSim {
        name: name.to_string(),
        max_coord,
        max_time,
        nodes,
        run_name,
    })
}

/// Closest sphere hit by a ray, or nothing when the ray misses every marker.
///
/// `direction` does not need to be a unit vector. Event roots use
/// [`ORIGIN_RADIUS`]; other nodes use [`NODE_RADIUS`]. A selected marker uses
/// [`SELECTED_SCALE`].
pub fn pick_marker(markers: &[Marker], origin: [f32; 3], direction: [f32; 3]) -> Option<i64> {
    let dir = normalize(direction)?;
    let mut best: Option<(f32, i64)> = None;
    for marker in markers {
        let mut radius = if marker.origin {
            ORIGIN_RADIUS
        } else {
            NODE_RADIUS
        };
        if marker.selected {
            radius *= SELECTED_SCALE;
        }
        let Some(t) = ray_sphere(origin, dir, marker.position, radius) else {
            continue;
        };
        if best.is_none_or(|(best_t, _)| t < best_t) {
            best = Some((t, marker.node_id));
        }
    }
    best.map(|(_, id)| id)
}

fn normalize(v: [f32; 3]) -> Option<[f32; 3]> {
    let len = (v[0] * v[0] + v[1] * v[1] + v[2] * v[2]).sqrt();
    if len < 1.0e-8 {
        None
    } else {
        Some([v[0] / len, v[1] / len, v[2] / len])
    }
}

fn ray_sphere(origin: [f32; 3], dir: [f32; 3], center: [f32; 3], radius: f32) -> Option<f32> {
    let oc = [
        origin[0] - center[0],
        origin[1] - center[1],
        origin[2] - center[2],
    ];
    let b = oc[0] * dir[0] + oc[1] * dir[1] + oc[2] * dir[2];
    let c = oc[0] * oc[0] + oc[1] * oc[1] + oc[2] * oc[2] - radius * radius;
    let disc = b * b - c;
    if disc < 0.0 {
        return None;
    }
    let root = disc.sqrt();
    let near = -b - root;
    if near > 0.0 {
        Some(near)
    } else {
        let far = -b + root;
        (far > 0.0).then_some(far)
    }
}

fn valid_name(name: &str) -> bool {
    !name.is_empty() && !name.contains('/') && !name.contains('\\') && name != "." && name != ".."
}

#[cfg(test)]
mod tests;
