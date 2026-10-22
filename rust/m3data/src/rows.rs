//! Row types for the five tables stored under a simulation folder.
//!
//! One simulation folder is one space. Space parameters live in `inputs.json`,
//! not in a table.
//!
//! Column names and nullability follow `docs/db/README.md`. Postgres `smallint`
//! and `integer` are `i32`. `bigserial` / `bigint` ids are `i64`. Iceberg has
//! no 16-bit integer.

use m3point::Point;

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PointRow {
    pub id: i64,
    pub x: i32,
    pub y: i32,
    pub z: i32,
}

impl PointRow {
    pub fn point(&self) -> Point {
        Point([self.x, self.y, self.z])
    }

    pub fn xyz(&self) -> (i32, i32, i32) {
        (self.x, self.y, self.z)
    }
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PathContextRow {
    pub id: i32,
    pub growth_ctx_id: i32,
    pub growth_offset: i32,
    pub path_builders_id: Option<i32>,
    pub max_dist: i32,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PathNodeRow {
    pub id: i64,
    pub path_ctx_id: i32,
    pub path_builders_id: i32,
    pub path_builder_idx: i32,
    pub trio_id: i32,
    pub point_id: i64,
    pub d: i32,
    pub connection_mask: i32,
    pub path_node1: Option<i64>,
    pub path_node2: Option<i64>,
    pub path_node3: Option<i64>,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct EventRow {
    pub id: i32,
    pub path_ctx_id: i32,
    pub creation_time: i32,
    pub color: i32,
    pub end_time: i32,
    pub max_node_time: i32,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct NodeRow {
    pub id: i64,
    pub event_id: i32,
    pub path_node_id: i64,
    pub trio_id: i32,
    pub point_id: i64,
    pub d: i32,
    pub creation_time: i32,
    pub connection_mask: i32,
    pub node1: Option<i64>,
    pub node2: Option<i64>,
    pub node3: Option<i64>,
}
