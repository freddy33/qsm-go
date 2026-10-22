//! Path contexts and the 4-bit connection state packed into a `u16`.
//!
//! Three connections share one mask. Each connection uses 4 bits; the low 2
//! bits are the [`ConnectionState`]. The bit operations match `m3path` and
//! `ConnectionsStateDb.SetConnectionState`. [`PathGrowth`] is the in-memory
//! port of `PathContextDb.calculateNextMaxDist`.

use m3point::{GrowthType, Point, ORIGIN};
use std::fmt;

pub const NB_CONNECTIONS: usize = 3;
pub const CONNECTION_MASK_BITS: u16 = 4;
pub const CONNECTION_STATE_MASK: u16 = 0x0003;
pub const SINGLE_CONNECTION_MASK: u16 = 0x000F;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
#[repr(u16)]
pub enum ConnectionState {
    NotSet = 0x0000,
    From = 0x0001,
    Next = 0x0002,
    Blocked = 0x0003,
}

pub const NEW_PATH_NODE_ID: i64 = -1;
pub const IN_POOL_ID: i64 = -2;
pub const LINK_ID_NOT_SET: i64 = -3;
pub const DEAD_END_ID: i64 = -4;
pub const NEXT_LINK_ID_NOT_ASSIGNED: i64 = -5;
pub const NIL_POINT_ID: i64 = -1;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct PathPoint {
    pub id: i64,
    pub p: Point,
}

impl PathPoint {
    pub const NIL: PathPoint = PathPoint {
        id: NIL_POINT_ID,
        p: ORIGIN,
    };
}

impl fmt::Display for PathPoint {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "PP={:04},{}", self.id, self.p)
    }
}

pub fn get_connection_mask_value(connection_mask: u16, conn_idx: i32) -> u16 {
    (connection_mask >> (conn_idx as u16 * CONNECTION_MASK_BITS)) & SINGLE_CONNECTION_MASK
}

pub fn get_connection_state(connection_mask: u16, conn_idx: i32) -> ConnectionState {
    match get_connection_mask_value(connection_mask, conn_idx) & CONNECTION_STATE_MASK {
        0 => ConnectionState::NotSet,
        1 => ConnectionState::From,
        2 => ConnectionState::Next,
        _ => ConnectionState::Blocked,
    }
}

/// Returns only the updated 4-bit nibble, as `m3path.SetConnectionState` does.
pub fn set_connection_state_nibble(
    connection_mask: u16,
    conn_idx: i32,
    state: ConnectionState,
) -> u16 {
    let mut new_conn_mask = get_connection_mask_value(connection_mask, conn_idx);
    new_conn_mask &= !CONNECTION_STATE_MASK;
    new_conn_mask |= state as u16;
    new_conn_mask
}

pub fn set_connection_mask(mut all: u16, conn_idx: i32, mask_value: u16) -> u16 {
    let shift = conn_idx as u16 * CONNECTION_MASK_BITS;
    all &= !(SINGLE_CONNECTION_MASK << shift);
    all |= mask_value << shift;
    all
}

/// One path node's three connection states, matching `ConnectionsStateDb`.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct ConnectionsState {
    pub trio_id: u8,
    pub connection_mask: u16,
    pub link_ids: [i64; NB_CONNECTIONS],
}

impl Default for ConnectionsState {
    fn default() -> Self {
        Self::new()
    }
}

impl ConnectionsState {
    pub fn new() -> Self {
        Self {
            trio_id: 255,
            connection_mask: ConnectionState::NotSet as u16,
            link_ids: [LINK_ID_NOT_SET; NB_CONNECTIONS],
        }
    }

    pub fn connection_state(self, conn_idx: i32) -> ConnectionState {
        get_connection_state(self.connection_mask, conn_idx)
    }

    pub fn set_connection_state(&mut self, conn_idx: i32, state: ConnectionState) {
        let nibble = set_connection_state_nibble(self.connection_mask, conn_idx, state);
        self.connection_mask = set_connection_mask(self.connection_mask, conn_idx, nibble);
    }

    pub fn has_open_connections(self) -> bool {
        (0..NB_CONNECTIONS as i32).any(|i| self.connection_state(i) == ConnectionState::NotSet)
    }

    pub fn is_from(self, conn_idx: i32) -> bool {
        self.connection_state(conn_idx) == ConnectionState::From
    }

    pub fn is_next(self, conn_idx: i32) -> bool {
        self.connection_state(conn_idx) == ConnectionState::Next
    }

    pub fn is_dead_end(self, conn_idx: i32) -> bool {
        self.connection_state(conn_idx) == ConnectionState::Blocked
    }
}

/// Link id stored on an event node for a path-node connection state.
/// `From` needs the database id of the source node, so it is returned as a request.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum EventLinkId {
    NotSet,
    NextUnassigned,
    DeadEnd,
    NeedsFromNode,
}

pub fn event_link_for_state(state: ConnectionState) -> EventLinkId {
    match state {
        ConnectionState::NotSet => EventLinkId::NotSet,
        ConnectionState::Next => EventLinkId::NextUnassigned,
        ConnectionState::Blocked => EventLinkId::DeadEnd,
        ConnectionState::From => EventLinkId::NeedsFromNode,
    }
}

/// Value written to a nullable link column. Go stores NULL for an unset, blocked,
/// or not-yet-assigned next link, and a positive id for `From` and an assigned `Next`.
pub fn stored_link(state: ConnectionState, link_id: i64) -> Option<i64> {
    match state {
        ConnectionState::NotSet | ConnectionState::Blocked => None,
        ConnectionState::Next | ConnectionState::From => {
            if link_id > 0 {
                Some(link_id)
            } else {
                None
            }
        }
    }
}

/// In-memory link id restored from a nullable column and the connection state.
pub fn restored_link(state: ConnectionState, stored: Option<i64>) -> i64 {
    match state {
        ConnectionState::NotSet => LINK_ID_NOT_SET,
        ConnectionState::Blocked => DEAD_END_ID,
        ConnectionState::Next => stored
            .filter(|id| *id > 0)
            .unwrap_or(NEXT_LINK_ID_NOT_ASSIGNED),
        ConnectionState::From => stored.filter(|id| *id > 0).unwrap_or(LINK_ID_NOT_SET),
    }
}

pub fn event_link_value(kind: EventLinkId, from_node_id: Option<i64>) -> i64 {
    match kind {
        EventLinkId::NotSet => LINK_ID_NOT_SET,
        EventLinkId::NextUnassigned => NEXT_LINK_ID_NOT_ASSIGNED,
        EventLinkId::DeadEnd => DEAD_END_ID,
        EventLinkId::NeedsFromNode => from_node_id.expect("from-link requires the source node id"),
    }
}

/// Quadratic estimate of how many path nodes exist at distance `d`.
///
/// The coefficients are the Go trend lines. The cast to `i32` truncates toward zero.
pub fn calculate_predicted_size(growth_type: GrowthType, d: i32) -> i32 {
    if d == 0 {
        return 3;
    }
    if d == 1 {
        return 6;
    }
    let buffer = 1.02_f32;
    let df = d as f32;
    // Coefficients are f32, matching the Go `float32` trend lines.
    let raw = if growth_type.0 == 2 {
        1.445_f32 * df * df - 0.065_f32 * df - 0.377_f32
    } else {
        1.775_f32 * df * df - 2.497_f32 * df + 5.039_f32
    };
    (raw * buffer) as i32
}

mod grow;
pub use grow::{LivePathNode, PathGrowth};

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn connection_mask_round_trip_matches_go_node_test() {
        let mut pn = ConnectionsState::new();
        assert_eq!(pn.connection_mask, ConnectionState::NotSet as u16);
        for i in 0..NB_CONNECTIONS as i32 {
            assert_eq!(pn.connection_state(i), ConnectionState::NotSet);
        }
        assert!(pn.has_open_connections());
        pn.set_connection_state(1, ConnectionState::From);
        assert_eq!(pn.connection_state(0), ConnectionState::NotSet);
        assert_eq!(pn.connection_state(1), ConnectionState::From);
        assert_eq!(pn.connection_state(2), ConnectionState::NotSet);
        assert!(pn.is_from(1));
        assert!(!pn.is_next(1));
        pn.set_connection_state(0, ConnectionState::Blocked);
        pn.set_connection_state(2, ConnectionState::Next);
        assert!(pn.is_dead_end(0));
        assert!(pn.is_next(2));
        assert!(!pn.has_open_connections());
    }

    #[test]
    fn event_link_ids_follow_connection_state() {
        assert_eq!(
            event_link_value(event_link_for_state(ConnectionState::NotSet), None),
            LINK_ID_NOT_SET
        );
        assert_eq!(
            event_link_value(event_link_for_state(ConnectionState::Next), None),
            NEXT_LINK_ID_NOT_ASSIGNED
        );
        assert_eq!(
            event_link_value(event_link_for_state(ConnectionState::Blocked), None),
            DEAD_END_ID
        );
        assert_eq!(
            event_link_value(event_link_for_state(ConnectionState::From), Some(42)),
            42
        );
    }

    #[test]
    fn predicted_size_anchors() {
        let t8 = GrowthType(8);
        let t2 = GrowthType(2);
        assert_eq!(calculate_predicted_size(t8, 0), 3);
        assert_eq!(calculate_predicted_size(t2, 1), 6);
        // d = 2, type 8: (1.775*4 - 2.497*2 + 5.039) * 1.02 = 7.2879 -> 7
        assert_eq!(calculate_predicted_size(t8, 2), 7);
        // d = 10, type 8: (177.5 - 24.97 + 5.039) * 1.02 = 160.720 -> 160
        assert_eq!(calculate_predicted_size(t8, 10), 160);
        // d = 10, type 2: (144.5 - 0.65 - 0.377) * 1.02 = 146.542 -> 146
        assert_eq!(calculate_predicted_size(t2, 10), 146);
        assert_eq!(
            calculate_predicted_size(GrowthType(1), 10),
            calculate_predicted_size(t8, 10)
        );
    }

    #[test]
    fn upper_nibble_of_a_connection_is_preserved() {
        let mut state = ConnectionsState::new();
        state.connection_mask = set_connection_mask(0, 0, 0b1101);
        state.set_connection_state(0, ConnectionState::From);
        assert_eq!(state.connection_state(0), ConnectionState::From);
        assert_eq!(get_connection_mask_value(state.connection_mask, 0), 0b1101);
    }
}
