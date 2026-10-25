use m3data::{EventRow, NodeRow, PointRow};

use crate::{
    build_frame, camera_fit, color_name, color_rgb, cycle_min_colors, display_rgb,
    event_texture_rgba, node_is_active, occupants_at, path_back_to_event, paths_for_selection,
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
fn color_mask_keeps_event_roots_and_hides_other_nodes() {
    let scene = sample();
    let mut filter = ViewFilter::at_time(2);
    filter.color_mask = 2;
    let frame = build_frame(&scene, &filter, 18);
    assert!(frame.markers.iter().any(|m| m.origin && m.color == 1));
    assert!(frame.markers.iter().any(|m| m.origin && m.color == 2));
    assert!(frame
        .markers
        .iter()
        .filter(|m| !m.origin)
        .all(|m| m.color == 2));
    assert_eq!(frame.origin_count(), 2);
    assert_eq!(frame.line_count(), 0, "the green Next link is not drawn");

    filter.color_mask = 1;
    let red = build_frame(&scene, &filter, 18);
    assert_eq!(red.node_count(), 4);
    assert_eq!(red.line_count(), 2);
    assert!(red.markers.iter().any(|m| m.origin && m.color == 2));
    assert!(red
        .markers
        .iter()
        .filter(|m| !m.origin)
        .all(|m| m.color == 1));
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
fn active_means_age_below_the_saved_threshold() {
    assert!(node_is_active(4, 6, 3));
    assert!(!node_is_active(3, 6, 3));
    assert!(node_is_active(6, 6, 3));
    assert!(!node_is_active(7, 6, 3), "a future node is not active");
    assert!(!node_is_active(6, 6, 0));
    assert!(node_is_active(2, 2, 1));
    assert!(!node_is_active(1, 2, 1));
}

#[test]
fn active_only_keeps_roots_and_the_selected_path() {
    let scene = sample();
    let mut filter = ViewFilter::at_time(2);
    filter.active_only = true;
    filter.active_threshold = 1;
    let only = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = only.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![10, 12, 20]);
    let active: Vec<_> = only
        .markers
        .iter()
        .filter(|m| m.active)
        .map(|m| m.node_id)
        .collect();
    assert_eq!(active, vec![10, 12, 20]);
    assert!(only.markers.iter().filter(|m| m.origin).all(|m| m.active));
    assert_eq!(only.line_count(), 0);

    filter.selected = Some(12);
    let with_path = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = with_path.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![10, 11, 12, 20]);
    assert_eq!(path_back_to_event(&scene, 12), vec![12, 11, 10]);
    assert_eq!(with_path.line_count(), 2);
    let active: Vec<_> = with_path
        .markers
        .iter()
        .filter(|m| m.active)
        .map(|m| m.node_id)
        .collect();
    assert_eq!(active, vec![10, 12, 20]);
    assert!(with_path
        .markers
        .iter()
        .any(|m| m.selected && m.node_id == 12));
    assert!(with_path
        .markers
        .iter()
        .any(|m| m.node_id == 11 && !m.active));

    filter.color_mask = 2;
    let hidden = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = hidden.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(
        ids,
        vec![10, 20],
        "a red path stays hidden when red is off; both roots stay"
    );
}

#[test]
fn a_root_created_after_the_slider_stays_hidden() {
    let mut scene = sample();
    scene.push(SceneNode {
        id: 30,
        position: [1, 0, 0],
        color: 4,
        creation_time: 5,
        origin: true,
        d: 0,
        event_id: 3,
        path_ctx_id: 1,
        from: vec![],
    });
    let frame = build_frame(&scene, &ViewFilter::at_time(2), 18);
    assert!(frame.markers.iter().all(|m| m.node_id != 30));
}

#[test]
fn multi_color_filter_keeps_active_shared_points_and_roots() {
    let scene = sample();
    let mut filter = ViewFilter::at_time(2);
    filter.min_colors = 2;
    filter.active_threshold = 3;
    let frame = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = frame.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![10, 11, 20, 21]);
    assert!(frame.markers.iter().filter(|m| !m.origin).all(|m| m.active));

    filter.min_colors = 3;
    let three = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = three.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![10, 20]);

    filter.min_colors = 4;
    assert_eq!(build_frame(&scene, &filter, 18).node_count(), 2);

    filter.min_colors = 2;
    filter.active_threshold = 1;
    let inactive = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = inactive.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![10, 20], "inactive nodes count, but are not drawn");

    let mut doubled = scene.clone();
    doubled.push(SceneNode {
        id: 13,
        position: [6, 3, 0],
        color: 1,
        creation_time: 2,
        origin: false,
        d: 2,
        event_id: 3,
        path_ctx_id: 1,
        from: vec![],
    });
    filter.active_threshold = 3;
    let still = build_frame(&doubled, &filter, 18);
    assert!(still
        .markers
        .iter()
        .all(|m| m.node_id != 12 && m.node_id != 13));

    filter.active_only = true;
    filter.selected = Some(12);
    let with_path = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = with_path.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![10, 11, 12, 20, 21]);

    let people = occupants_at(&scene, [3, 0, 0], 2, 3);
    let mut colors: Vec<_> = people.iter().map(|p| (p.color, p.event_id)).collect();
    colors.sort();
    assert_eq!(colors, vec![(1, 1), (2, 2)]);
    assert!(people.iter().all(|p| p.active));
    assert_eq!(color_name(1), "red");
    assert_eq!(color_name(2), "green");
    assert!(occupants_at(&scene, [3, 0, 0], 0, 3).is_empty());

    let opening = ViewFilter::opening(6, 3);
    assert!(opening.active_only);
    assert_eq!(opening.active_threshold, 3);
    assert_eq!(opening.min_colors, 0);
    assert!(!ViewFilter::at_time(6).active_only);
    assert_eq!(cycle_min_colors(0), 2);
    assert_eq!(cycle_min_colors(2), 3);
    assert_eq!(cycle_min_colors(3), 4);
    assert_eq!(cycle_min_colors(4), 0);
}

fn placed(
    id: i64,
    position: [i32; 3],
    color: u8,
    creation_time: i32,
    origin: bool,
    event_id: i32,
    from: Vec<i64>,
) -> SceneNode {
    SceneNode {
        id,
        position,
        color,
        creation_time,
        origin,
        d: if origin { 0 } else { 1 },
        event_id,
        path_ctx_id: event_id,
        from,
    }
}

#[test]
fn a_shared_point_draws_every_color_path_back_to_its_root() {
    let scene = vec![
        placed(1, [0, 0, 0], 1, 0, true, 1, vec![]),
        placed(2, [3, 0, 0], 1, 1, false, 1, vec![1]),
        placed(5, [6, 0, 0], 1, 2, false, 1, vec![2]),
        placed(3, [0, 6, 0], 2, 0, true, 2, vec![]),
        placed(4, [3, 0, 0], 2, 1, false, 2, vec![3]),
        placed(6, [6, 3, 0], 2, 2, false, 2, vec![4]),
        placed(7, [9, 0, 0], 2, 1, false, 2, vec![3]),
    ];
    let mut filter = ViewFilter::at_time(2);
    filter.active_only = true;
    filter.active_threshold = 1;
    filter.selected = Some(2);
    let paths = paths_for_selection(&scene, 2, 2);
    assert_eq!(paths, vec![vec![2, 1], vec![4, 3]]);

    let frame = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = frame.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![1, 2, 3, 4, 5, 6]);
    assert!(frame.markers.iter().all(|m| m.node_id != 7));
    assert!(frame.markers.iter().filter(|m| m.origin).all(|m| m.active));
    assert!(frame.markers.iter().any(|m| m.node_id == 4 && !m.active));
    let colors: Vec<_> = frame
        .segments
        .iter()
        .map(|seg| (seg.color, seg.a, seg.b))
        .collect();
    assert!(
        colors.iter().any(|seg| seg.0 == 1),
        "the red path keeps its color: {colors:?}"
    );
    assert!(
        colors.iter().any(|seg| seg.0 == 2),
        "the green path keeps its color: {colors:?}"
    );

    filter.selected = Some(5);
    let only_red = build_frame(&scene, &filter, 18);
    assert!(only_red.markers.iter().all(|m| m.node_id != 4));
    assert_eq!(paths_for_selection(&scene, 5, 2), vec![vec![5, 2, 1]]);

    filter.selected = Some(2);
    filter.color_mask = 1;
    let masked = build_frame(&scene, &filter, 18);
    assert!(masked.markers.iter().all(|m| m.node_id != 4));
    assert!(masked
        .markers
        .iter()
        .any(|m| m.node_id == 3 && m.origin && m.active));

    filter.selected = Some(2);
    filter.color_mask = 1 | 2 | 4 | 8;
    filter.active_only = false;
    let shown = build_frame(&scene, &filter, 18);
    assert_eq!(shown.node_count(), scene.len());
}

#[test]
fn two_nodes_of_one_color_draw_only_the_selected_path() {
    let scene = vec![
        placed(1, [0, 0, 0], 1, 0, true, 1, vec![]),
        placed(2, [3, 0, 0], 1, 1, false, 1, vec![1]),
        placed(5, [6, 0, 0], 1, 2, false, 1, vec![2]),
        placed(8, [0, 3, 0], 1, 0, true, 3, vec![]),
        placed(9, [3, 0, 0], 1, 1, false, 3, vec![8]),
    ];
    assert_eq!(paths_for_selection(&scene, 2, 2), vec![vec![2, 1]]);
    let mut filter = ViewFilter::at_time(2);
    filter.active_only = true;
    filter.active_threshold = 1;
    filter.selected = Some(2);
    let frame = build_frame(&scene, &filter, 18);
    let mut ids: Vec<_> = frame.markers.iter().map(|m| m.node_id).collect();
    ids.sort();
    assert_eq!(ids, vec![1, 2, 5, 8]);
    assert!(frame.markers.iter().all(|m| m.node_id != 9));
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
