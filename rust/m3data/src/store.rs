//! Local Iceberg warehouse for one named simulation.
//!
//! iceberg-rust 0.10 has no filesystem catalog. The catalog index is a
//! [`MemoryCatalog`] whose warehouse is `file://<data>/<simulation>` on the
//! local filesystem. Table metadata JSON is the durable record: opening the
//! same folder scans `qsm/<table>/metadata/*.metadata.json`, picks the highest
//! version, and `register_table`s it. Nothing talks to Postgres or to a network.

use std::collections::{HashMap, HashSet};
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;

use arrow::array::RecordBatch;
use futures::TryStreamExt;
use iceberg::io::LocalFsStorageFactory;
use iceberg::memory::{MemoryCatalog, MemoryCatalogBuilder, MEMORY_CATALOG_WAREHOUSE};
use iceberg::spec::{DataFile, Literal, PartitionKey, Struct};
use iceberg::table::Table;
use iceberg::transaction::{ApplyTransactionAction, Transaction};
use iceberg::writer::base_writer::data_file_writer::DataFileWriterBuilder;
use iceberg::writer::file_writer::location_generator::{
    DefaultFileNameGenerator, DefaultLocationGenerator,
};
use iceberg::writer::file_writer::rolling_writer::RollingFileWriterBuilder;
use iceberg::writer::file_writer::ParquetWriterBuilder;
use iceberg::writer::partitioning::fanout_writer::FanoutWriter;
use iceberg::writer::partitioning::PartitioningWriter;
use iceberg::writer::{IcebergWriter, IcebergWriterBuilder};
use iceberg::{Catalog, CatalogBuilder, NamespaceIdent, TableCreation, TableIdent};
use parquet::file::properties::WriterProperties;

use m3path::{get_connection_state, ConnectionState};

use crate::batch::{
    decode_events, decode_nodes, decode_path_contexts, decode_path_nodes, decode_points,
    encode_events, encode_nodes, encode_path_contexts, encode_path_nodes, encode_points,
};
use crate::error::{Error, Result};
use crate::rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow};
use crate::schema::{bucket_of, TableKind, NAMESPACE};

const MAX_NAME_LEN: usize = 175;

/// In-memory copy of every stored table for one simulation.
///
/// `open` loads each table fully. That is the read API a later viewer uses:
/// pass the simulation name, then borrow the rows. Appends are batches. Each
/// append is one Iceberg snapshot for that table.
pub struct SimulationStore {
    root: PathBuf,
    name: String,
    catalog: MemoryCatalog,
    tables: HashMap<TableKind, Table>,
    points: Vec<PointRow>,
    path_contexts: Vec<PathContextRow>,
    path_nodes: Vec<PathNodeRow>,
    events: Vec<EventRow>,
    nodes: Vec<NodeRow>,
    point_ids: HashSet<i64>,
    point_xyz: HashSet<(i32, i32, i32)>,
    path_ctx_ids: HashSet<i32>,
    path_ctx_key: HashSet<(i32, i32)>,
    path_node_ids: HashSet<i64>,
    path_node_key: HashSet<(i32, i64)>,
    event_ids: HashSet<i32>,
    node_ids: HashSet<i64>,
    node_key: HashSet<(i32, i64)>,
}

impl SimulationStore {
    /// Create `data_root/<simulation>` when it is missing, or reload it.
    pub async fn open(data_root: &Path, simulation: &str) -> Result<Self> {
        validate_name(simulation)?;
        let root = data_root.join(simulation);
        std::fs::create_dir_all(&root)?;
        let root = root.canonicalize()?;

        let catalog = MemoryCatalogBuilder::default()
            .with_storage_factory(Arc::new(LocalFsStorageFactory))
            .load(
                "qsm",
                HashMap::from([(MEMORY_CATALOG_WAREHOUSE.to_string(), file_uri(&root))]),
            )
            .await?;
        let namespace = NamespaceIdent::new(NAMESPACE.to_string());
        catalog.create_namespace(&namespace, HashMap::new()).await?;

        let mut tables = HashMap::new();
        for kind in TableKind::all() {
            let ident = TableIdent::new(namespace.clone(), kind.name().to_string());
            let table = if let Some(metadata) = latest_metadata(&root, kind.name()) {
                catalog.register_table(&ident, file_uri(&metadata)).await?
            } else {
                catalog
                    .create_table(&namespace, table_creation(kind)?)
                    .await?
            };
            tables.insert(kind, table);
        }

        let points = read_points(&tables[&TableKind::Points]).await?;
        // Iceberg 0.10 can only append. A changed path node, path context, or
        // event is appended again with the same id. The latest logical row is
        // the one whose growth cursor moved forward (fewer open connections, or a
        // higher max distance / time). That pick does not depend on scan order.
        let path_contexts =
            collapse_path_contexts(read_path_contexts(&tables[&TableKind::PathContexts]).await?)?;
        let path_nodes =
            collapse_path_nodes(read_path_nodes(&tables[&TableKind::PathNodes]).await?)?;
        let events = collapse_events(read_events(&tables[&TableKind::Events]).await?)?;
        let nodes = read_nodes(&tables[&TableKind::Nodes]).await?;

        let mut store = Self {
            root,
            name: simulation.to_string(),
            catalog,
            tables,
            points: Vec::new(),
            path_contexts: Vec::new(),
            path_nodes: Vec::new(),
            events: Vec::new(),
            nodes: Vec::new(),
            point_ids: HashSet::new(),
            point_xyz: HashSet::new(),
            path_ctx_ids: HashSet::new(),
            path_ctx_key: HashSet::new(),
            path_node_ids: HashSet::new(),
            path_node_key: HashSet::new(),
            event_ids: HashSet::new(),
            node_ids: HashSet::new(),
            node_key: HashSet::new(),
        };
        store.absorb_points(&points)?;
        store.absorb_path_contexts(&path_contexts)?;
        store.absorb_path_nodes(&path_nodes)?;
        store.absorb_events(&events)?;
        store.absorb_nodes(&nodes)?;
        Ok(store)
    }

    /// Absolute `data/<simulation>` directory.
    pub fn root(&self) -> &Path {
        &self.root
    }

    pub fn name(&self) -> &str {
        &self.name
    }

    pub fn points(&self) -> &[PointRow] {
        &self.points
    }

    pub fn path_contexts(&self) -> &[PathContextRow] {
        &self.path_contexts
    }

    pub fn path_nodes(&self) -> &[PathNodeRow] {
        &self.path_nodes
    }

    pub fn events(&self) -> &[EventRow] {
        &self.events
    }

    pub fn nodes(&self) -> &[NodeRow] {
        &self.nodes
    }

    /// True when any stored table has a row. Empty Iceberg metadata from `open` does not count.
    pub fn has_rows(&self) -> bool {
        !self.points.is_empty()
            || !self.path_contexts.is_empty()
            || !self.path_nodes.is_empty()
            || !self.events.is_empty()
            || !self.nodes.is_empty()
    }

    pub async fn append_points(&mut self, rows: &[PointRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.check_point_rows(rows)?;
        let batch = encode_points(rows)?;
        self.write(TableKind::Points, vec![(None, batch)]).await?;
        self.absorb_points(rows)
    }

    pub async fn append_path_contexts(&mut self, rows: &[PathContextRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.check_path_context_rows(rows)?;
        let batch = encode_path_contexts(rows)?;
        self.write(TableKind::PathContexts, vec![(None, batch)])
            .await?;
        self.absorb_path_contexts(rows)
    }

    pub async fn append_path_nodes(&mut self, rows: &[PathNodeRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.check_path_node_rows(rows)?;
        let mut groups: HashMap<(i32, i32), Vec<PathNodeRow>> = HashMap::new();
        for row in rows {
            groups
                .entry((row.path_ctx_id, bucket_of(row.d)))
                .or_default()
                .push(row.clone());
        }
        let table = self.tables[&TableKind::PathNodes].clone();
        let mut batches = Vec::new();
        for ((path_ctx_id, d_bucket), group) in groups {
            let key = partition_key(
                &table,
                vec![
                    Some(Literal::int(path_ctx_id)),
                    Some(Literal::int(d_bucket)),
                ],
            );
            batches.push((Some(key), encode_path_nodes(&group)?));
        }
        self.write(TableKind::PathNodes, batches).await?;
        self.absorb_path_nodes(rows)
    }

    pub async fn append_events(&mut self, rows: &[EventRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.check_event_rows(rows)?;
        let batch = encode_events(rows)?;
        self.write(TableKind::Events, vec![(None, batch)]).await?;
        self.absorb_events(rows)
    }

    pub async fn append_nodes(&mut self, rows: &[NodeRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.check_node_rows(rows)?;
        let mut groups: HashMap<(i32, i32), Vec<NodeRow>> = HashMap::new();
        for row in rows {
            groups
                .entry((row.event_id, bucket_of(row.creation_time)))
                .or_default()
                .push(row.clone());
        }
        let table = self.tables[&TableKind::Nodes].clone();
        let mut batches = Vec::new();
        for ((event_id, time_bucket), group) in groups {
            let key = partition_key(
                &table,
                vec![
                    Some(Literal::int(event_id)),
                    Some(Literal::int(time_bucket)),
                ],
            );
            batches.push((Some(key), encode_nodes(&group)?));
        }
        self.write(TableKind::Nodes, batches).await?;
        self.absorb_nodes(rows)
    }

    async fn write(
        &mut self,
        kind: TableKind,
        batches: Vec<(Option<PartitionKey>, RecordBatch)>,
    ) -> Result<()> {
        let table = self.tables[&kind].clone();
        let files = write_batches(&table, kind.partitioned(), batches).await?;
        let tx = Transaction::new(&table);
        let tx = tx.fast_append().add_data_files(files).apply(tx)?;
        let updated = tx.commit(&self.catalog).await?;
        self.tables.insert(kind, updated);
        Ok(())
    }

    fn check_point_rows(&self, rows: &[PointRow]) -> Result<()> {
        let mut ids = self.point_ids.clone();
        let mut xyz = self.point_xyz.clone();
        for row in rows {
            insert_id(&mut ids, row.id, "points")?;
            if !xyz.insert(row.xyz()) {
                return Err(Error::new(format!(
                    "duplicate point ({}, {}, {})",
                    row.x, row.y, row.z
                )));
            }
        }
        Ok(())
    }

    fn check_path_context_rows(&self, rows: &[PathContextRow]) -> Result<()> {
        let mut new_ids = HashSet::new();
        let mut new_keys = HashSet::new();
        for row in rows {
            if row.id <= 0 {
                return Err(Error::new(format!(
                    "path_contexts id must be positive, got {}",
                    row.id
                )));
            }
            if let Some(existing) = self.path_contexts.iter().find(|ctx| ctx.id == row.id) {
                if existing.growth_ctx_id != row.growth_ctx_id
                    || existing.growth_offset != row.growth_offset
                {
                    return Err(Error::new(format!(
                        "path context {} changed (growth_ctx_id, growth_offset)",
                        row.id
                    )));
                }
                continue;
            }
            if self.path_ctx_ids.contains(&row.id) || !new_ids.insert(row.id) {
                return Err(Error::new(format!("duplicate path_contexts id {}", row.id)));
            }
            let key = (row.growth_ctx_id, row.growth_offset);
            if self.path_ctx_key.contains(&key) || !new_keys.insert(key) {
                return Err(Error::new(format!(
                    "duplicate path context ({}, {})",
                    row.growth_ctx_id, row.growth_offset
                )));
            }
        }
        Ok(())
    }

    fn check_path_node_rows(&self, rows: &[PathNodeRow]) -> Result<()> {
        let mut new_ids = HashSet::new();
        let mut new_keys = HashSet::new();
        for row in rows {
            if row.id <= 0 {
                return Err(Error::new(format!(
                    "path_nodes id must be positive, got {}",
                    row.id
                )));
            }
            if let Some(existing) = self.path_nodes.iter().find(|node| node.id == row.id) {
                if existing.path_ctx_id != row.path_ctx_id || existing.point_id != row.point_id {
                    return Err(Error::new(format!(
                        "path node {} changed (path_ctx_id, point_id)",
                        row.id
                    )));
                }
                continue;
            }
            if self.path_node_ids.contains(&row.id) || !new_ids.insert(row.id) {
                return Err(Error::new(format!("duplicate path_nodes id {}", row.id)));
            }
            let key = (row.path_ctx_id, row.point_id);
            if self.path_node_key.contains(&key) || !new_keys.insert(key) {
                return Err(Error::new(format!(
                    "duplicate path node ({}, {})",
                    row.path_ctx_id, row.point_id
                )));
            }
        }
        Ok(())
    }

    fn check_event_rows(&self, rows: &[EventRow]) -> Result<()> {
        let mut new_ids = HashSet::new();
        for row in rows {
            if row.id <= 0 {
                return Err(Error::new(format!(
                    "events id must be positive, got {}",
                    row.id
                )));
            }
            if self.events.iter().any(|event| event.id == row.id) {
                continue;
            }
            if self.event_ids.contains(&row.id) || !new_ids.insert(row.id) {
                return Err(Error::new(format!("duplicate events id {}", row.id)));
            }
        }
        Ok(())
    }

    fn check_node_rows(&self, rows: &[NodeRow]) -> Result<()> {
        let mut ids = self.node_ids.clone();
        let mut keys = self.node_key.clone();
        for row in rows {
            insert_id(&mut ids, row.id, "nodes")?;
            if !keys.insert((row.event_id, row.point_id)) {
                return Err(Error::new(format!(
                    "duplicate node ({}, {})",
                    row.event_id, row.point_id
                )));
            }
        }
        Ok(())
    }

    fn absorb_points(&mut self, rows: &[PointRow]) -> Result<()> {
        self.check_point_rows(rows)?;
        for row in rows {
            self.point_ids.insert(row.id);
            self.point_xyz.insert(row.xyz());
            self.points.push(row.clone());
        }
        Ok(())
    }

    fn absorb_path_contexts(&mut self, rows: &[PathContextRow]) -> Result<()> {
        self.check_path_context_rows(rows)?;
        for row in rows {
            if let Some(pos) = self.path_contexts.iter().position(|ctx| ctx.id == row.id) {
                self.path_contexts[pos] = row.clone();
            } else {
                self.path_ctx_ids.insert(row.id);
                self.path_ctx_key
                    .insert((row.growth_ctx_id, row.growth_offset));
                self.path_contexts.push(row.clone());
            }
        }
        Ok(())
    }

    fn absorb_path_nodes(&mut self, rows: &[PathNodeRow]) -> Result<()> {
        self.check_path_node_rows(rows)?;
        for row in rows {
            if let Some(pos) = self.path_nodes.iter().position(|node| node.id == row.id) {
                self.path_nodes[pos] = row.clone();
            } else {
                self.path_node_ids.insert(row.id);
                self.path_node_key.insert((row.path_ctx_id, row.point_id));
                self.path_nodes.push(row.clone());
            }
        }
        Ok(())
    }

    fn absorb_events(&mut self, rows: &[EventRow]) -> Result<()> {
        self.check_event_rows(rows)?;
        for row in rows {
            if let Some(pos) = self.events.iter().position(|event| event.id == row.id) {
                self.events[pos] = row.clone();
            } else {
                self.event_ids.insert(row.id);
                self.events.push(row.clone());
            }
        }
        Ok(())
    }

    fn absorb_nodes(&mut self, rows: &[NodeRow]) -> Result<()> {
        self.check_node_rows(rows)?;
        for row in rows {
            self.node_ids.insert(row.id);
            self.node_key.insert((row.event_id, row.point_id));
            self.nodes.push(row.clone());
        }
        Ok(())
    }
}

fn insert_id(ids: &mut HashSet<i64>, id: i64, table: &str) -> Result<()> {
    if id <= 0 {
        return Err(Error::new(format!("{table} id must be positive, got {id}")));
    }
    if !ids.insert(id) {
        return Err(Error::new(format!("duplicate {table} id {id}")));
    }
    Ok(())
}

fn validate_name(name: &str) -> Result<()> {
    if name.is_empty() || name.len() > MAX_NAME_LEN {
        return Err(Error::new(
            "simulation name must be 1..=175 characters".to_string(),
        ));
    }
    if name == "."
        || name == ".."
        || name.contains('/')
        || name.contains('\\')
        || name.contains('\0')
    {
        return Err(Error::new(format!(
            "simulation name {name:?} is not a single folder name"
        )));
    }
    Ok(())
}

fn file_uri(path: &Path) -> String {
    format!("file://{}", path.display())
}

fn unique_file_suffix() -> String {
    static SEQ: AtomicU64 = AtomicU64::new(0);
    format!(
        "{}-{}",
        std::process::id(),
        SEQ.fetch_add(1, Ordering::Relaxed)
    )
}

fn latest_metadata(root: &Path, table: &str) -> Option<PathBuf> {
    let dir = root.join(NAMESPACE).join(table).join("metadata");
    let mut files: Vec<PathBuf> = std::fs::read_dir(&dir)
        .ok()?
        .flatten()
        .map(|e| e.path())
        .filter(|path| {
            path.file_name()
                .and_then(|name| name.to_str())
                .is_some_and(|name| name.ends_with(".metadata.json"))
        })
        .collect();
    files.sort();
    files.pop()
}

fn table_creation(kind: TableKind) -> Result<TableCreation> {
    let schema = kind.iceberg_schema()?;
    let name = kind.name().to_string();
    if let Some(spec) = kind.partition_spec()? {
        Ok(TableCreation::builder()
            .name(name)
            .schema(schema)
            .partition_spec(spec)
            .build())
    } else {
        Ok(TableCreation::builder().name(name).schema(schema).build())
    }
}

fn partition_key(table: &Table, values: Vec<Option<Literal>>) -> PartitionKey {
    PartitionKey::new(
        table.metadata().default_partition_spec().as_ref().clone(),
        table.metadata().current_schema().clone(),
        Struct::from_iter(values),
    )
}

async fn write_batches(
    table: &Table,
    partitioned: bool,
    batches: Vec<(Option<PartitionKey>, RecordBatch)>,
) -> Result<Vec<DataFile>> {
    let location = DefaultLocationGenerator::new(table.metadata())?;
    // The generator's counter starts at zero on every writer. A stable prefix
    // would reuse `part-00000.parquet` and the next commit would see that path
    // as already referenced. The suffix makes each batch's files unique.
    let names = DefaultFileNameGenerator::new(
        "part".to_string(),
        Some(unique_file_suffix()),
        iceberg::spec::DataFileFormat::Parquet,
    );
    let parquet = ParquetWriterBuilder::new(
        WriterProperties::default(),
        table.metadata().current_schema().clone(),
    );
    let rolling = RollingFileWriterBuilder::new_with_default_file_size(
        parquet,
        table.file_io().clone(),
        location,
        names,
    );
    let builder = DataFileWriterBuilder::new(rolling);
    if partitioned {
        let mut fanout = FanoutWriter::new(builder);
        for (key, batch) in batches {
            let key = key.ok_or_else(|| Error::new("partitioned write is missing a key"))?;
            fanout.write(key, batch).await?;
        }
        Ok(PartitioningWriter::close(fanout).await?)
    } else {
        let mut writer = builder.build(None).await?;
        for (_, batch) in batches {
            writer.write(batch).await?;
        }
        Ok(writer.close().await?)
    }
}

fn collapse_path_contexts(rows: Vec<PathContextRow>) -> Result<Vec<PathContextRow>> {
    collapse(
        rows,
        |row| row.id as i64,
        |row| row.max_dist,
        |prev, row| {
            if prev.growth_ctx_id != row.growth_ctx_id || prev.growth_offset != row.growth_offset {
                return Err(Error::new(format!(
                    "path context {} was appended with a different growth key",
                    row.id
                )));
            }
            Ok(())
        },
    )
}

fn collapse_path_nodes(rows: Vec<PathNodeRow>) -> Result<Vec<PathNodeRow>> {
    collapse(
        rows,
        |row| row.id,
        path_node_rank,
        |prev, row| {
            if prev.path_ctx_id != row.path_ctx_id || prev.point_id != row.point_id {
                return Err(Error::new(format!(
                    "path node {} was appended with a different (path_ctx_id, point_id)",
                    row.id
                )));
            }
            Ok(())
        },
    )
}

fn path_node_rank(row: &PathNodeRow) -> i32 {
    let open = (0..3)
        .filter(|&conn| {
            get_connection_state(row.connection_mask as u16, conn) == ConnectionState::NotSet
        })
        .count() as i32;
    let links = [row.path_node1, row.path_node2, row.path_node3]
        .iter()
        .filter(|link| link.is_some())
        .count() as i32;
    // Fewer open connections, then more assigned neighbour ids.
    -open * 10 + links
}

fn collapse_events(rows: Vec<EventRow>) -> Result<Vec<EventRow>> {
    collapse(
        rows,
        |row| row.id as i64,
        |row| row.max_node_time,
        |prev, row| {
            if prev.path_ctx_id != row.path_ctx_id
                || prev.creation_time != row.creation_time
                || prev.color != row.color
            {
                return Err(Error::new(format!(
                    "event {} was appended with different identity columns",
                    row.id
                )));
            }
            Ok(())
        },
    )
}

fn collapse<T>(
    rows: Vec<T>,
    id_of: impl Fn(&T) -> i64,
    rank: impl Fn(&T) -> i32,
    same: impl Fn(&T, &T) -> Result<()>,
) -> Result<Vec<T>> {
    let mut order = Vec::new();
    let mut best: HashMap<i64, T> = HashMap::new();
    let mut best_rank: HashMap<i64, i32> = HashMap::new();
    for row in rows {
        let id = id_of(&row);
        if let Some(prev) = best.get(&id) {
            same(prev, &row)?;
            if rank(&row) > best_rank[&id] {
                best_rank.insert(id, rank(&row));
                best.insert(id, row);
            }
        } else {
            order.push(id);
            best_rank.insert(id, rank(&row));
            best.insert(id, row);
        }
    }
    Ok(order
        .into_iter()
        .map(|id| best.remove(&id).expect("id was inserted"))
        .collect())
}

async fn read_batches(table: &Table) -> Result<Vec<RecordBatch>> {
    if table.metadata().current_snapshot().is_none() {
        return Ok(Vec::new());
    }
    let stream = table.scan().build()?.to_arrow().await?;
    Ok(stream.try_collect().await?)
}

async fn read_points(table: &Table) -> Result<Vec<PointRow>> {
    decode_points(&read_batches(table).await?)
}

async fn read_path_contexts(table: &Table) -> Result<Vec<PathContextRow>> {
    decode_path_contexts(&read_batches(table).await?)
}

async fn read_path_nodes(table: &Table) -> Result<Vec<PathNodeRow>> {
    decode_path_nodes(&read_batches(table).await?)
}

async fn read_events(table: &Table) -> Result<Vec<EventRow>> {
    decode_events(&read_batches(table).await?)
}

async fn read_nodes(table: &Table) -> Result<Vec<NodeRow>> {
    decode_nodes(&read_batches(table).await?)
}
