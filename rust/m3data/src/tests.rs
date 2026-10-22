use std::fs;
use std::path::PathBuf;
use std::time::{SystemTime, UNIX_EPOCH};

use std::collections::{HashMap, HashSet};

use m3path::{get_connection_state, ConnectionState, NB_CONNECTIONS};
use m3point::{
    Point, PointPackData, TrioIndex, EXPECTED_CONNECTIONS, EXPECTED_CUBES,
    EXPECTED_GROWTH_CONTEXTS, EXPECTED_TRIOS,
};
use m3space::{pyramid_centers, EventColor};

use crate::inputs::{default_inputs, load_inputs, save_inputs};
use crate::rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow};
use crate::runfile::{self, RunRecord};
use crate::schema::{bucket_of, BUCKET_WIDTH};
use crate::sim::run_named;
use crate::static_pack;
use crate::store::SimulationStore;

struct TempRoot(PathBuf);

impl TempRoot {
    fn new() -> Self {
        let nanos = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_nanos();
        let path = std::env::temp_dir().join(format!("qsm-m3data-{nanos}-{}", std::process::id()));
        fs::create_dir_all(&path).unwrap();
        Self(path)
    }
}

impl Drop for TempRoot {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.0);
    }
}

fn metadata_files(root: &std::path::Path, table: &str) -> Vec<PathBuf> {
    let dir = root.join("qsm").join(table).join("metadata");
    let mut files: Vec<_> = fs::read_dir(&dir)
        .unwrap()
        .map(|entry| entry.unwrap().path())
        .filter(|path| {
            path.file_name()
                .and_then(|name| name.to_str())
                .is_some_and(|name| name.ends_with(".metadata.json"))
        })
        .collect();
    files.sort();
    files
}

fn has_partition(root: &std::path::Path, table: &str, component: &str) -> bool {
    let data = root.join("qsm").join(table).join("data");
    walk_has(&data, component)
}

fn walk_has(dir: &std::path::Path, component: &str) -> bool {
    let Ok(entries) = fs::read_dir(dir) else {
        return false;
    };
    for entry in entries.flatten() {
        let path = entry.path();
        if path.file_name().and_then(|name| name.to_str()) == Some(component) {
            return true;
        }
        if path.is_dir() && walk_has(&path, component) {
            return true;
        }
    }
    false
}

#[test]
fn bucket_edges() {
    assert_eq!(BUCKET_WIDTH, 8);
    assert_eq!(bucket_of(0), 0);
    assert_eq!(bucket_of(7), 0);
    assert_eq!(bucket_of(8), 8);
    assert_eq!(bucket_of(-1), -8);
}

#[test]
fn static_reference_is_not_a_table() {
    let pack = static_pack();
    assert_eq!(pack.connections.len(), EXPECTED_CONNECTIONS);
    assert_eq!(pack.trios.len(), EXPECTED_TRIOS);
    assert_eq!(pack.growth_contexts.len(), EXPECTED_GROWTH_CONTEXTS);
    assert_eq!(pack.cube_ids.len(), EXPECTED_CUBES);
    assert_eq!(
        pack.path_builders
            .iter()
            .filter(|builder| builder.is_some())
            .count(),
        EXPECTED_CUBES
    );
}

#[tokio::test(flavor = "multi_thread")]
async fn round_trip_each_table() {
    let tmp = TempRoot::new();
    let mut store = SimulationStore::open(&tmp.0, "round").await.unwrap();
    let root = store.root().to_path_buf();

    let points = vec![
        PointRow {
            id: 1,
            x: 0,
            y: 0,
            z: 0,
        },
        PointRow {
            id: 2,
            x: 3,
            y: 0,
            z: 0,
        },
    ];
    assert_eq!(metadata_files(&root, "points").len(), 1);
    store.append_points(&points).await.unwrap();
    assert_eq!(
        metadata_files(&root, "points").len(),
        2,
        "two rows share one snapshot"
    );
    store
        .append_points(&[PointRow {
            id: 3,
            x: 6,
            y: 0,
            z: 0,
        }])
        .await
        .unwrap();
    assert_eq!(metadata_files(&root, "points").len(), 3);

    let path_contexts = vec![PathContextRow {
        id: 1,
        growth_ctx_id: 0,
        growth_offset: 0,
        path_builders_id: Some(4),
        max_dist: 0,
    }];
    store.append_path_contexts(&path_contexts).await.unwrap();

    let path_nodes = vec![
        PathNodeRow {
            id: 10,
            path_ctx_id: 1,
            path_builders_id: 4,
            path_builder_idx: 0,
            trio_id: 1,
            point_id: 1,
            d: 0,
            connection_mask: 0,
            path_node1: None,
            path_node2: None,
            path_node3: None,
        },
        PathNodeRow {
            id: 11,
            path_ctx_id: 1,
            path_builders_id: 4,
            path_builder_idx: 0,
            trio_id: 2,
            point_id: 2,
            d: 8,
            connection_mask: 3,
            path_node1: Some(10),
            path_node2: None,
            path_node3: None,
        },
    ];
    let before_nodes = metadata_files(&root, "path_nodes").len();
    store.append_path_nodes(&path_nodes).await.unwrap();
    assert_eq!(metadata_files(&root, "path_nodes").len(), before_nodes + 1);
    assert!(has_partition(&root, "path_nodes", "path_ctx_id=1"));
    assert!(has_partition(&root, "path_nodes", "d_bucket=0"));
    assert!(has_partition(&root, "path_nodes", "d_bucket=8"));

    let events = vec![
        EventRow {
            id: 1,
            path_ctx_id: 1,
            creation_time: 0,
            color: 1,
            end_time: 0,
            max_node_time: 0,
        },
        EventRow {
            id: 2,
            path_ctx_id: 1,
            creation_time: 4,
            color: 2,
            end_time: 9,
            max_node_time: 4,
        },
    ];
    store.append_events(&events).await.unwrap();
    assert!(!has_partition(&root, "events", "space_id=1"));

    let nodes = vec![
        NodeRow {
            id: 20,
            event_id: 1,
            path_node_id: 10,
            trio_id: 1,
            point_id: 1,
            d: 0,
            creation_time: 0,
            connection_mask: 0,
            node1: None,
            node2: None,
            node3: None,
        },
        NodeRow {
            id: 21,
            event_id: 1,
            path_node_id: 11,
            trio_id: 2,
            point_id: 2,
            d: 8,
            creation_time: 8,
            connection_mask: 1,
            node1: Some(20),
            node2: None,
            node3: None,
        },
    ];
    store.append_nodes(&nodes).await.unwrap();
    assert!(has_partition(&root, "nodes", "event_id=1"));
    assert!(has_partition(&root, "nodes", "creation_time_bucket=0"));
    assert!(has_partition(&root, "nodes", "creation_time_bucket=8"));

    drop(store);
    let loaded = SimulationStore::open(&tmp.0, "round").await.unwrap();
    let mut loaded_points = loaded.points().to_vec();
    loaded_points.sort_by_key(|row| row.id);
    let mut expected_points = points;
    expected_points.push(PointRow {
        id: 3,
        x: 6,
        y: 0,
        z: 0,
    });
    assert_eq!(loaded_points, expected_points);
    assert_eq!(loaded.path_contexts(), path_contexts.as_slice());
    let mut loaded_path_nodes = loaded.path_nodes().to_vec();
    loaded_path_nodes.sort_by_key(|row| row.id);
    assert_eq!(loaded_path_nodes, path_nodes);
    let mut loaded_events = loaded.events().to_vec();
    loaded_events.sort_by_key(|row| row.id);
    assert_eq!(loaded_events, events);
    let mut loaded_nodes = loaded.nodes().to_vec();
    loaded_nodes.sort_by_key(|row| row.id);
    assert_eq!(loaded_nodes, nodes);

    let names = table_names(&loaded.root().to_path_buf());
    assert_eq!(
        names,
        vec!["events", "nodes", "path_contexts", "path_nodes", "points"]
    );
}

fn table_names(root: &std::path::Path) -> Vec<String> {
    let mut names: Vec<String> = fs::read_dir(root.join("qsm"))
        .unwrap()
        .map(|entry| entry.unwrap().file_name().into_string().unwrap())
        .collect();
    names.sort();
    names
}

#[tokio::test(flavor = "multi_thread")]
async fn uniqueness_is_enforced_in_rust() {
    let tmp = TempRoot::new();
    let mut store = SimulationStore::open(&tmp.0, "uniq").await.unwrap();
    store
        .append_points(&[PointRow {
            id: 1,
            x: 1,
            y: 2,
            z: 3,
        }])
        .await
        .unwrap();
    let dup_xyz = store
        .append_points(&[PointRow {
            id: 2,
            x: 1,
            y: 2,
            z: 3,
        }])
        .await
        .unwrap_err();
    assert!(dup_xyz.to_string().contains("duplicate point"));
    let dup_id = store
        .append_points(&[PointRow {
            id: 1,
            x: 4,
            y: 0,
            z: 0,
        }])
        .await
        .unwrap_err();
    assert!(dup_id.to_string().contains("duplicate points id"));

    store
        .append_path_nodes(&[PathNodeRow {
            id: 5,
            path_ctx_id: 7,
            path_builders_id: 1,
            path_builder_idx: 0,
            trio_id: 1,
            point_id: 1,
            d: 0,
            connection_mask: 0,
            path_node1: None,
            path_node2: None,
            path_node3: None,
        }])
        .await
        .unwrap();
    let dup_path = store
        .append_path_nodes(&[PathNodeRow {
            id: 6,
            path_ctx_id: 7,
            path_builders_id: 1,
            path_builder_idx: 0,
            trio_id: 1,
            point_id: 1,
            d: 1,
            connection_mask: 0,
            path_node1: None,
            path_node2: None,
            path_node3: None,
        }])
        .await
        .unwrap_err();
    assert!(dup_path.to_string().contains("duplicate path node (7, 1)"));

    store
        .append_nodes(&[NodeRow {
            id: 8,
            event_id: 3,
            path_node_id: 5,
            trio_id: 1,
            point_id: 1,
            d: 0,
            creation_time: 0,
            connection_mask: 0,
            node1: None,
            node2: None,
            node3: None,
        }])
        .await
        .unwrap();
    let dup_node = store
        .append_nodes(&[NodeRow {
            id: 9,
            event_id: 3,
            path_node_id: 6,
            trio_id: 1,
            point_id: 1,
            d: 1,
            creation_time: 1,
            connection_mask: 0,
            node1: None,
            node2: None,
            node3: None,
        }])
        .await
        .unwrap_err();
    assert!(dup_node.to_string().contains("duplicate node (3, 1)"));

    drop(store);
    let loaded = SimulationStore::open(&tmp.0, "uniq").await.unwrap();
    assert_eq!(loaded.points().len(), 1);
    assert_eq!(loaded.path_nodes().len(), 1);
    assert_eq!(loaded.nodes().len(), 1);
}

#[tokio::test(flavor = "multi_thread")]
async fn path_node_update_reloads_the_closed_connection() {
    let tmp = TempRoot::new();
    let mut store = SimulationStore::open(&tmp.0, "mask").await.unwrap();
    let original = PathNodeRow {
        id: 5,
        path_ctx_id: 7,
        path_builders_id: 1,
        path_builder_idx: 0,
        trio_id: 1,
        point_id: 1,
        d: 0,
        connection_mask: 0,
        path_node1: None,
        path_node2: None,
        path_node3: None,
    };
    store.append_path_nodes(&[original.clone()]).await.unwrap();
    let updated = PathNodeRow {
        connection_mask: ConnectionState::From as i32,
        path_node1: Some(9),
        ..original
    };
    store.append_path_nodes(&[updated.clone()]).await.unwrap();
    assert_eq!(store.path_nodes().len(), 1);
    assert_eq!(
        store.path_nodes()[0].connection_mask,
        updated.connection_mask
    );
    drop(store);
    let loaded = SimulationStore::open(&tmp.0, "mask").await.unwrap();
    assert_eq!(loaded.path_nodes(), &[updated]);
}

/// playgl pyramid, three growth steps after the time-0 roots.
///
/// Per path context the shell sizes are 1, 3, 6, 12. That is
/// `TestMakeNewPathCtx` at distances 0 and 1 (type 8, index 0, offset 0) and the
/// same 6 / 12 counts `server_path_test.go` records for type 8 at distances 2 and 3.
/// Four events, so each time step stores 4, 12, 24, 48 event nodes.
#[test]
fn pyramid_three_steps_matches_a_resumed_run() {
    let straight = run_pyramid(3);
    assert!(!straight.report.resumed);
    assert_eq!(straight.report.max_time_steps, 3);
    assert_eq!(straight.report.time, 3);
    assert_eq!(straight.report.events, 4);
    assert_eq!(straight.report.nodes, 88);
    assert_eq!(straight.report.steps_computed, 3);
    assert_pyramid(&straight.store);

    let reloaded = open_store(&straight.root, "pyramid");
    assert_eq!(snapshot(&reloaded), snapshot(&straight.store));

    assert_eq!(sorted_runs(straight.store.root()).len(), 1);
    let again = run_named(&straight.root, "pyramid").unwrap();
    assert!(again.resumed);
    assert_eq!(again.steps_computed, 0);
    assert_eq!(again.time, 3);
    assert_eq!(again.nodes, 88);
    assert_eq!(sorted_runs(straight.store.root()).len(), 2);

    for stop in [1, 2] {
        let partial = TempRoot::new();
        write_pyramid(&partial.0, "pyramid", stop);
        let mid = run_named(&partial.0, "pyramid").unwrap();
        assert!(!mid.resumed);
        assert_eq!(mid.time, stop);
        assert_eq!(mid.steps_computed, stop);
        let per_step = [4, 12, 24, 48];
        let expected_nodes: usize = per_step.iter().take(stop as usize + 1).sum();
        assert_eq!(mid.nodes, expected_nodes);

        write_pyramid(&partial.0, "pyramid", 3);
        let finished = run_named(&partial.0, "pyramid").unwrap();
        assert!(finished.resumed);
        assert_eq!(finished.time, 3);
        assert_eq!(finished.steps_computed, 3 - stop);
        assert_eq!(finished.nodes, 88);
        let loaded = open_store(&partial.0, "pyramid");
        assert_eq!(snapshot(&loaded), snapshot(&straight.store));
    }
}

struct PyramidRun {
    root: std::path::PathBuf,
    report: crate::sim::StepReport,
    store: SimulationStore,
}

impl Drop for PyramidRun {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.root);
    }
}

fn run_pyramid(max_time_steps: i32) -> PyramidRun {
    let tmp = TempRoot::new();
    write_pyramid(&tmp.0, "pyramid", max_time_steps);
    let report = run_named(&tmp.0, "pyramid").unwrap();
    let store = open_store(&tmp.0, "pyramid");
    let root = tmp.0.clone();
    std::mem::forget(tmp);
    PyramidRun {
        root,
        report,
        store,
    }
}

fn write_pyramid(root: &std::path::Path, name: &str, max_time_steps: i32) {
    let mut inputs = default_inputs();
    inputs.max_time_steps = max_time_steps;
    save_inputs(&root.join(name).join("inputs.json"), &inputs).unwrap();
}

fn open_store(root: &std::path::Path, name: &str) -> SimulationStore {
    let runtime = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .unwrap();
    runtime.block_on(SimulationStore::open(root, name)).unwrap()
}

fn assert_pyramid(store: &SimulationStore) {
    let inputs = load_inputs(&store.root().join("inputs.json")).unwrap();
    assert_eq!(inputs.max_time_steps, 3);
    let runs = sorted_runs(store.root());
    assert_eq!(runs.len(), 1);
    let record = load_run(&runs[0]);
    assert_eq!(record.max_time_steps, 3);
    assert_eq!(record.creation_time, 3);
    assert_eq!(record.space.active_path_node_threshold, 3);
    assert_eq!(record.space.max_trios_per_point, 2);
    assert_eq!(record.space.max_path_nodes_per_point, 4);
    assert_eq!(record.space.max_coord, 18);
    assert_eq!(record.events.len(), 4);
    let raw = fs::read(store.root().join("inputs.json")).unwrap();
    assert_eq!(record.inputs_sha256, runfile::sha256_hex(&raw));

    let centers = pyramid_centers(4);
    let colors = [
        EventColor::Red as i32,
        EventColor::Green as i32,
        EventColor::Blue as i32,
        EventColor::Yellow as i32,
    ];
    let offsets = [0, 2, 4, 6];
    assert_eq!(store.events().len(), 4);
    assert_eq!(store.path_contexts().len(), 4);
    let points = point_map(store);
    let mut events = store.events().to_vec();
    events.sort_by_key(|event| event.id);
    for (index, event) in events.iter().enumerate() {
        assert_eq!(event.color, colors[index]);
        assert_eq!(event.creation_time, 0);
        assert_eq!(event.max_node_time, 3);
        let ctx = store
            .path_contexts()
            .iter()
            .find(|ctx| ctx.id == event.path_ctx_id)
            .unwrap();
        assert_eq!(ctx.growth_ctx_id, 40);
        assert_eq!(ctx.growth_offset, offsets[index]);
        assert_eq!(ctx.max_dist, 3);
        let root = store
            .nodes()
            .iter()
            .find(|node| node.event_id == event.id && node.d == 0)
            .unwrap();
        assert_eq!(points[&root.point_id], centers[index]);
    }

    let shells = [1, 3, 6, 12];
    for ctx in store.path_contexts() {
        for (distance, expect) in shells.iter().enumerate() {
            let count = store
                .path_nodes()
                .iter()
                .filter(|node| node.path_ctx_id == ctx.id && node.d == distance as i32)
                .count();
            assert_eq!(count, *expect, "context {} distance {distance}", ctx.id);
        }
    }
    assert_eq!(store.path_nodes().len(), 88);
    let created = [4, 12, 24, 48];
    for (time, expect) in created.iter().enumerate() {
        let count = store
            .nodes()
            .iter()
            .filter(|node| node.creation_time == time as i32)
            .count();
        assert_eq!(count, *expect, "creation time {time}");
    }
    assert_eq!(store.nodes().len(), 88);
    assert_unique(store);
    assert_links(store);
}

fn point_map(store: &SimulationStore) -> HashMap<i64, Point> {
    store
        .points()
        .iter()
        .map(|row| (row.id, row.point()))
        .collect()
}

fn assert_unique(store: &SimulationStore) {
    let mut ids = HashSet::new();
    let mut xyz = HashSet::new();
    for row in store.points() {
        assert!(ids.insert(row.id));
        assert!(xyz.insert(row.xyz()), "duplicate point {:?}", row.xyz());
    }
    let mut ids = HashSet::new();
    let mut keys = HashSet::new();
    for row in store.path_contexts() {
        assert!(ids.insert(row.id));
        assert!(keys.insert((row.growth_ctx_id, row.growth_offset)));
    }
    let mut ids = HashSet::new();
    let mut keys = HashSet::new();
    for row in store.path_nodes() {
        assert!(ids.insert(row.id), "duplicate path node id {}", row.id);
        assert!(
            keys.insert((row.path_ctx_id, row.point_id)),
            "duplicate path node ({}, {})",
            row.path_ctx_id,
            row.point_id
        );
    }
    let mut ids = HashSet::new();
    let mut keys = HashSet::new();
    for row in store.nodes() {
        assert!(ids.insert(row.id), "duplicate node id {}", row.id);
        assert!(
            keys.insert((row.event_id, row.point_id)),
            "duplicate node ({}, {})",
            row.event_id,
            row.point_id
        );
    }
}

fn assert_links(store: &SimulationStore) {
    let pack = PointPackData::shared();
    let points = point_map(store);
    let path_by_id: HashMap<i64, &PathNodeRow> = store
        .path_nodes()
        .iter()
        .map(|node| (node.id, node))
        .collect();
    for node in store.path_nodes() {
        let trio = *pack.trio(TrioIndex(node.trio_id as u8));
        let here = points[&node.point_id];
        let links = [node.path_node1, node.path_node2, node.path_node3];
        for conn in 0..NB_CONNECTIONS {
            let state = get_connection_state(node.connection_mask as u16, conn as i32);
            match state {
                ConnectionState::NotSet | ConnectionState::Blocked => {
                    assert!(links[conn].is_none(), "path {} conn {conn}", node.id);
                }
                ConnectionState::Next => {
                    let next = path_by_id[&links[conn].unwrap()];
                    assert_eq!(next.path_ctx_id, node.path_ctx_id);
                    assert_eq!(next.d, node.d + 1);
                    assert_eq!(points[&next.point_id], here.add(trio.conns[conn].vector));
                }
                ConnectionState::From => {
                    let prev = path_by_id[&links[conn].unwrap()];
                    assert_eq!(prev.path_ctx_id, node.path_ctx_id);
                    assert_eq!(prev.d + 1, node.d);
                    // The from-slot vector points back at the parent.
                    assert_eq!(points[&prev.point_id], here.add(trio.conns[conn].vector));
                }
            }
        }
    }
    let node_by_id: HashMap<i64, &NodeRow> =
        store.nodes().iter().map(|node| (node.id, node)).collect();
    for node in store.nodes() {
        let trio = *pack.trio(TrioIndex(node.trio_id as u8));
        let here = points[&node.point_id];
        let links = [node.node1, node.node2, node.node3];
        let path = path_by_id[&node.path_node_id];
        let path_links = [path.path_node1, path.path_node2, path.path_node3];
        for conn in 0..NB_CONNECTIONS {
            let state = get_connection_state(node.connection_mask as u16, conn as i32);
            match state {
                ConnectionState::NotSet | ConnectionState::Blocked | ConnectionState::Next => {
                    assert!(links[conn].is_none(), "node {} conn {conn}", node.id);
                }
                ConnectionState::From => {
                    let prev = node_by_id[&links[conn].unwrap()];
                    assert_eq!(prev.event_id, node.event_id);
                    assert!(prev.creation_time < node.creation_time);
                    assert_eq!(points[&prev.point_id], here.add(trio.conns[conn].vector));
                    assert_eq!(path_links[conn], Some(prev.path_node_id));
                }
            }
        }
    }
}

#[derive(PartialEq, Eq, Debug)]
struct Snapshot {
    points: Vec<PointRow>,
    path_contexts: Vec<PathContextRow>,
    path_nodes: Vec<PathNodeRow>,
    events: Vec<EventRow>,
    nodes: Vec<NodeRow>,
}

fn snapshot(store: &SimulationStore) -> Snapshot {
    let mut points = store.points().to_vec();
    points.sort_by_key(|row| row.id);
    let mut path_contexts = store.path_contexts().to_vec();
    path_contexts.sort_by_key(|row| row.id);
    let mut path_nodes = store.path_nodes().to_vec();
    path_nodes.sort_by_key(|row| row.id);
    let mut events = store.events().to_vec();
    events.sort_by_key(|row| row.id);
    let mut nodes = store.nodes().to_vec();
    nodes.sort_by_key(|row| row.id);
    Snapshot {
        points,
        path_contexts,
        path_nodes,
        events,
        nodes,
    }
}

fn sorted_runs(root: &std::path::Path) -> Vec<PathBuf> {
    let mut runs = runfile::run_files(root).unwrap();
    runs.sort();
    runs
}

fn load_run(path: &std::path::Path) -> RunRecord {
    serde_json::from_str(&fs::read_to_string(path).unwrap()).unwrap()
}

/// A finished run records its space. Raising only `max_time_steps` resumes.
/// Any other input change, or rows without a run file, is a hard error and writes nothing.
#[test]
fn run_file_locks_space_and_events() {
    let tmp = TempRoot::new();
    write_pyramid(&tmp.0, "pyramid", 0);
    let first = run_named(&tmp.0, "pyramid").unwrap();
    assert!(!first.resumed);
    assert_eq!(first.time, 0);
    assert_eq!(first.nodes, 4);
    assert_eq!(first.steps_computed, 0);
    let runs = sorted_runs(&tmp.0.join("pyramid"));
    assert_eq!(runs.len(), 1);
    let name = runs[0].file_name().unwrap().to_str().unwrap();
    assert!(
        name.starts_with("run-") && name.ends_with("Z.json"),
        "{name}"
    );
    let record = load_run(&runs[0]);
    assert_eq!(record.creation_time, 0);
    assert_eq!(record.max_time_steps, 0);
    assert_eq!(record.events.len(), 4);
    assert_eq!(record.events[0].growth_offset, 0);
    assert_eq!(record.events[0].x, 12);

    write_pyramid(&tmp.0, "pyramid", 1);
    let second = run_named(&tmp.0, "pyramid").unwrap();
    assert!(second.resumed);
    assert_eq!(second.time, 1);
    assert_eq!(second.steps_computed, 1);
    assert_eq!(second.nodes, 4 + 12);
    assert_eq!(sorted_runs(&tmp.0.join("pyramid")).len(), 2);
    let saved_nodes = open_store(&tmp.0, "pyramid").nodes().len();

    let space_err = rewrite_and_fail(&tmp.0, |inputs| inputs.space.max_coord = 21);
    assert!(
        space_err.contains("space.max_coord: last run 18, inputs 21"),
        "{space_err}"
    );
    assert!(
        space_err.contains("only max_time_steps may change"),
        "{space_err}"
    );
    assert_eq!(sorted_runs(&tmp.0.join("pyramid")).len(), 2);
    assert_eq!(open_store(&tmp.0, "pyramid").nodes().len(), saved_nodes);

    write_pyramid(&tmp.0, "pyramid", 1);
    let pyramid_err = rewrite_and_fail(&tmp.0, |inputs| match &mut inputs.shortcuts[0] {
        crate::inputs::Shortcut::Pyramid(spec) => spec.size = 5,
    });
    assert!(
        pyramid_err.contains("events[0].x: last run 12, inputs 15"),
        "{pyramid_err}"
    );
    assert_eq!(sorted_runs(&tmp.0.join("pyramid")).len(), 2);

    write_pyramid(&tmp.0, "pyramid", 1);
    let added = rewrite_and_fail(&tmp.0, |inputs| {
        inputs.events.push(crate::inputs::EventInput {
            x: 0,
            y: 0,
            z: 0,
            creation_time: 0,
            color: 1,
            growth_type: 8,
            growth_index: 0,
            growth_offset: 1,
        });
    });
    assert!(
        added.contains("events: last run has 4 events, inputs expand to 5"),
        "{added}"
    );

    write_pyramid(&tmp.0, "pyramid", 1);
    let removed = rewrite_and_fail(&tmp.0, |inputs| inputs.shortcuts.clear());
    assert!(
        removed.contains("events: last run has 4 events, inputs expand to 0"),
        "{removed}"
    );

    write_pyramid(&tmp.0, "pyramid", 1);
    let offset = rewrite_and_fail(&tmp.0, |inputs| match &mut inputs.shortcuts[0] {
        crate::inputs::Shortcut::Pyramid(spec) => spec.offsets[0] = 1,
    });
    assert!(
        offset.contains("events[0].growth_offset: last run 0, inputs 1"),
        "{offset}"
    );
    assert_eq!(open_store(&tmp.0, "pyramid").nodes().len(), saved_nodes);
    assert_eq!(sorted_runs(&tmp.0.join("pyramid")).len(), 2);

    for path in sorted_runs(&tmp.0.join("pyramid")) {
        fs::remove_file(path).unwrap();
    }
    let missing = run_named(&tmp.0, "pyramid").unwrap_err();
    let missing = missing.to_string();
    assert!(
        missing.contains("has saved rows but no qsm/run-*.json"),
        "{missing}"
    );
    assert!(sorted_runs(&tmp.0.join("pyramid")).is_empty());
    assert_eq!(open_store(&tmp.0, "pyramid").nodes().len(), saved_nodes);
}

fn rewrite_and_fail(
    root: &std::path::Path,
    edit: impl FnOnce(&mut crate::inputs::SimulationInputs),
) -> String {
    let path = root.join("pyramid").join("inputs.json");
    let mut inputs = load_inputs(&path).unwrap();
    edit(&mut inputs);
    save_inputs(&path, &inputs).unwrap();
    run_named(root, "pyramid").unwrap_err().to_string()
}
