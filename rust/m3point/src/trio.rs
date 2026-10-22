//! Trio index and the three connections that leave a node.

use std::fmt;

use crate::conn::{is_less_conn_id, ConnectionDetails, ConnectionId};
use crate::point::DInt;

/// Index of a trio. `255` is the unset value (`NilTrioIndex`).
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct TrioIndex(pub u8);

impl TrioIndex {
    pub const NIL: TrioIndex = TrioIndex(255);

    pub fn is_base(self) -> bool {
        self.0 < 8
    }
}

impl fmt::Display for TrioIndex {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "T{:03}", self.0)
    }
}

pub const NB_TRIO_DS_INDEX: usize = 7;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct TrioDetails {
    pub id: TrioIndex,
    pub conns: [ConnectionDetails; 3],
}

impl TrioDetails {
    pub fn connection_idx(self, conn_id: ConnectionId) -> i32 {
        for (idx, c) in self.conns.iter().enumerate() {
            if c.id == conn_id {
                return idx as i32;
            }
        }
        -1
    }

    pub fn has_connection(self, conn_id: ConnectionId) -> bool {
        self.conns.iter().any(|c| c.id == conn_id)
    }

    pub fn has_connections(self, ids: &[ConnectionId]) -> bool {
        ids.iter().all(|id| self.has_connection(*id))
    }

    /// The two connections that are not `conn_id`. `conn_id` is the incoming (negative) one.
    pub fn other_connections_from(self, conn_id: ConnectionId) -> [Option<ConnectionDetails>; 2] {
        let mut res = [None, None];
        if !self.has_connection(conn_id) {
            return res;
        }
        let mut idx = 0;
        for c in &self.conns {
            if c.id != conn_id {
                res[idx] = Some(*c);
                idx += 1;
            }
        }
        res
    }

    pub fn last_other_connection(self, c_ids: &[ConnectionId]) -> Option<ConnectionDetails> {
        for c in &self.conns {
            if !c_ids.iter().any(|id| c.id == *id) {
                return Some(*c);
            }
        }
        None
    }

    pub fn is_base_trio(self) -> bool {
        self.id.0 < 8
    }

    /// Class of a trio from the squared lengths of its three connections.
    /// Returns -1 when the pattern is not one of the seven legal classes.
    pub fn ds_index(self) -> i32 {
        if self.conns[0].distance_squared() == 1 {
            match self.conns[1].distance_squared() {
                1 => return 1,
                2 => match self.conns[2].distance_squared() {
                    3 => return 2,
                    5 => return 3,
                    _ => {}
                },
                _ => {}
            }
        } else {
            match self.conns[1].distance_squared() {
                2 => return 0,
                3 => match self.conns[2].distance_squared() {
                    3 => return 4,
                    5 => return 5,
                    _ => {}
                },
                5 => return 6,
                _ => {}
            }
        }
        -1
    }
}

impl fmt::Display for TrioDetails {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "T{:02}: ({}, {}, {})",
            self.id.0, self.conns[0], self.conns[1], self.conns[2]
        )
    }
}

pub fn sort_three_conns(cds: &mut [ConnectionDetails; 3]) {
    cds.sort_by(|a, b| {
        if is_less_conn_id(a, b) {
            std::cmp::Ordering::Less
        } else if is_less_conn_id(b, a) {
            std::cmp::Ordering::Greater
        } else {
            std::cmp::Ordering::Equal
        }
    });
}

/// Allowed squared-length triples, in `GetDSIndex` order.
pub fn possible_ds_arrays() -> [[DInt; 3]; NB_TRIO_DS_INDEX] {
    [
        [2, 2, 2],
        [1, 1, 2],
        [1, 2, 3],
        [1, 2, 5],
        [2, 3, 3],
        [2, 3, 5],
        [2, 5, 5],
    ]
}
