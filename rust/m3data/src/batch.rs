//! Arrow RecordBatch encode and decode for each stored table.
//!
//! Batches are the unit of a write. A simulation step builds one batch (or one
//! batch per partition) and commits that, rather than one file per row.

use std::sync::Arc;

use arrow::array::{Array, ArrayRef, Int32Array, Int64Array, RecordBatch};

use crate::error::{Error, Result};
use crate::rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow};
use crate::schema::{arrow_schema, TableKind};

pub fn encode_points(rows: &[PointRow]) -> Result<RecordBatch> {
    let schema = arrow_schema(TableKind::Points)?;
    batch(
        schema,
        vec![
            i64_col(rows.iter().map(|r| r.id)),
            i32_col(rows.iter().map(|r| r.x)),
            i32_col(rows.iter().map(|r| r.y)),
            i32_col(rows.iter().map(|r| r.z)),
        ],
    )
}

pub fn decode_points(batches: &[RecordBatch]) -> Result<Vec<PointRow>> {
    let mut rows = Vec::new();
    for batch in batches {
        let id = i64s(batch, "id")?;
        let x = i32s(batch, "x")?;
        let y = i32s(batch, "y")?;
        let z = i32s(batch, "z")?;
        for i in 0..batch.num_rows() {
            rows.push(PointRow {
                id: id.value(i),
                x: x.value(i),
                y: y.value(i),
                z: z.value(i),
            });
        }
    }
    Ok(rows)
}

pub fn encode_path_contexts(rows: &[PathContextRow]) -> Result<RecordBatch> {
    let schema = arrow_schema(TableKind::PathContexts)?;
    batch(
        schema,
        vec![
            i32_col(rows.iter().map(|r| r.id)),
            i32_col(rows.iter().map(|r| r.growth_ctx_id)),
            i32_col(rows.iter().map(|r| r.growth_offset)),
            opt_i32_col(rows.iter().map(|r| r.path_builders_id)),
            i32_col(rows.iter().map(|r| r.max_dist)),
        ],
    )
}

pub fn decode_path_contexts(batches: &[RecordBatch]) -> Result<Vec<PathContextRow>> {
    let mut rows = Vec::new();
    for batch in batches {
        let id = i32s(batch, "id")?;
        let growth_ctx_id = i32s(batch, "growth_ctx_id")?;
        let growth_offset = i32s(batch, "growth_offset")?;
        let path_builders_id = i32s(batch, "path_builders_id")?;
        let max_dist = i32s(batch, "max_dist")?;
        for i in 0..batch.num_rows() {
            rows.push(PathContextRow {
                id: id.value(i),
                growth_ctx_id: growth_ctx_id.value(i),
                growth_offset: growth_offset.value(i),
                path_builders_id: opt_i32(path_builders_id, i),
                max_dist: max_dist.value(i),
            });
        }
    }
    Ok(rows)
}

pub fn encode_path_nodes(rows: &[PathNodeRow]) -> Result<RecordBatch> {
    let schema = arrow_schema(TableKind::PathNodes)?;
    batch(
        schema,
        vec![
            i64_col(rows.iter().map(|r| r.id)),
            i32_col(rows.iter().map(|r| r.path_ctx_id)),
            i32_col(rows.iter().map(|r| r.path_builders_id)),
            i32_col(rows.iter().map(|r| r.path_builder_idx)),
            i32_col(rows.iter().map(|r| r.trio_id)),
            i64_col(rows.iter().map(|r| r.point_id)),
            i32_col(rows.iter().map(|r| r.d)),
            i32_col(rows.iter().map(|r| r.connection_mask)),
            opt_i64_col(rows.iter().map(|r| r.path_node1)),
            opt_i64_col(rows.iter().map(|r| r.path_node2)),
            opt_i64_col(rows.iter().map(|r| r.path_node3)),
        ],
    )
}

pub fn decode_path_nodes(batches: &[RecordBatch]) -> Result<Vec<PathNodeRow>> {
    let mut rows = Vec::new();
    for batch in batches {
        let id = i64s(batch, "id")?;
        let path_ctx_id = i32s(batch, "path_ctx_id")?;
        let path_builders_id = i32s(batch, "path_builders_id")?;
        let path_builder_idx = i32s(batch, "path_builder_idx")?;
        let trio_id = i32s(batch, "trio_id")?;
        let point_id = i64s(batch, "point_id")?;
        let d = i32s(batch, "d")?;
        let connection_mask = i32s(batch, "connection_mask")?;
        let path_node1 = i64s(batch, "path_node1")?;
        let path_node2 = i64s(batch, "path_node2")?;
        let path_node3 = i64s(batch, "path_node3")?;
        for i in 0..batch.num_rows() {
            rows.push(PathNodeRow {
                id: id.value(i),
                path_ctx_id: path_ctx_id.value(i),
                path_builders_id: path_builders_id.value(i),
                path_builder_idx: path_builder_idx.value(i),
                trio_id: trio_id.value(i),
                point_id: point_id.value(i),
                d: d.value(i),
                connection_mask: connection_mask.value(i),
                path_node1: opt_i64(path_node1, i),
                path_node2: opt_i64(path_node2, i),
                path_node3: opt_i64(path_node3, i),
            });
        }
    }
    Ok(rows)
}

pub fn encode_events(rows: &[EventRow]) -> Result<RecordBatch> {
    let schema = arrow_schema(TableKind::Events)?;
    batch(
        schema,
        vec![
            i32_col(rows.iter().map(|r| r.id)),
            i32_col(rows.iter().map(|r| r.path_ctx_id)),
            i32_col(rows.iter().map(|r| r.creation_time)),
            i32_col(rows.iter().map(|r| r.color)),
            i32_col(rows.iter().map(|r| r.end_time)),
            i32_col(rows.iter().map(|r| r.max_node_time)),
        ],
    )
}

pub fn decode_events(batches: &[RecordBatch]) -> Result<Vec<EventRow>> {
    let mut rows = Vec::new();
    for batch in batches {
        let id = i32s(batch, "id")?;
        let path_ctx_id = i32s(batch, "path_ctx_id")?;
        let creation_time = i32s(batch, "creation_time")?;
        let color = i32s(batch, "color")?;
        let end_time = i32s(batch, "end_time")?;
        let max_node_time = i32s(batch, "max_node_time")?;
        for i in 0..batch.num_rows() {
            rows.push(EventRow {
                id: id.value(i),
                path_ctx_id: path_ctx_id.value(i),
                creation_time: creation_time.value(i),
                color: color.value(i),
                end_time: end_time.value(i),
                max_node_time: max_node_time.value(i),
            });
        }
    }
    Ok(rows)
}

pub fn encode_nodes(rows: &[NodeRow]) -> Result<RecordBatch> {
    let schema = arrow_schema(TableKind::Nodes)?;
    batch(
        schema,
        vec![
            i64_col(rows.iter().map(|r| r.id)),
            i32_col(rows.iter().map(|r| r.event_id)),
            i64_col(rows.iter().map(|r| r.path_node_id)),
            i32_col(rows.iter().map(|r| r.trio_id)),
            i64_col(rows.iter().map(|r| r.point_id)),
            i32_col(rows.iter().map(|r| r.d)),
            i32_col(rows.iter().map(|r| r.creation_time)),
            i32_col(rows.iter().map(|r| r.connection_mask)),
            opt_i64_col(rows.iter().map(|r| r.node1)),
            opt_i64_col(rows.iter().map(|r| r.node2)),
            opt_i64_col(rows.iter().map(|r| r.node3)),
        ],
    )
}

pub fn decode_nodes(batches: &[RecordBatch]) -> Result<Vec<NodeRow>> {
    let mut rows = Vec::new();
    for batch in batches {
        let id = i64s(batch, "id")?;
        let event_id = i32s(batch, "event_id")?;
        let path_node_id = i64s(batch, "path_node_id")?;
        let trio_id = i32s(batch, "trio_id")?;
        let point_id = i64s(batch, "point_id")?;
        let d = i32s(batch, "d")?;
        let creation_time = i32s(batch, "creation_time")?;
        let connection_mask = i32s(batch, "connection_mask")?;
        let node1 = i64s(batch, "node1")?;
        let node2 = i64s(batch, "node2")?;
        let node3 = i64s(batch, "node3")?;
        for i in 0..batch.num_rows() {
            rows.push(NodeRow {
                id: id.value(i),
                event_id: event_id.value(i),
                path_node_id: path_node_id.value(i),
                trio_id: trio_id.value(i),
                point_id: point_id.value(i),
                d: d.value(i),
                creation_time: creation_time.value(i),
                connection_mask: connection_mask.value(i),
                node1: opt_i64(node1, i),
                node2: opt_i64(node2, i),
                node3: opt_i64(node3, i),
            });
        }
    }
    Ok(rows)
}

fn batch(schema: arrow::datatypes::SchemaRef, columns: Vec<ArrayRef>) -> Result<RecordBatch> {
    Ok(RecordBatch::try_new(schema, columns)?)
}

fn i32_col(values: impl Iterator<Item = i32>) -> ArrayRef {
    Arc::new(Int32Array::from(values.collect::<Vec<_>>()))
}

fn i64_col(values: impl Iterator<Item = i64>) -> ArrayRef {
    Arc::new(Int64Array::from(values.collect::<Vec<_>>()))
}

fn opt_i32_col(values: impl Iterator<Item = Option<i32>>) -> ArrayRef {
    Arc::new(Int32Array::from(values.collect::<Vec<_>>()))
}

fn opt_i64_col(values: impl Iterator<Item = Option<i64>>) -> ArrayRef {
    Arc::new(Int64Array::from(values.collect::<Vec<_>>()))
}

fn column<'a, T: Array + 'static>(batch: &'a RecordBatch, name: &str) -> Result<&'a T> {
    let array = batch
        .column_by_name(name)
        .ok_or_else(|| Error::new(format!("missing column {name}")))?;
    array.as_any().downcast_ref::<T>().ok_or_else(|| {
        Error::new(format!(
            "column {name} has type {}, expected {}",
            array.data_type(),
            std::any::type_name::<T>()
        ))
    })
}

fn i32s<'a>(batch: &'a RecordBatch, name: &str) -> Result<&'a Int32Array> {
    column(batch, name)
}

fn i64s<'a>(batch: &'a RecordBatch, name: &str) -> Result<&'a Int64Array> {
    column(batch, name)
}

fn opt_i32(array: &Int32Array, index: usize) -> Option<i32> {
    if array.is_null(index) {
        None
    } else {
        Some(array.value(index))
    }
}

fn opt_i64(array: &Int64Array, index: usize) -> Option<i64> {
    if array.is_null(index) {
        None
    } else {
        Some(array.value(index))
    }
}
