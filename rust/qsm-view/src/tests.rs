use m3data::{EventRow, NodeRow, PointRow};

use crate::{
    build_frame, camera_fit, color_rgb, display_rgb, event_texture_rgba, path_back_to_event,
    pick_marker, scene_from_rows, shade_rgb, Marker, SceneNode, ViewFilter, AXE_EXTRA, GRID_STEP,
    INACTIVE_DIMMER, NODE_RADIUS, ORIGIN_RADIUS,
};

fn point(id: i64, x: i32, y: i32, z: i32) -> PointRow {
    PointRow { id, x, y, z }
}

fn event(id: i32, color: i32) -> EventRow {
    EventRow {
        id,
        path_ctx_id: 1,
        creation_time: 0,
        color,
        end_time: 0,
    }
}

fn node(
    id: i64,
    point_id: i64,
    event_id: i32,
    d: i32,
    creation_time: i32,
    connection_mask: i32,
    node1: Option<i64>,
    node2: Option<i64>,
    node3: Option<i64>,
) -> NodeRow {
    NodeRow {
        id,
        event_id,
        path_node_id: id,
        trio_id: 1,
        point_id,
        d,
        creation_time,
        connection_mask,
        node1,
        node2,
        node3,
    }
}

/// `From` lives in the low 2 bits of the connection nibble.
fn from_on(conn: i32) -> i32 {
    1 << (conn * 4)
}

fn sample() -> Vec<SceneNode> {
    let points = vec![
        point(1, 0, 0, 0),
        point(2, 3, 0, 0),
        point(3, 6, 3, 0),
        point(4, 0, 6, 0),
    ];
    let events = vec![event(1, 1), event(2, 2)];
    let nodes = vec![
        node(10, 1, 1, 0, 0, 0, None, None, None),
        node(11, 2, 1, 1, 1, from_on(0), Some(10), None, None),
        node(12, 3, 1, 2, 2, from_on(1), None, Some(11), None),
        node(20, 4, 2, 0, 0, 0, None, None, None),
        // Next (state 2) must not become a line, even with an id stored.
        node(21, 2, 2, 1, 1, 2, Some(20), None, None),
    ];
    scene_from_rows(&points, &events, &nodes)
}

#[test]
fn from_links_and_origins_come_from_rows() {
    let scene = sample();
    let by_id: std::collections::HashMap<_, _> = scene.iter().map(|n| (n.id, n)).collect();
    assert!(by_id[&10].origin);
    assert!(!by_id[&11].origin);
    assert_eq!(by_id[&11].from, vec![10]);
    assert_eq!(by_id[&12].from, vec![11]);
    assert!(by_id[&21].from.is_empty(), "Next is not a drawn link");
    assert_eq!(by_id[&10].color, 1);
    assert_eq!(by_id[&20].color, 2);
    assert_eq!(by_id[&11].position, [3, 0, 0]);
    assert_eq!(by_id[&10].d, 0);
    assert_eq!(by_id[&12].d, 2);
    assert_eq!(by_id[&11].event_id, 1);
    assert_eq!(by_id[&11].path_ctx_id, 1);
}

#[test]
fn rows_without_a_point_or_event_are_skipped() {
    let scene = scene_from_rows(
        &[point(1, 1, 0, 0)],
        &[event(1, 4)],
        &[
            node(1, 1, 1, 0, 0, 0, None, None, None),
            node(2, 99, 1, 0, 0, 0, None, None, None),
            node(3, 1, 99, 0, 0, 0, None, None, None),
        ],
    );
    assert_eq!(scene.len(), 1);
    assert_eq!(scene[0].id, 1);
    assert_eq!(scene[0].color, 4);
}

#[test]
fn time_hides_later_nodes_and_their_links() {
    let scene = sample();
    let frame = build_frame(&scene, &ViewFilter::at_time(0), 18);
    let ids: Vec<_> = frame.markers.iter().map(|m| m.node_id).collect();
    assert_eq!(ids, vec![10, 20]);
    assert_eq!(frame.origin_count(), 2);
    assert_eq!(frame.line_count(), 0);

    let at_one = build_frame(&scene, &ViewFilter::at_time(1), 18);
    assert_eq!(at_one.node_count(), 4);
    assert_eq!(at_one.line_count(), 1);
    assert_eq!(at_one.segments[0].color, 1);
    assert_eq!(at_one.segments[0].a, [0.0, 0.0, 0.0]);
    assert_eq!(at_one.segments[0].b, [3.0, 0.0, 0.0]);

    let at_two = build_frame(&scene, &ViewFilter::at_time(2), 18);
    assert_eq!(at_two.node_count(), 5);
    assert_eq!(at_two.line_count(), 2);
}

#[test]
fn color_mask_hides_that_color_including_roots_and_touching_lines() {
    let scene = sample();
    let mut filter = ViewFilter::at_time(2);
    filter.color_mask = 2;
    let frame = build_frame(&scene, &filter, 18);
    assert!(frame.markers.iter().all(|m| m.color == 2));
    assert_eq!(frame.origin_count(), 1);
    assert_eq!(frame.line_count(), 0, "the green Next link is not drawn");

    filter.color_mask = 1;
    let red = build_frame(&scene, &filter, 18);
    assert_eq!(red.node_count(), 3);
    assert_eq!(red.line_count(), 2);
    assert!(red.markers.iter().any(|m| m.origin && m.color == 1));
}

#[test]
fn a_mutual_from_link_is_one_segment() {
    let nodes = vec![
        SceneNode {
            id: 1,
            position: [0, 0, 0],
            color: 1,
            creation_time: 0,
            origin: true,
            d: 0,
            event_id: 1,
            path_ctx_id: 7,
            from: vec![2],
        },
        SceneNode {
            id: 2,
            position: [3, 0, 0],
            color: 2,
            creation_time: 1,
            origin: false,
            d: 1,
            event_id: 2,
            path_ctx_id: 8,
            from: vec![1],
        },
    ];
    let frame = build_frame(&nodes, &ViewFilter::at_time(1), 9);
    assert_eq!(frame.line_count(), 1);
}

#[test]
fn line_and_axis_toggles() {
    let scene = sample();
    let mut filter = ViewFilter::at_time(2);
    filter.show_lines = false;
    let frame = build_frame(&scene, &filter, 6);
    assert_eq!(frame.line_count(), 0);
    assert!(frame.node_count() > 0);
    assert!(frame.axes.len() > 3);

    filter.show_axes = false;
    let bare = build_frame(&scene, &filter, 6);
    assert!(bare.axes.is_empty());
}

#[test]
fn axes_use_go_length_and_colors() {
    let frame = build_frame(&[], &ViewFilter::at_time(0), 18);
    let grid = (18 / GRID_STEP) as usize * 2 * 6;
    assert_eq!(frame.axes.len(), 3 + grid);
    let len = (18 + AXE_EXTRA) as f32;
    assert_eq!(frame.axes[0].color, 1);
    assert_eq!(frame.axes[0].a, [-len, 0.0, 0.0]);
    assert_eq!(frame.axes[0].b, [len, 0.0, 0.0]);
    assert_eq!(frame.axes[1].color, 2);
    assert_eq!(frame.axes[2].color, 4);
    assert!(frame.axes.iter().skip(3).all(|seg| seg.color == 0));
}

#[test]
fn event_colors_keep_their_channel() {
    let red = color_rgb(1);
    let green = color_rgb(2);
    let blue = color_rgb(4);
    let yellow = color_rgb(8);
    assert!(red[0] > red[1] && red[0] > red[2]);
    assert!(green[1] > green[0] && green[1] > green[2]);
    assert!(blue[2] > blue[0] && blue[2] > blue[1]);
    assert!(yellow[0] > 0.5 && yellow[1] > 0.5 && yellow[2] < 0.4);
    assert_eq!(NODE_RADIUS, 0.2);
    assert_eq!(ORIGIN_RADIUS, 0.5);
    let active = display_rgb(1, true);
    let inactive = display_rgb(1, false);
    assert_eq!(inactive[0], active[0] * INACTIVE_DIMMER);
    assert!(inactive[0] < active[0]);
    let selected = shade_rgb(1, true, true);
    assert!(selected[0] > active[0]);
    let bands = event_texture_rgba(32);
    assert_eq!(bands.len(), 32 * 32 * 4);
    let samples: std::collections::HashSet<u8> = bands.iter().step_by(4).copied().collect();
    assert!(samples.contains(&255));
    assert!(samples.contains(&40));
}

#[test]
fn latest_only_keeps_the_current_time_and_the_selected_path() {
    let scene = sample();
    let mut filter = ViewFilter::at_time(2);
    filter.latest_only = true;
    let only = build_frame(&scene, &filter, 18);
    assert_eq!(
        only.markers.iter().map(|m| m.node_id).collect::<Vec<_>>(),
        vec![12]
    );
    assert!(only.markers[0].active);
    assert_eq!(only.line_count(), 0);

    filter.selected = Some(12);
    let with_path = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = with_path.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![10, 11, 12]);
    assert_eq!(path_back_to_event(&scene, 12), vec![12, 11, 10]);
    assert_eq!(with_path.line_count(), 2);
    let active: Vec<_> = with_path
        .markers
        .iter()
        .filter(|m| m.active)
        .map(|m| m.node_id)
        .collect();
    assert_eq!(active, vec![12]);
    assert!(with_path
        .markers
        .iter()
        .any(|m| m.selected && m.node_id == 12));

    filter.color_mask = 2;
    let hidden = build_frame(&scene, &filter, 18);
    assert!(
        hidden.markers.is_empty(),
        "a red path stays hidden when red is off"
    );
}

#[test]
fn a_ray_hits_the_nearest_sphere() {
    let markers = vec![
        Marker {
            node_id: 1,
            position: [0.0, 0.0, 0.0],
            color: 1,
            origin: true,
            active: true,
            selected: false,
        },
        Marker {
            node_id: 2,
            position: [3.0, 0.0, 0.0],
            color: 1,
            origin: false,
            active: true,
            selected: false,
        },
    ];
    assert_eq!(
        pick_marker(&markers, [0.0, 0.0, 5.0], [0.0, 0.0, -1.0]),
        Some(1)
    );
    assert_eq!(
        pick_marker(&markers, [-5.0, 0.0, 0.0], [1.0, 0.0, 0.0]),
        Some(1)
    );
    assert_eq!(
        pick_marker(&markers, [0.0, 2.0, 5.0], [0.0, 0.0, -1.0]),
        None
    );
}

#[test]
fn camera_fits_the_loaded_extent() {
    let scene = sample();
    let (center, distance) = camera_fit(&scene);
    assert_eq!(center[0], 3.0);
    assert!(distance > 6.0);
}
