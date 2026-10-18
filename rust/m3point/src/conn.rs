//! Connection ids and the six unit directions of the lattice.

use std::fmt;

use crate::point::{DInt, Point, ORIGIN, X_FIRST, Y_FIRST, Z_FIRST};
use crate::trio::TrioDetails;

/// Axis-aligned unit direction. Discriminant values match the Go iota order.
#[repr(u8)]
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Hash)]
pub enum UnitDirection {
    PlusX = 0,
    MinusX = 1,
    PlusY = 2,
    MinusY = 3,
    PlusZ = 4,
    MinusZ = 5,
}

impl UnitDirection {
    pub fn from_u8(v: u8) -> Self {
        match v {
            0 => Self::PlusX,
            1 => Self::MinusX,
            2 => Self::PlusY,
            3 => Self::MinusY,
            4 => Self::PlusZ,
            5 => Self::MinusZ,
            _ => panic!("Impossible! Did not find {v} unit direction"),
        }
    }

    pub fn all() -> [UnitDirection; 6] {
        [
            Self::PlusX,
            Self::MinusX,
            Self::PlusY,
            Self::MinusY,
            Self::PlusZ,
            Self::MinusZ,
        ]
    }

    pub fn opposite(self) -> Self {
        match self {
            Self::PlusX => Self::MinusX,
            Self::MinusX => Self::PlusX,
            Self::PlusY => Self::MinusY,
            Self::MinusY => Self::PlusY,
            Self::PlusZ => Self::MinusZ,
            Self::MinusZ => Self::PlusZ,
        }
    }

    pub fn first_point(self) -> Point {
        match self {
            Self::PlusX => X_FIRST,
            Self::MinusX => X_FIRST.neg(),
            Self::PlusY => Y_FIRST,
            Self::MinusY => Y_FIRST.neg(),
            Self::PlusZ => Z_FIRST,
            Self::MinusZ => Z_FIRST.neg(),
        }
    }

    /// Connection of a base trio that moves toward this unit direction.
    pub fn find_connection(self, td: &TrioDetails) -> ConnectionDetails {
        if !td.is_base_trio() {
            panic!("cannot look for {} conn on non base trio {}", self, td);
        }
        let (axis, axis_value) = match self {
            Self::PlusX => (0, 1),
            Self::MinusX => (0, -1),
            Self::PlusY => (1, 1),
            Self::MinusY => (1, -1),
            Self::PlusZ => (2, 1),
            Self::MinusZ => (2, -1),
        };
        for c in &td.conns {
            if c.vector.0[axis] == axis_value {
                return *c;
            }
        }
        panic!("Impossible! Did not find {self} unit direction in {td}");
    }
}

impl fmt::Display for UnitDirection {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(match self {
            Self::PlusX => "+X",
            Self::MinusX => "-X",
            Self::PlusY => "+Y",
            Self::MinusY => "-Y",
            Self::PlusZ => "+Z",
            Self::MinusZ => "-Z",
        })
    }
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, Default)]
pub struct ConnectionId(pub i8);

impl ConnectionId {
    pub const NIL: ConnectionId = ConnectionId(0);

    pub fn is_valid(self) -> bool {
        self.0 != 0
    }

    pub fn neg(self) -> Self {
        ConnectionId(-self.0)
    }

    pub fn pos(self) -> Self {
        if self.0 < 0 {
            ConnectionId(-self.0)
        } else {
            self
        }
    }

    pub fn is_base(self) -> bool {
        let p = self.pos().0;
        (4..=9).contains(&p)
    }
}

impl fmt::Display for ConnectionId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.0 < 0 {
            write!(f, "CN{:02}", -self.0)
        } else {
            write!(f, "CP{:02}", self.0)
        }
    }
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub struct ConnectionDetails {
    pub id: ConnectionId,
    pub vector: Point,
    pub conn_ds: DInt,
}

pub const EMPTY_CONN: ConnectionDetails = ConnectionDetails {
    id: ConnectionId::NIL,
    vector: ORIGIN,
    conn_ds: 0,
};

impl Default for ConnectionDetails {
    fn default() -> Self {
        EMPTY_CONN
    }
}

impl ConnectionDetails {
    pub fn is_valid(self) -> bool {
        self.id.is_valid()
    }

    pub fn neg_id(self) -> ConnectionId {
        self.id.neg()
    }

    pub fn pos_id(self) -> ConnectionId {
        self.id.pos()
    }

    pub fn is_base(self) -> bool {
        self.id.is_base()
    }

    /// The two unit directions of a base connection (squared length 2).
    pub fn directions(self) -> [UnitDirection; 2] {
        if !self.is_base() {
            panic!(
                "cannot extract unit directions on non base connection {}",
                self.id
            );
        }
        let mut idx = 0;
        let mut res = [UnitDirection::PlusX; 2];
        let c_vec = self.vector;
        match c_vec.x() {
            0 => {}
            1 => {
                res[idx] = UnitDirection::PlusX;
                idx += 1;
            }
            -1 => {
                res[idx] = UnitDirection::MinusX;
                idx += 1;
            }
            _ => {}
        }
        match c_vec.y() {
            0 => {}
            1 => {
                res[idx] = UnitDirection::PlusY;
                idx += 1;
            }
            -1 => {
                res[idx] = UnitDirection::MinusY;
                idx += 1;
            }
            _ => {}
        }
        match c_vec.z() {
            0 => {}
            1 => {
                res[idx] = UnitDirection::PlusZ;
                idx += 1;
            }
            -1 => {
                res[idx] = UnitDirection::MinusZ;
                idx += 1;
            }
            _ => {}
        }
        let _ = idx;
        res
    }

    /// Squared length implied by the positive connection id, as in Go.
    pub fn distance_squared(self) -> DInt {
        let abs_id = self.pos_id().0;
        if abs_id <= 3 {
            1
        } else if abs_id <= 9 {
            2
        } else if abs_id <= 13 {
            3
        } else if abs_id <= 25 {
            5
        } else {
            panic!("abs Id of connection details {} invalid", self.id);
        }
    }
}

impl fmt::Display for ConnectionDetails {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.id)
    }
}

/// Positive id is ordered before its negative when the absolute ids match.
pub fn is_less_conn_id(cd1: &ConnectionDetails, cd2: &ConnectionDetails) -> bool {
    let abs_diff = i16::from(cd1.pos_id().0) - i16::from(cd2.pos_id().0);
    if abs_diff < 0 {
        true
    } else if abs_diff > 0 {
        false
    } else {
        cd1.id.0 > 0
    }
}

pub fn conn_vector_less(cd1: &ConnectionDetails, cd2: &ConnectionDetails) -> bool {
    use crate::point::abs_dint_from_c;
    let ds_diff = cd1.conn_ds - cd2.conn_ds;
    if ds_diff == 0 {
        for c in 0..3 {
            let d = abs_dint_from_c(cd1.vector.0[c]) - abs_dint_from_c(cd2.vector.0[c]);
            if d != 0 {
                return d > 0;
            }
        }
        for c in 0..3 {
            let d = i64::from(cd1.vector.0[c]) - i64::from(cd2.vector.0[c]);
            if d != 0 {
                return d > 0;
            }
        }
        return false;
    }
    ds_diff < 0
}
