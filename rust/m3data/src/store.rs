//! Local Iceberg warehouse for one named simulation.
//!
//! iceberg-rust 0.10 has no filesystem catalog and no public overwrite action
//! (`TableCommit`'s builder and `TransactionAction` are crate-private, and
//! `fast_append` always keeps the previous data files). Parquet itself cannot
//! be appended in place. Each commit therefore:
//!
//! 1. rewrites the table's rows into one new Parquet file (the rolling writer
//!    target is `usize::MAX`, so it does not split);
//! 2. writes a manifest and a manifest list that name only that file;
//! 3. builds the next table metadata with an `overwrite` snapshot and
//!    `SetSnapshotRef` on `main`, and writes that metadata JSON;
//! 4. points the memory catalog at the new metadata.
//!
//! The metadata file is the per-table commit. `qsm/step.json` is written only
//! after every dirty table in the step has been committed. A crash before that
//! journal leaves the new snapshot unreferenced by the journal; the next open
//! moves `main` back to the journaled snapshot (or clears it) and then deletes
//! Parquet files the restored snapshot does not name. Old snapshots stay in
//! the metadata so that rollback can still see them. Their Parquet is removed
//! only after the journal names the new snapshot.

use std::collections::{HashMap, HashSet};
use std::path::{Path, PathBuf};
use std::str::FromStr;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};

use arrow::array::RecordBatch;
use futures::TryStreamExt;
use iceberg::io::LocalFsStorageFactory;
use iceberg::memory::{MemoryCatalog, MemoryCatalogBuilder, MEMORY_CATALOG_WAREHOUSE};
use iceberg::spec::{
    DataFile, FormatVersion, ManifestListWriter, ManifestWriterBuilder, Operation, Snapshot,
    SnapshotReference, SnapshotRetention, Summary, MAIN_BRANCH,
};
use iceberg::table::Table;
use iceberg::writer::base_writer::data_file_writer::DataFileWriterBuilder;
use iceberg::writer::file_writer::location_generator::{
    DefaultFileNameGenerator, DefaultLocationGenerator,
};
use iceberg::writer::file_writer::rolling_writer::RollingFileWriterBuilder;
use iceberg::writer::file_writer::ParquetWriterBuilder;
use iceberg::writer::{IcebergWriter, IcebergWriterBuilder};
use iceberg::{
    Catalog, CatalogBuilder, MetadataLocation, NamespaceIdent, TableCreation, TableIdent,
    TableUpdate,
};
use parquet::file::properties::WriterProperties;
use serde::{Deserialize, Serialize};

use crate::batch::{
    decode_events, decode_nodes, decode_path_contexts, decode_path_nodes, decode_points,
    encode_events, encode_nodes, encode_path_contexts, encode_path_nodes, encode_points,
};
use crate::error::{Error, Result};
use crate::rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow};
use crate::schema::{TableKind, NAMESPACE};

const MAX_NAME_LEN: usize = 175;
const STEP_JOURNAL: &str = "step.json";

/// In-memory copy of every stored table for one simulation.
///
/// `open` loads each table fully. Appends stage rows and [`SimulationStore::sync`]
/// rewrites each dirty table as one data file. `append_*` stages and syncs, so a
/// direct append is its own commit.
pub struct SimulationStore {
    root: PathBuf,
    name: String,
    catalog: MemoryCatalog,
    tables: HashMap<TableKind, Table>,
    dirty: HashSet<TableKind>,
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
            let ident = table_ident(kind);
            let table = if let Some(metadata) = latest_metadata(&root, kind.name()) {
                catalog.register_table(&ident, file_uri(&metadata)).await?
            } else {
                catalog
                    .create_table(&namespace, table_creation(kind)?)
                    .await?
            };
            tables.insert(kind, table);
        }

        let mut store = Self {
            root,
            name: simulation.to_string(),
            catalog,
            tables,
            dirty: HashSet::new(),
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
        store.reconcile_journal().await?;
        store.gc_unreferenced().await?;

        let points = read_points(&store.tables[&TableKind::Points]).await?;
        let path_contexts = read_path_contexts(&store.tables[&TableKind::PathContexts]).await?;
        let path_nodes = read_path_nodes(&store.tables[&TableKind::PathNodes]).await?;
        let events = read_events(&store.tables[&TableKind::Events]).await?;
        let nodes = read_nodes(&store.tables[&TableKind::Nodes]).await?;
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

    pub(crate) fn stage_points(&mut self, rows: &[PointRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.absorb_points(rows)?;
        self.dirty.insert(TableKind::Points);
        Ok(())
    }

    pub(crate) fn stage_path_contexts(&mut self, rows: &[PathContextRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.absorb_path_contexts(rows)?;
        self.dirty.insert(TableKind::PathContexts);
        Ok(())
    }

    pub(crate) fn stage_path_nodes(&mut self, rows: &[PathNodeRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.absorb_path_nodes(rows)?;
        self.dirty.insert(TableKind::PathNodes);
        Ok(())
    }

    /// Replace link columns of path nodes that are already stored.
    ///
    /// The id and `(path_ctx_id, point_id)` stay the same. The single data file
    /// is rewritten on the next [`SimulationStore::sync`], so the id still
    /// appears once. This is the one mutation growth cannot avoid: the next
    /// shell fills `Next` and `Blocked` after the node has been saved.
    pub(crate) fn stage_path_node_revisions(&mut self, rows: &[PathNodeRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        for row in rows {
            let Some(existing) = self.path_nodes.iter().find(|node| node.id == row.id) else {
                return Err(Error::new(format!(
                    "path node {} is not stored and cannot be revised",
                    row.id
                )));
            };
            if existing.path_ctx_id != row.path_ctx_id || existing.point_id != row.point_id {
                return Err(Error::new(format!(
                    "path node {} changed (path_ctx_id, point_id)",
                    row.id
                )));
            }
        }
        for row in rows {
            let pos = self
                .path_nodes
                .iter()
                .position(|node| node.id == row.id)
                .expect("revision target was checked");
            self.path_nodes[pos] = row.clone();
        }
        self.dirty.insert(TableKind::PathNodes);
        Ok(())
    }

    pub(crate) fn stage_events(&mut self, rows: &[EventRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.absorb_events(rows)?;
        self.dirty.insert(TableKind::Events);
        Ok(())
    }

    pub(crate) fn stage_nodes(&mut self, rows: &[NodeRow]) -> Result<()> {
        if rows.is_empty() {
            return Ok(());
        }
        self.absorb_nodes(rows)?;
        self.dirty.insert(TableKind::Nodes);
        Ok(())
    }

    /// Rewrite every dirty table, then record the snapshot ids in `qsm/step.json`.
    pub(crate) async fn sync(&mut self) -> Result<()> {
        self.commit_dirty().await?;
        if self.journal_matches()? {
            return Ok(());
        }
        self.write_journal()?;
        self.gc_unreferenced().await
    }

    /// Commit staged rows without moving the step journal.
    ///
    /// The next `open` treats that as a crash and restores the journaled snapshot.
    #[cfg(test)]
    pub(crate) async fn commit_without_journal_for_test(&mut self) -> Result<()> {
        self.commit_dirty().await
    }

    pub async fn append_points(&mut self, rows: &[PointRow]) -> Result<()> {
        self.stage_points(rows)?;
        self.sync().await
    }

    pub async fn append_path_contexts(&mut self, rows: &[PathContextRow]) -> Result<()> {
        self.stage_path_contexts(rows)?;
        self.sync().await
    }

    pub async fn append_path_nodes(&mut self, rows: &[PathNodeRow]) -> Result<()> {
        self.stage_path_nodes(rows)?;
        self.sync().await
    }

    pub async fn append_events(&mut self, rows: &[EventRow]) -> Result<()> {
        self.stage_events(rows)?;
        self.sync().await
    }

    pub async fn append_nodes(&mut self, rows: &[NodeRow]) -> Result<()> {
        self.stage_nodes(rows)?;
        self.sync().await
    }

    async fn commit_dirty(&mut self) -> Result<()> {
        for kind in TableKind::all() {
            if !self.dirty.contains(&kind) {
                continue;
            }
            self.rewrite(kind).await?;
            self.dirty.remove(&kind);
        }
        Ok(())
    }

    async fn rewrite(&mut self, kind: TableKind) -> Result<()> {
        let batch = match kind {
            TableKind::Points => encode_points(&self.points)?,
            TableKind::PathContexts => encode_path_contexts(&self.path_contexts)?,
            TableKind::PathNodes => encode_path_nodes(&self.path_nodes)?,
            TableKind::Events => encode_events(&self.events)?,
            TableKind::Nodes => encode_nodes(&self.nodes)?,
        };
        if batch.num_rows() == 0 {
            return Err(Error::new(format!("{} rewrite has no rows", kind.name())));
        }
        let table = self.tables[&kind].clone();
        let files = write_one_file(&table, batch).await?;
        let updated = commit_overwrite(&self.catalog, &table, files).await?;
        self.tables.insert(kind, updated);
        Ok(())
    }

    fn journal_matches(&self) -> Result<bool> {
        let Some(journal) = read_journal(&self.root)? else {
            return Ok(false);
        };
        Ok(journal == self.current_journal())
    }

    fn current_journal(&self) -> StepJournal {
        StepJournal {
            points: snapshot_id(&self.tables[&TableKind::Points]),
            path_contexts: snapshot_id(&self.tables[&TableKind::PathContexts]),
            path_nodes: snapshot_id(&self.tables[&TableKind::PathNodes]),
            events: snapshot_id(&self.tables[&TableKind::Events]),
            nodes: snapshot_id(&self.tables[&TableKind::Nodes]),
        }
    }

    fn write_journal(&self) -> Result<()> {
        let qsm = self.root.join(NAMESPACE);
        std::fs::create_dir_all(&qsm)?;
        let tmp = qsm.join("step.json.tmp");
        let dest = qsm.join(STEP_JOURNAL);
        std::fs::write(&tmp, serde_json::to_vec_pretty(&self.current_journal())?)?;
        std::fs::rename(&tmp, &dest)?;
        Ok(())
    }

    async fn reconcile_journal(&mut self) -> Result<()> {
        let journal = read_journal(&self.root)?;
        for kind in TableKind::all() {
            let have = snapshot_id(&self.tables[&kind]);
            let want = journal.as_ref().and_then(|journal| journal.of(kind));
            // No journal and no snapshot: a fresh folder. No journal and a
            // snapshot: the first step crashed before `step.json`.
            if journal.is_none() && have.is_none() {
                continue;
            }
            if have == want {
                continue;
            }
            self.roll_snapshot(kind, want).await?;
        }
        Ok(())
    }

    async fn roll_snapshot(&mut self, kind: TableKind, snapshot_id: Option<i64>) -> Result<()> {
        let table = self.tables[&kind].clone();
        let update = match snapshot_id {
            Some(id) => {
                if table.metadata().snapshot_by_id(id).is_none() {
                    return Err(Error::new(format!(
                        "{} step journal names missing snapshot {id}",
                        kind.name()
                    )));
                }
                TableUpdate::SetSnapshotRef {
                    ref_name: MAIN_BRANCH.to_string(),
                    reference: branch_ref(id),
                }
            }
            None => TableUpdate::RemoveSnapshotRef {
                ref_name: MAIN_BRANCH.to_string(),
            },
        };
        let updated = publish_metadata(&self.catalog, &table, vec![update]).await?;
        self.tables.insert(kind, updated);
        Ok(())
    }

    async fn gc_unreferenced(&self) -> Result<()> {
        for kind in TableKind::all() {
            let keep = referenced_parquet(&self.tables[&kind]).await?;
            delete_unreferenced(&self.root, kind.name(), &keep)?;
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
            self.path_ctx_ids.insert(row.id);
            self.path_ctx_key
                .insert((row.growth_ctx_id, row.growth_offset));
            self.path_contexts.push(row.clone());
        }
        Ok(())
    }

    fn absorb_path_nodes(&mut self, rows: &[PathNodeRow]) -> Result<()> {
        self.check_path_node_rows(rows)?;
        for row in rows {
            self.path_node_ids.insert(row.id);
            self.path_node_key.insert((row.path_ctx_id, row.point_id));
            self.path_nodes.push(row.clone());
        }
        Ok(())
    }

    fn absorb_events(&mut self, rows: &[EventRow]) -> Result<()> {
        self.check_event_rows(rows)?;
        for row in rows {
            self.event_ids.insert(row.id);
            self.events.push(row.clone());
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
        let mut ids = self.path_ctx_ids.clone();
        let mut keys = self.path_ctx_key.clone();
        for row in rows {
            insert_i32(&mut ids, row.id, "path_contexts")?;
            if !keys.insert((row.growth_ctx_id, row.growth_offset)) {
                return Err(Error::new(format!(
                    "duplicate path context ({}, {})",
                    row.growth_ctx_id, row.growth_offset
                )));
            }
        }
        Ok(())
    }

    fn check_path_node_rows(&self, rows: &[PathNodeRow]) -> Result<()> {
        let mut ids = self.path_node_ids.clone();
        let mut keys = self.path_node_key.clone();
        for row in rows {
            insert_id(&mut ids, row.id, "path_nodes")?;
            if !keys.insert((row.path_ctx_id, row.point_id)) {
                return Err(Error::new(format!(
                    "duplicate path node ({}, {})",
                    row.path_ctx_id, row.point_id
                )));
            }
        }
        Ok(())
    }

    fn check_event_rows(&self, rows: &[EventRow]) -> Result<()> {
        let mut ids = self.event_ids.clone();
        for row in rows {
            insert_i32(&mut ids, row.id, "events")?;
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
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
struct StepJournal {
    points: Option<i64>,
    path_contexts: Option<i64>,
    path_nodes: Option<i64>,
    events: Option<i64>,
    nodes: Option<i64>,
}

impl StepJournal {
    fn of(&self, kind: TableKind) -> Option<i64> {
        match kind {
            TableKind::Points => self.points,
            TableKind::PathContexts => self.path_contexts,
            TableKind::PathNodes => self.path_nodes,
            TableKind::Events => self.events,
            TableKind::Nodes => self.nodes,
        }
    }
}

fn snapshot_id(table: &Table) -> Option<i64> {
    table.metadata().current_snapshot_id()
}

fn branch_ref(snapshot_id: i64) -> SnapshotReference {
    SnapshotReference::new(snapshot_id, SnapshotRetention::branch(None, None, None))
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

fn insert_i32(ids: &mut HashSet<i32>, id: i32, table: &str) -> Result<()> {
    if id <= 0 {
        return Err(Error::new(format!("{table} id must be positive, got {id}")));
    }
    if !ids.insert(id) {
        return Err(Error::new(format!("duplicate {table} id {id}")));
    }
    Ok(())
}

pub fn validate_name(name: &str) -> Result<()> {
    if name.is_empty() || name.len() > MAX_NAME_LEN {
        return Err(Error::new(
            "simulation name must be 1..=175 characters".to_string(),
        ));
    }
    if name == "."
        || name == ".."
        || name.chars().any(is_forbidden_in_file_name)
        || is_windows_reserved(name)
    {
        return Err(Error::new(format!(
            "simulation name {name:?} is not a single folder name"
        )));
    }
    Ok(())
}

/// Characters Windows rejects in a file name, plus `/` and NUL.
fn is_forbidden_in_file_name(c: char) -> bool {
    matches!(
        c,
        '/' | '\\' | ':' | '*' | '?' | '"' | '<' | '>' | '|' | '\0'
    )
}

/// `CON`, `NUL`, `COM1`, and the other device names Windows will not create.
fn is_windows_reserved(name: &str) -> bool {
    let base = name.split('.').next().unwrap_or(name);
    matches!(
        base.to_ascii_uppercase().as_str(),
        "CON"
            | "PRN"
            | "AUX"
            | "NUL"
            | "COM1"
            | "COM2"
            | "COM3"
            | "COM4"
            | "COM5"
            | "COM6"
            | "COM7"
            | "COM8"
            | "COM9"
            | "LPT1"
            | "LPT2"
            | "LPT3"
            | "LPT4"
            | "LPT5"
            | "LPT6"
            | "LPT7"
            | "LPT8"
            | "LPT9"
    )
}

/// `file://` URI for a local path.
///
/// Unix absolute paths stay `file:///tmp/...`. A Windows path, including the
/// `\\?\` prefix from `canonicalize`, becomes `file:///C:/...` with forward
/// slashes. The drive-letter colon is the only colon in that URI.
fn file_uri(path: &Path) -> String {
    let text = path.to_string_lossy();
    let text = text.strip_prefix(r"\\?\").unwrap_or(&text);
    let mut text = text.replace('\\', "/");
    if !text.starts_with('/') {
        text.insert(0, '/');
    }
    format!("file://{text}")
}

fn unique_file_suffix() -> String {
    static SEQ: AtomicU64 = AtomicU64::new(0);
    format!(
        "{}-{}",
        std::process::id(),
        SEQ.fetch_add(1, Ordering::Relaxed)
    )
}

fn now_millis() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|elapsed| elapsed.as_millis() as i64)
        .unwrap_or(0)
}

fn latest_metadata(root: &Path, table: &str) -> Option<PathBuf> {
    let dir = root.join(NAMESPACE).join(table).join("metadata");
    let mut files: Vec<PathBuf> = std::fs::read_dir(&dir)
        .ok()?
        .flatten()
        .map(|entry| entry.path())
        .filter(|path| {
            path.file_name()
                .and_then(|name| name.to_str())
                .is_some_and(|name| name.ends_with(".metadata.json"))
        })
        .collect();
    files.sort();
    files.pop()
}

fn table_ident(kind: TableKind) -> TableIdent {
    TableIdent::new(
        NamespaceIdent::new(NAMESPACE.to_string()),
        kind.name().to_string(),
    )
}

fn table_creation(kind: TableKind) -> Result<TableCreation> {
    Ok(TableCreation::builder()
        .name(kind.name().to_string())
        .schema(kind.iceberg_schema()?)
        .build())
}

fn read_journal(root: &Path) -> Result<Option<StepJournal>> {
    let path = root.join(NAMESPACE).join(STEP_JOURNAL);
    if !path.is_file() {
        return Ok(None);
    }
    let text = std::fs::read_to_string(&path)?;
    Ok(Some(serde_json::from_str(&text)?))
}

async fn write_one_file(table: &Table, batch: RecordBatch) -> Result<Vec<DataFile>> {
    let location = DefaultLocationGenerator::new(table.metadata())?;
    let names = DefaultFileNameGenerator::new(
        "part".to_string(),
        Some(unique_file_suffix()),
        iceberg::spec::DataFileFormat::Parquet,
    );
    let parquet = ParquetWriterBuilder::new(
        WriterProperties::default(),
        table.metadata().current_schema().clone(),
    );
    // Parquet has no in-place append. One file per commit means the writer must
    // not roll. A real simulation that outgrows memory is a separate problem;
    // the file-size cap is not what splits this file.
    let rolling = RollingFileWriterBuilder::new(
        parquet,
        usize::MAX,
        table.file_io().clone(),
        location,
        names,
    );
    let mut writer = DataFileWriterBuilder::new(rolling).build(None).await?;
    writer.write(batch).await?;
    Ok(writer.close().await?)
}

async fn commit_overwrite(
    catalog: &MemoryCatalog,
    table: &Table,
    files: Vec<DataFile>,
) -> Result<Table> {
    if files.len() != 1 {
        return Err(Error::new(format!(
            "overwrite expected 1 data file, writer produced {}",
            files.len()
        )));
    }
    let snapshot_id = fresh_snapshot_id(table);
    let sequence = table.metadata().next_sequence_number();
    let suffix = unique_file_suffix();
    let location = table.metadata().location().trim_end_matches('/');
    let manifest_path = format!("{location}/metadata/{suffix}-m0.avro");
    let manifest_output = table.file_io().new_output(&manifest_path)?;
    let schema = table.metadata().current_schema().clone();
    let spec = table.metadata().default_partition_spec().as_ref().clone();
    let builder = ManifestWriterBuilder::new(manifest_output, Some(snapshot_id), schema, spec);
    let mut manifest_writer = match table.metadata().format_version() {
        FormatVersion::V1 => builder.build_v1(),
        FormatVersion::V2 => builder.build_v2_data(),
        FormatVersion::V3 => builder.build_v3_data(),
    };
    for file in files {
        manifest_writer.add_file(file, sequence)?;
    }
    let manifest = manifest_writer.write_manifest_file().await?;

    let list_path = format!("{location}/metadata/snap-{snapshot_id}-0-{suffix}.avro");
    let list_output = table.file_io().new_output(&list_path)?.writer().await?;
    let parent = table.metadata().current_snapshot_id();
    let mut list_writer = match table.metadata().format_version() {
        FormatVersion::V1 => ManifestListWriter::v1(list_output, snapshot_id, parent),
        FormatVersion::V2 => ManifestListWriter::v2(list_output, snapshot_id, parent, sequence),
        FormatVersion::V3 => ManifestListWriter::v3(
            list_output,
            snapshot_id,
            parent,
            sequence,
            Some(table.metadata().next_row_id()),
        ),
    };
    list_writer.add_manifests(std::iter::once(manifest))?;
    let next_row_id = list_writer.next_row_id();
    list_writer.close().await?;

    let snapshot = Snapshot::builder()
        .with_manifest_list(list_path)
        .with_snapshot_id(snapshot_id)
        .with_parent_snapshot_id(parent)
        .with_sequence_number(sequence)
        .with_summary(Summary {
            operation: Operation::Overwrite,
            additional_properties: HashMap::new(),
        })
        .with_schema_id(table.metadata().current_schema_id())
        .with_timestamp_ms(now_millis());
    let snapshot = if table.metadata().format_version() >= FormatVersion::V3 {
        let first = table.metadata().next_row_id();
        let added = next_row_id.unwrap_or(first).saturating_sub(first);
        snapshot.with_row_range(first, added).build()
    } else {
        snapshot.build()
    };
    publish_metadata(
        catalog,
        table,
        vec![
            TableUpdate::AddSnapshot { snapshot },
            TableUpdate::SetSnapshotRef {
                ref_name: MAIN_BRANCH.to_string(),
                reference: branch_ref(snapshot_id),
            },
        ],
    )
    .await
}

async fn publish_metadata(
    catalog: &MemoryCatalog,
    table: &Table,
    updates: Vec<TableUpdate>,
) -> Result<Table> {
    let current_loc = table.metadata_location_result()?.to_string();
    let mut builder = table
        .metadata()
        .clone()
        .into_builder(Some(current_loc.clone()));
    for update in updates {
        builder = update.apply(builder)?;
    }
    let metadata = builder.build()?.metadata;
    let next = MetadataLocation::from_str(&current_loc)?.with_next_version();
    metadata.write_to(table.file_io(), &next).await?;
    let ident = table.identifier().clone();
    catalog.drop_table(&ident).await?;
    Ok(catalog.register_table(&ident, next.to_string()).await?)
}

fn fresh_snapshot_id(table: &Table) -> i64 {
    let mut id = now_millis().saturating_mul(1000);
    if id <= 0 {
        id = 1;
    }
    let used: HashSet<i64> = table
        .metadata()
        .snapshots()
        .map(|snapshot| snapshot.snapshot_id())
        .collect();
    while used.contains(&id) {
        id = id.saturating_add(1);
    }
    id
}

async fn referenced_parquet(table: &Table) -> Result<HashSet<PathBuf>> {
    let Some(snapshot) = table.metadata().current_snapshot() else {
        return Ok(HashSet::new());
    };
    let list = table.manifest_list_reader(snapshot).load().await?;
    let mut keep = HashSet::new();
    for manifest in list.entries() {
        let loaded = manifest.load_manifest(table.file_io()).await?;
        for entry in loaded.entries() {
            if entry.is_alive() {
                keep.insert(normalize_path(&uri_to_path(entry.file_path())));
            }
        }
    }
    Ok(keep)
}

fn delete_unreferenced(root: &Path, table: &str, keep: &HashSet<PathBuf>) -> Result<()> {
    let data = root.join(NAMESPACE).join(table).join("data");
    if !data.is_dir() {
        return Ok(());
    }
    for path in walk_parquet(&data) {
        if !keep.contains(&normalize_path(&path)) {
            std::fs::remove_file(&path)?;
        }
    }
    Ok(())
}

fn walk_parquet(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let Ok(entries) = std::fs::read_dir(dir) else {
        return out;
    };
    for entry in entries.flatten() {
        let path = entry.path();
        if path.is_dir() {
            out.extend(walk_parquet(&path));
        } else if path.extension().and_then(|ext| ext.to_str()) == Some("parquet") {
            out.push(path);
        }
    }
    out
}

fn uri_to_path(uri: &str) -> PathBuf {
    let Some(rest) = uri.strip_prefix("file://") else {
        return PathBuf::from(uri);
    };
    // `file:///C:/qsm` strips to `/C:/qsm`. Drop that extra slash so Windows
    // sees a drive path. A Unix path such as `/tmp/qsm` has no drive letter.
    let bytes = rest.as_bytes();
    let rest = if bytes.len() >= 3 && bytes[0] == b'/' && bytes[2] == b':' {
        &rest[1..]
    } else {
        rest
    };
    PathBuf::from(rest)
}

fn normalize_path(path: &Path) -> PathBuf {
    path.canonicalize().unwrap_or_else(|_| path.to_path_buf())
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

#[cfg(test)]
mod tests {
    use std::path::Path;

    use super::*;

    #[test]
    fn file_uris_use_forward_slashes_and_round_trip() {
        let unix = file_uri(Path::new("/tmp/qsm/pyramid"));
        assert_eq!(unix, "file:///tmp/qsm/pyramid");
        assert_eq!(uri_to_path(&unix), Path::new("/tmp/qsm/pyramid"));

        let drive = file_uri(Path::new(r"\\?\C:\qsm\pyramid"));
        assert_eq!(drive, "file:///C:/qsm/pyramid");
        assert!(!drive.contains('\\'));
        assert_eq!(uri_to_path(&drive), Path::new("C:/qsm/pyramid"));
    }

    #[test]
    fn names_reject_windows_reserved_characters_and_devices() {
        assert!(validate_name("pyramid").is_ok());
        for bad in [
            "a:b", "a*b", "a?b", "a\"b", "a<b", "a>b", "a|b", "con", "NUL", "COM1",
        ] {
            assert!(validate_name(bad).is_err(), "{bad}");
        }
    }
}
