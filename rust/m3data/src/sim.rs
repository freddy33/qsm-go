//! One step of a named simulation, and the synchronous entry used by `qsm`.
//!
//! The full Go growth loop is not ported. Each invocation either seeds a run
//! or advances one main point, so a later invocation of the same name can
//! resume from the files.

use std::fmt;
use std::path::{Path, PathBuf};

use m3point::{GrowthType, NodeHandle, Point, PointPackData, ORIGIN, X_FIRST};
use m3space::{EventColor, MIN_MAX_COORD};

use crate::error::Result;
use crate::rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow, SpaceRow};
use crate::store::SimulationStore;

/// What the last saved step was after this invocation.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct StepReport {
    pub simulation: String,
    pub folder: PathBuf,
    pub resumed: bool,
    pub distance: i32,
    pub creation_time: i32,
    pub point: Point,
}

impl fmt::Display for StepReport {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        writeln!(f, "simulation: {}", self.simulation)?;
        writeln!(f, "folder: {}", self.folder.display())?;
        writeln!(f, "resumed: {}", if self.resumed { "yes" } else { "no" })?;
        writeln!(f, "distance: {}", self.distance)?;
        writeln!(f, "creation_time: {}", self.creation_time)?;
        write!(
            f,
            "point: {},{},{}",
            self.point.0[0], self.point.0[1], self.point.0[2]
        )
    }
}

/// Data root for the CLI: `QSM_DATA_ROOT` when set, otherwise `data`.
pub fn default_data_root() -> PathBuf {
    match std::env::var("QSM_DATA_ROOT") {
        Ok(value) => PathBuf::from(value),
        Err(_) => PathBuf::from("data"),
    }
}

/// Open `data_root/<simulation>` and seed it, or advance one step.
pub fn run_named(data_root: impl AsRef<Path>, simulation: &str) -> Result<StepReport> {
    let runtime = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()?;
    runtime.block_on(run_named_async(data_root.as_ref(), simulation))
}

pub async fn run_named_async(data_root: &Path, simulation: &str) -> Result<StepReport> {
    // Static reference data is built here and never written.
    let _pack = PointPackData::shared();
    let mut store = SimulationStore::open(data_root, simulation).await?;
    let report = if store.spaces().is_empty() {
        seed(&mut store).await?
    } else {
        advance(&mut store).await?
    };
    Ok(report)
}

async fn seed(store: &mut SimulationStore) -> Result<StepReport> {
    let pack = PointPackData::shared();
    let growth = pack.growth_by_type_and_index(GrowthType(1), 0);
    let (path_builders_id, trio_id) = root_at(ORIGIN, growth.id, 0)?;
    let point = PointRow {
        id: 1,
        x: 0,
        y: 0,
        z: 0,
    };
    let path_ctx = PathContextRow {
        id: 1,
        growth_ctx_id: growth.id,
        growth_offset: 0,
        path_builders_id: Some(path_builders_id),
        max_dist: 0,
    };
    let path_node = PathNodeRow {
        id: 1,
        path_ctx_id: path_ctx.id,
        path_builders_id,
        path_builder_idx: 0,
        trio_id,
        point_id: point.id,
        d: 0,
        connection_mask: 0,
        path_node1: None,
        path_node2: None,
        path_node3: None,
    };
    let space = SpaceRow {
        id: 1,
        name: store.name().to_string(),
        active_path_node_threshold: 0,
        max_trios_per_point: 1,
        max_path_nodes_per_point: 1,
        max_coord: MIN_MAX_COORD,
        max_node_time: 0,
    };
    let event = EventRow {
        id: 1,
        space_id: space.id,
        path_ctx_id: path_ctx.id,
        creation_time: 0,
        color: i32::from(EventColor::Red as u8),
        end_time: 0,
        max_node_time: 0,
    };
    let node = NodeRow {
        id: 1,
        event_id: event.id,
        path_node_id: path_node.id,
        trio_id,
        point_id: point.id,
        d: 0,
        creation_time: 0,
        connection_mask: 0,
        node1: None,
        node2: None,
        node3: None,
    };
    store.append_points(&[point]).await?;
    store.append_path_contexts(&[path_ctx]).await?;
    store.append_path_nodes(&[path_node]).await?;
    store.append_spaces(&[space]).await?;
    store.append_events(&[event]).await?;
    store.append_nodes(&[node]).await?;
    Ok(StepReport {
        simulation: store.name().to_string(),
        folder: store.root().to_path_buf(),
        resumed: false,
        distance: 0,
        creation_time: 0,
        point: ORIGIN,
    })
}

async fn advance(store: &mut SimulationStore) -> Result<StepReport> {
    let cursor = store
        .path_nodes()
        .iter()
        .max_by_key(|node| (node.d, node.id))
        .cloned()
        .ok_or_else(|| {
            crate::error::Error::new("simulation has a space but no path node to resume from")
        })?;
    let last_point = store
        .points()
        .iter()
        .find(|point| point.id == cursor.point_id)
        .cloned()
        .ok_or_else(|| {
            crate::error::Error::new(format!(
                "path node {} points at missing point {}",
                cursor.id, cursor.point_id
            ))
        })?;
    let event = store
        .events()
        .iter()
        .max_by_key(|event| event.id)
        .cloned()
        .ok_or_else(|| crate::error::Error::new("simulation has no event to resume from"))?;
    let path_ctx = store
        .path_contexts()
        .iter()
        .find(|ctx| ctx.id == event.path_ctx_id)
        .cloned()
        .ok_or_else(|| {
            crate::error::Error::new(format!("missing path context {}", event.path_ctx_id))
        })?;
    let next_point_xyz = last_point.point().add(X_FIRST);
    let (path_builders_id, trio_id) = root_at(
        next_point_xyz,
        path_ctx.growth_ctx_id,
        path_ctx.growth_offset,
    )?;
    let creation_time = store
        .nodes()
        .iter()
        .map(|node| node.creation_time)
        .max()
        .unwrap_or(0)
        + 1;
    let distance = cursor.d + 1;
    let point = PointRow {
        id: next_i64(store.points().iter().map(|row| row.id)),
        x: next_point_xyz.0[0],
        y: next_point_xyz.0[1],
        z: next_point_xyz.0[2],
    };
    let path_node = PathNodeRow {
        id: next_i64(store.path_nodes().iter().map(|row| row.id)),
        path_ctx_id: path_ctx.id,
        path_builders_id,
        path_builder_idx: 0,
        trio_id,
        point_id: point.id,
        d: distance,
        connection_mask: 0,
        path_node1: None,
        path_node2: None,
        path_node3: None,
    };
    let node = NodeRow {
        id: next_i64(store.nodes().iter().map(|row| row.id)),
        event_id: event.id,
        path_node_id: path_node.id,
        trio_id,
        point_id: point.id,
        d: distance,
        creation_time,
        connection_mask: 0,
        node1: None,
        node2: None,
        node3: None,
    };
    let point_id_xyz = point.point();
    store.append_points(&[point]).await?;
    store.append_path_nodes(&[path_node]).await?;
    store.append_nodes(&[node]).await?;
    Ok(StepReport {
        simulation: store.name().to_string(),
        folder: store.root().to_path_buf(),
        resumed: true,
        distance,
        creation_time,
        point: point_id_xyz,
    })
}

fn next_i64(ids: impl Iterator<Item = i64>) -> i64 {
    ids.max().unwrap_or(0) + 1
}

/// Root path-builder id (the trio-cube id) and trio index at a main point.
fn root_at(point: Point, growth_ctx_id: i32, offset: i32) -> Result<(i32, i32)> {
    let pack = PointPackData::shared();
    let ctx = pack.growth_by_id(growth_ctx_id);
    let handle = pack.path_node_builder(ctx, offset, point);
    let trio_id = i32::from(handle.trio_index(pack).0);
    match handle {
        NodeHandle::Root(cube_id) => Ok((cube_id, trio_id)),
        _ => Err(crate::error::Error::new(format!(
            "expected a root path builder at {point} in growth context {growth_ctx_id}"
        ))),
    }
}
