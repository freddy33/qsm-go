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
/// Go `SphereRadius` default (`SizeVar{0.1, 0.8, 0.4}`).
pub const NODE_RADIUS: f32 = 0.4;
/// Event roots (`d == 0`) use a larger sphere so the origin reads as a root.
pub const ORIGIN_RADIUS: f32 = 0.7;
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
    let colors: HashMap<i32, u8> = events.iter().map(|e| (e.id, e.color as u8)).collect();
    let mut scene = Vec::with_capacity(nodes.len());
    for node in nodes {
        let Some(&position) = points.get(&node.point_id) else {
            continue;
        };
        let Some(&color) = colors.get(&node.event_id) else {
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
            color,
            creation_time: node.creation_time,
            origin: node.d == 0,
            from,
        });
    }
    scene
}

/// Markers with `creation_time <= filter.time` whose color bit is enabled.
///
/// A connection is drawn once, and only when both ends are visible. Turning a
/// color off hides every node of that color, including event roots. Axes run
/// from `-(max_coord + 3)` to `+(max_coord + 3)`. Lattice lines sit on the
/// three planes through the origin, every [`GRID_STEP`] units, out to
/// `max_coord`.
pub fn build_frame(nodes: &[SceneNode], filter: &ViewFilter, max_coord: i32) -> Frame {
    let by_id: HashMap<i64, &SceneNode> = nodes.iter().map(|n| (n.id, n)).collect();
    let visible =
        |node: &SceneNode| node.creation_time <= filter.time && node.color & filter.color_mask != 0;
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

fn valid_name(name: &str) -> bool {
    !name.is_empty() && !name.contains('/') && !name.contains('\\') && name != "." && name != ".."
}

#[cfg(test)]
mod tests;
