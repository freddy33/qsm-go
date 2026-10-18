//! Space events on the integer lattice.
//!
//! The database-backed time stepper from Go `spacedb` is not ported. The rules
//! that decide which events meet, how colors combine, and how a pyramid is
//! extracted from those meetings are.

use m3path::ConnectionState;
use m3point::{CInt, DInt, Point};
use m3util::count_the_ones;
use std::collections::BTreeMap;

pub type DistAndTime = i32;
pub const ZERO_DIST_AND_TIME: DistAndTime = 0;
pub const MIN_MAX_COORD: CInt = 2 * 9;

pub type EventId = i32;
pub const NIL_EVENT: EventId = -1;

#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
#[repr(u8)]
pub enum EventColor {
    Red = 1,
    Green = 2,
    Blue = 4,
    Yellow = 8,
}

pub const ALL_COLORS: [EventColor; 4] = [
    EventColor::Red,
    EventColor::Green,
    EventColor::Blue,
    EventColor::Yellow,
];

impl EventColor {
    pub fn from_u8(v: u8) -> Option<Self> {
        match v {
            1 => Some(Self::Red),
            2 => Some(Self::Green),
            4 => Some(Self::Blue),
            8 => Some(Self::Yellow),
            _ => None,
        }
    }
}

/// One event contribution sitting on a space-time node.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct NodeEvent {
    pub event_id: EventId,
    pub color: EventColor,
    pub is_root: bool,
    pub creation_time: DistAndTime,
}

/// Color mask of a node. A root contributes only its own color, matching `GetColorMask`.
pub fn color_mask(events: &[NodeEvent]) -> u8 {
    if events.is_empty() {
        return 0;
    }
    for ev in events {
        if ev.is_root {
            return ev.color as u8;
        }
    }
    let mut mask = 0u8;
    for ev in events {
        mask |= ev.color as u8;
    }
    mask
}

pub fn how_many_colors(events: &[NodeEvent]) -> u8 {
    count_the_ones(color_mask(events))
}

pub fn event_ids(events: &[NodeEvent]) -> Vec<EventId> {
    events.iter().map(|e| e.event_id).collect()
}

/// A node link is active when its id is positive and it is younger than `threshold`.
pub fn link_is_active(
    link_id: i64,
    current_time: DistAndTime,
    creation_time: DistAndTime,
    threshold: DistAndTime,
) -> bool {
    threshold != 0 && link_id > 0 && current_time - creation_time < threshold
}

/// Count active links, counting each connection id once, as `GetNbActiveLinks` does.
pub fn count_active_links(
    links: &[(m3point::ConnectionId, i64, DistAndTime)],
    current: DistAndTime,
    threshold: DistAndTime,
) -> usize {
    let mut seen = Vec::new();
    let mut n = 0;
    for (conn, link_id, created) in links {
        if link_is_active(*link_id, current, *created, threshold) && !seen.contains(conn) {
            seen.push(*conn);
            n += 1;
        }
    }
    n
}

/// Three event ids, sorted ascending. `NilThreeIds` is three nil events.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Hash)]
pub struct ThreeIds(pub [EventId; 3]);

pub const NIL_THREE_IDS: ThreeIds = ThreeIds([NIL_EVENT, NIL_EVENT, NIL_EVENT]);

impl ThreeIds {
    pub fn contains(self, id: EventId) -> bool {
        self.0.contains(&id)
    }
}

pub fn sort_event_ids(ids: &mut [EventId]) {
    ids.sort_unstable();
}

/// Every unordered triple inside a sorted event-id list of length 3 or 4.
pub fn make_three_ids(mut ids: Vec<EventId>) -> Vec<ThreeIds> {
    sort_event_ids(&mut ids);
    if ids.len() == 3 {
        vec![ThreeIds([ids[0], ids[1], ids[2]])]
    } else if ids.len() == 4 {
        vec![
            ThreeIds([ids[0], ids[1], ids[2]]),
            ThreeIds([ids[0], ids[2], ids[3]]),
            ThreeIds([ids[0], ids[1], ids[3]]),
            ThreeIds([ids[1], ids[2], ids[3]]),
        ]
    } else {
        panic!("WHAT!");
    }
}

/// Index combinations used to place four pyramid events.
///
/// The returned pair is the list of 4-tuples and the twelve counters
/// `(expected, actual)` for total, t1, t2a, t2b, t3, t4.
pub fn create_all_indexes(nb_indexes: i32) -> (Vec<[i32; 4]>, [i32; 12]) {
    let (t4, t3, t2a, t2b, t1) = if nb_indexes == 8 {
        (8, 56, 28, 168, 70)
    } else if nb_indexes == 12 {
        (12, 132, 66, 660, 495)
    } else {
        panic!("Nb indexes {nb_indexes} is not supported");
    };
    let nb = (t4 + t3 + t2a + t2b + t1) as usize;
    let mut res = vec![[0; 4]; nb];
    let mut idx = 0usize;
    let mut nb_t4 = 0;
    let mut nb_t3 = 0;
    let mut nb_t2a = 0;
    let mut nb_t2b = 0;
    let mut nb_t1 = 0;
    for i1 in 0..nb_indexes {
        res[idx] = [i1, i1, i1, i1];
        idx += 1;
        nb_t4 += 1;
        for i_t3 in 0..nb_indexes {
            if i_t3 != i1 {
                res[idx] = [i1, i1, i1, i_t3];
                idx += 1;
                nb_t3 += 1;
            }
        }
        for i2 in (i1 + 1)..nb_indexes {
            res[idx] = [i1, i1, i2, i2];
            idx += 1;
            nb_t2a += 1;
            for i_t2b in 0..nb_indexes {
                if i_t2b != i1 && i_t2b != i2 {
                    res[idx] = [i1, i2, i_t2b, i_t2b];
                    idx += 1;
                    nb_t2b += 1;
                }
            }
            for i3 in (i2 + 1)..nb_indexes {
                for i4 in (i3 + 1)..nb_indexes {
                    res[idx] = [i1, i2, i3, i4];
                    idx += 1;
                    nb_t1 += 1;
                }
            }
        }
    }
    (
        res,
        [
            (t4 + t3 + t2a + t2b + t1),
            idx as i32,
            t1,
            nb_t1,
            t2a,
            nb_t2a,
            t2b,
            nb_t2b,
            t3,
            nb_t3,
            t4,
            nb_t4,
        ],
    )
}

pub type Pyramid = [Point; 4];

pub fn pyramid_size(points: Pyramid) -> DInt {
    let mut total = 0;
    total += m3point::make_vector(points[0], points[1]).distance_squared();
    total += m3point::make_vector(points[0], points[2]).distance_squared();
    total += m3point::make_vector(points[0], points[3]).distance_squared();
    total += m3point::make_vector(points[1], points[2]).distance_squared();
    total += m3point::make_vector(points[1], points[3]).distance_squared();
    total += m3point::make_vector(points[2], points[3]).distance_squared();
    total
}

pub fn order_pyramid(pyramid: Pyramid) -> Pyramid {
    let mut slice = pyramid;
    slice.sort_by(|i_p, j_p| {
        if i_p.x() != j_p.x() {
            return i_p.x().cmp(&j_p.x());
        }
        if i_p.y() != j_p.y() {
            return i_p.y().cmp(&j_p.y());
        }
        i_p.z().cmp(&j_p.z())
    });
    slice
}

/// Collect every pyramid that can be built by picking one point from each of
/// four three-id groups. Ties on the shortest group are broken by `ThreeIds` order
/// so the result does not depend on hash iteration.
pub fn build_pyramids(points: &BTreeMap<ThreeIds, Vec<Point>>) -> BTreeMap<Pyramid, DInt> {
    let mut out = BTreeMap::new();
    if points.len() < 4 {
        return out;
    }
    create_pyramids(
        points,
        &[Point::default(); 4],
        0,
        points.len() as i32 - 4,
        &mut out,
    );
    out
}

fn create_pyramids(
    current: &BTreeMap<ThreeIds, Vec<Point>>,
    pyramid: &Pyramid,
    pos: usize,
    possible_skip: i32,
    out: &mut BTreeMap<Pyramid, DInt>,
) {
    let cur_length = current.len();
    if cur_length == 0 {
        panic!("Should never reach here with an empty map");
    }
    if cur_length == 1 && pos != 3 {
        panic!("Reached the end of the map but not the end of the pyramid at {pos}");
    }
    if pos == 3 {
        for points in current.values() {
            for picked in points {
                let mut new_pyramid = *pyramid;
                new_pyramid[pos] = *picked;
                let ordered = order_pyramid(new_pyramid);
                out.insert(ordered, pyramid_size(ordered));
            }
        }
        return;
    }
    let (small, _) = current
        .iter()
        .min_by_key(|(ids, pts)| (pts.len(), *ids))
        .expect("map is not empty");
    let small = *small;
    if possible_skip > 0 {
        let mut next = current.clone();
        next.remove(&small);
        create_pyramids(&next, pyramid, pos, possible_skip - 1, out);
    }
    let picked_points = current[&small].clone();
    for picked in picked_points {
        let mut new_pyramid = *pyramid;
        new_pyramid[pos] = picked;
        let mut next = BTreeMap::new();
        for (ids, pts) in current {
            if *ids != small {
                next.insert(*ids, pts.iter().copied().filter(|p| *p != picked).collect());
            }
        }
        create_pyramids(&next, &new_pyramid, pos + 1, possible_skip, out);
    }
}

/// The four centers of a demo pyramid of the given scale, in creation order.
pub fn pyramid_centers(scale: CInt) -> [Point; 4] {
    [
        Point([3, 0, 3]).mul(scale),
        Point([-3, 3, 3]).mul(scale),
        Point([-3, -3, 3]).mul(scale),
        Point([0, 0, -3]).mul(scale),
    ]
}

/// Active events are those whose creation time is `<= at` and whose end time is `>= at`.
pub fn active_at(creation: DistAndTime, end: DistAndTime, at: DistAndTime) -> bool {
    creation <= at && end >= at
}

/// A blocked connection does not grow an event node. Open and from/next do.
pub fn connection_grows(state: ConnectionState) -> bool {
    !matches!(state, ConnectionState::Blocked | ConnectionState::NotSet)
}

#[cfg(test)]
mod tests {
    use super::*;
    use m3point::{ConnectionId, GrowthType, ALL_GROWTH_TYPES};

    #[test]
    fn color_mask_root_wins_and_otherwise_unions() {
        let red = NodeEvent {
            event_id: 1,
            color: EventColor::Red,
            is_root: false,
            creation_time: 0,
        };
        let green = NodeEvent {
            event_id: 2,
            color: EventColor::Green,
            is_root: false,
            creation_time: 1,
        };
        let root = NodeEvent {
            event_id: 3,
            color: EventColor::Blue,
            is_root: true,
            creation_time: 2,
        };
        assert_eq!(color_mask(&[]), 0);
        assert_eq!(how_many_colors(&[red, green]), 2);
        assert_eq!(
            color_mask(&[red, green]),
            (EventColor::Red as u8) | (EventColor::Green as u8)
        );
        assert_eq!(color_mask(&[red, root, green]), EventColor::Blue as u8);
        assert_eq!(how_many_colors(&[red, root]), 1);
    }

    #[test]
    fn three_id_groups_for_three_and_four_events() {
        let three = make_three_ids(vec![3, 1, 2]);
        assert_eq!(three, vec![ThreeIds([1, 2, 3])]);
        assert!(three[0].contains(2));
        let four = make_three_ids(vec![4, 1, 3, 2]);
        assert_eq!(
            four,
            vec![
                ThreeIds([1, 2, 3]),
                ThreeIds([1, 3, 4]),
                ThreeIds([1, 2, 4]),
                ThreeIds([2, 3, 4]),
            ]
        );
    }

    #[test]
    fn create_all_indexes_counts_match() {
        for ctx in ALL_GROWTH_TYPES {
            let (_rows, idxs) = create_all_indexes(ctx.nb_indexes());
            for i in 0..idxs.len() / 2 {
                assert_eq!(idxs[i * 2], idxs[i * 2 + 1], "{ctx} bucket {i}");
            }
        }
        let (rows8, _) = create_all_indexes(8);
        assert_eq!(rows8.len(), 8 + 56 + 28 + 168 + 70);
        let (rows12, _) = create_all_indexes(12);
        assert_eq!(rows12.len(), 12 + 132 + 66 + 660 + 495);
        let _ = GrowthType(8);
    }

    #[test]
    fn pyramid_orders_points_and_sums_edges() {
        let centers = pyramid_centers(1);
        let ordered = order_pyramid(centers);
        assert!(ordered[0].x() <= ordered[1].x());
        let size = pyramid_size(ordered);
        assert!(size > 0);
        assert_eq!(pyramid_size(ordered), pyramid_size(centers));
    }

    #[test]
    fn pyramid_builder_picks_four_points() {
        let mut points = BTreeMap::new();
        points.insert(
            ThreeIds([1, 2, 3]),
            vec![Point([0, 0, 0]), Point([3, 0, 0])],
        );
        points.insert(ThreeIds([1, 2, 4]), vec![Point([0, 3, 0])]);
        points.insert(ThreeIds([1, 3, 4]), vec![Point([0, 0, 3])]);
        points.insert(ThreeIds([2, 3, 4]), vec![Point([3, 3, 3])]);
        let pyramids = build_pyramids(&points);
        assert!(!pyramids.is_empty());
        for (pyramid, size) in &pyramids {
            assert_eq!(*size, pyramid_size(*pyramid));
            assert_eq!(*pyramid, order_pyramid(*pyramid));
        }
    }

    #[test]
    fn active_link_threshold() {
        let links = [
            (ConnectionId(1), 10, 2),
            (ConnectionId(1), 11, 2),
            (ConnectionId(-2), 12, 2),
            (ConnectionId(3), -1, 2),
            (ConnectionId(4), 13, 0),
        ];
        assert!(!link_is_active(10, 3, 0, 0));
        assert!(link_is_active(10, 3, 0, 4));
        assert!(!link_is_active(-1, 3, 0, 4));
        // Current time 5, threshold 4: the two young ids count once each; the
        // duplicate of connection 1, the negative link id, and the old link do not.
        assert_eq!(count_active_links(&links, 5, 4), 2);
    }

    #[test]
    fn growth_skips_blocked_and_unset_connections() {
        assert!(!connection_grows(ConnectionState::NotSet));
        assert!(connection_grows(ConnectionState::From));
        assert!(connection_grows(ConnectionState::Next));
        assert!(!connection_grows(ConnectionState::Blocked));
        assert!(active_at(0, 5, 5));
        assert!(!active_at(2, 5, 1));
    }
}
