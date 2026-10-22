use std::fs;
use std::path::PathBuf;
use std::time::{SystemTime, UNIX_EPOCH};

use m3point::Point;
use m3point::{EXPECTED_CONNECTIONS, EXPECTED_CUBES, EXPECTED_GROWTH_CONTEXTS, EXPECTED_TRIOS};

use crate::rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow, SpaceRow};
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

    let spaces = vec![SpaceRow {
        id: 1,
        name: "round".to_string(),
        active_path_node_threshold: 0,
        max_trios_per_point: 1,
        max_path_nodes_per_point: 1,
        max_coord: 18,
        max_node_time: 0,
    }];
    store.append_spaces(&spaces).await.unwrap();

    let events = vec![
        EventRow {
            id: 1,
            space_id: 1,
            path_ctx_id: 1,
            creation_time: 0,
            color: 1,
            end_time: 0,
            max_node_time: 0,
        },
        EventRow {
            id: 2,
            space_id: 2,
            path_ctx_id: 1,
            creation_time: 4,
            color: 2,
            end_time: 9,
            max_node_time: 4,
        },
    ];
    store.append_events(&events).await.unwrap();
    assert!(has_partition(&root, "events", "space_id=1"));
    assert!(has_partition(&root, "events", "space_id=2"));

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
    assert_eq!(loaded.spaces(), spaces.as_slice());
    let mut loaded_events = loaded.events().to_vec();
    loaded_events.sort_by_key(|row| row.id);
    assert_eq!(loaded_events, events);
    let mut loaded_nodes = loaded.nodes().to_vec();
    loaded_nodes.sort_by_key(|row| row.id);
    assert_eq!(loaded_nodes, nodes);

    let names = table_names(&loaded.root().to_path_buf());
    assert_eq!(
        names,
        vec![
            "events",
            "nodes",
            "path_contexts",
            "path_nodes",
            "points",
            "spaces"
        ]
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

#[test]
fn resume_from_previous_run() {
    let tmp = TempRoot::new();
    let first = run_named(&tmp.0, "resume-sim").unwrap();
    assert!(!first.resumed);
    assert_eq!(first.distance, 0);
    assert_eq!(first.creation_time, 0);
    assert_eq!(first.point, Point([0, 0, 0]));

    let second = run_named(&tmp.0, "resume-sim").unwrap();
    assert!(second.resumed);
    assert_eq!(second.distance, 1);
    assert_eq!(second.creation_time, 1);
    assert_eq!(second.point, Point([3, 0, 0]));

    let third = run_named(&tmp.0, "resume-sim").unwrap();
    assert!(third.resumed);
    assert_eq!(third.distance, 2);
    assert_eq!(third.point, Point([6, 0, 0]));
    assert_eq!(third.folder, first.folder);

    let runtime = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .unwrap();
    let loaded = runtime
        .block_on(SimulationStore::open(&tmp.0, "resume-sim"))
        .unwrap();
    assert_eq!(loaded.spaces().len(), 1);
    assert_eq!(loaded.path_contexts().len(), 1);
    assert_eq!(loaded.events().len(), 1);
    assert_eq!(loaded.points().len(), 3);
    assert_eq!(loaded.path_nodes().len(), 3);
    assert_eq!(loaded.nodes().len(), 3);
    assert_eq!(loaded.spaces()[0].name, "resume-sim");
    assert_eq!(loaded.path_nodes().iter().map(|row| row.d).max(), Some(2));
}
