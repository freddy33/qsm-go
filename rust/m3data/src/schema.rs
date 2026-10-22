//! Iceberg schemas and partition specs for the stored QSM tables.
//!
//! Static reference tables (`connection_details`, `trio_details`,
//! `growth_contexts`, `trio_cubes`, `path_builders`) are rebuilt by
//! [`m3point::PointPackData::shared`] and have no files.

use std::sync::Arc;

use iceberg::arrow::schema_to_arrow_schema;
use iceberg::spec::Schema as IcebergSchema;
use iceberg::spec::{NestedField, PrimitiveType, Schema, Transform, Type, UnboundPartitionSpec};

use crate::error::Result;

/// Namespace inside each simulation warehouse.
pub const NAMESPACE: &str = "qsm";

/// Width of the `truncate` transform on `path_nodes.d` and `nodes.creation_time`.
///
/// Growth type 8 is the widest permutation, so a band of 8 distances (or
/// creation times) lines up with one full permutation cycle. Identity on `d`
/// would create one partition per distance and a huge metadata list; a band
/// keeps each simulation step's rows together while still letting a reader
/// skip distant bands.
pub const BUCKET_WIDTH: i32 = 8;

#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub enum TableKind {
    Points,
    PathContexts,
    PathNodes,
    Events,
    Nodes,
}

impl TableKind {
    pub fn all() -> [TableKind; 5] {
        [
            TableKind::Points,
            TableKind::PathContexts,
            TableKind::PathNodes,
            TableKind::Events,
            TableKind::Nodes,
        ]
    }

    pub fn name(self) -> &'static str {
        match self {
            TableKind::Points => "points",
            TableKind::PathContexts => "path_contexts",
            TableKind::PathNodes => "path_nodes",
            TableKind::Events => "events",
            TableKind::Nodes => "nodes",
        }
    }

    pub fn partitioned(self) -> bool {
        matches!(self, TableKind::PathNodes | TableKind::Nodes)
    }

    pub fn iceberg_schema(self) -> Result<IcebergSchema> {
        let fields: Vec<Arc<NestedField>> = match self {
            TableKind::Points => vec![
                req(1, "id", PrimitiveType::Long),
                req(2, "x", PrimitiveType::Int),
                req(3, "y", PrimitiveType::Int),
                req(4, "z", PrimitiveType::Int),
            ],
            TableKind::PathContexts => vec![
                req(1, "id", PrimitiveType::Int),
                req(2, "growth_ctx_id", PrimitiveType::Int),
                req(3, "growth_offset", PrimitiveType::Int),
                opt(4, "path_builders_id", PrimitiveType::Int),
                req(5, "max_dist", PrimitiveType::Int),
            ],
            TableKind::PathNodes => vec![
                req(1, "id", PrimitiveType::Long),
                req(2, "path_ctx_id", PrimitiveType::Int),
                req(3, "path_builders_id", PrimitiveType::Int),
                req(4, "path_builder_idx", PrimitiveType::Int),
                req(5, "trio_id", PrimitiveType::Int),
                req(6, "point_id", PrimitiveType::Long),
                req(7, "d", PrimitiveType::Int),
                req(8, "connection_mask", PrimitiveType::Int),
                opt(9, "path_node1", PrimitiveType::Long),
                opt(10, "path_node2", PrimitiveType::Long),
                opt(11, "path_node3", PrimitiveType::Long),
            ],
            TableKind::Events => vec![
                req(1, "id", PrimitiveType::Int),
                req(2, "path_ctx_id", PrimitiveType::Int),
                req(3, "creation_time", PrimitiveType::Int),
                req(4, "color", PrimitiveType::Int),
                req(5, "end_time", PrimitiveType::Int),
                req(6, "max_node_time", PrimitiveType::Int),
            ],
            TableKind::Nodes => vec![
                req(1, "id", PrimitiveType::Long),
                req(2, "event_id", PrimitiveType::Int),
                req(3, "path_node_id", PrimitiveType::Long),
                req(4, "trio_id", PrimitiveType::Int),
                req(5, "point_id", PrimitiveType::Long),
                req(6, "d", PrimitiveType::Int),
                req(7, "creation_time", PrimitiveType::Int),
                req(8, "connection_mask", PrimitiveType::Int),
                opt(9, "node1", PrimitiveType::Long),
                opt(10, "node2", PrimitiveType::Long),
                opt(11, "node3", PrimitiveType::Long),
            ],
        };
        Ok(Schema::builder()
            .with_fields(fields)
            .with_identifier_field_ids([1])
            .build()?)
    }

    /// Partition spec, when this table is partitioned.
    ///
    /// - `points`, `path_contexts`, and `events` are unpartitioned. Resume loads
    ///   each of them in full. One simulation folder is one space, so `events`
    ///   stays a handful of rows and does not need a partition.
    /// - `path_nodes` is `identity(path_ctx_id)` plus `truncate(d, 8)`.
    /// - `nodes` is `identity(event_id)` plus `truncate(creation_time, 8)`.
    pub fn partition_spec(self) -> Result<Option<UnboundPartitionSpec>> {
        let spec = match self {
            TableKind::PathNodes => Some(
                UnboundPartitionSpec::builder()
                    .add_partition_field(2, "path_ctx_id", Transform::Identity)?
                    .add_partition_field(7, "d_bucket", Transform::Truncate(BUCKET_WIDTH as u32))?
                    .build(),
            ),
            TableKind::Nodes => Some(
                UnboundPartitionSpec::builder()
                    .add_partition_field(2, "event_id", Transform::Identity)?
                    .add_partition_field(
                        7,
                        "creation_time_bucket",
                        Transform::Truncate(BUCKET_WIDTH as u32),
                    )?
                    .build(),
            ),
            TableKind::Points | TableKind::PathContexts | TableKind::Events => None,
        };
        Ok(spec)
    }
}

/// Iceberg `truncate` for a 32-bit integer: `v - (v % W)` with a positive remainder.
pub fn bucket_of(value: i32) -> i32 {
    value - value.rem_euclid(BUCKET_WIDTH)
}

pub fn arrow_schema(kind: TableKind) -> Result<arrow::datatypes::SchemaRef> {
    let schema = kind.iceberg_schema()?;
    Ok(Arc::new(schema_to_arrow_schema(&schema)?))
}

fn req(id: i32, name: &str, ty: PrimitiveType) -> Arc<NestedField> {
    Arc::new(NestedField::required(id, name, Type::Primitive(ty)))
}

fn opt(id: i32, name: &str, ty: PrimitiveType) -> Arc<NestedField> {
    Arc::new(NestedField::optional(id, name, Type::Primitive(ty)))
}
