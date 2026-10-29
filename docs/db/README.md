# QSM database tables

The column lists below are the reference for the Rust Iceberg tables in `m3data`. Static reference tables are built in memory and are not written. `points`, `path_contexts`, `path_nodes`, `events`, and `nodes` are Parquet files with an Iceberg metadata index under `data/<simulation name>/`. One simulation folder is one space: space parameters live in `inputs.json`, not in a table. See [Iceberg layout](#iceberg-layout).

The diagram and the Postgres column types are the historical Go-era schema. Go stored those tables in a Postgres schema `qsm<envId>`. That code is no longer in this tree; it remains in git history at [83f3e19](https://github.com/freddy33/qsm-go/commit/83f3e19) (`backend/m3db/m3table.go`, `backend/m3db/m3db.go`). The Rust workspace does not use Postgres.

| Table | Go-era source (historical) | Rows | Purpose |
|---|---|---|---|
| connection_details | pointdb · [m3conndb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/pointdb/m3conndb.go) | 50 | The 50 unit connection vectors (x,y,z) with squared length ds. |
| trio_details | pointdb · [m3conndb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/pointdb/m3conndb.go) | 200 | The 200 trios, each made of 3 connections. |
| growth_contexts | pointdb · [m3growthctxdb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/pointdb/m3growthctxdb.go) | 52 | Growth contexts identified by (ctx_type, ctx_index). |
| trio_cubes | pointdb · [m3triocubedb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/pointdb/m3triocubedb.go) | 5192 | Per-context cube of trios: center, 6 face centers, 12 edge middles. |
| path_builders | pointdb · [m3pathbuilderdb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/pointdb/m3pathbuilderdb.go) | 5192 | Path-builder rules for each trio cube (1:1 with trio_cubes): root trio, intermediate trios, connections. |
| points | pathdb · [pathdb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/pathdb/pathdb.go) | open | Unique 3D integer lattice points. |
| path_contexts | pathdb · [pathdb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/pathdb/pathdb.go) | 200 | A growth context + offset, bound to its path builder. |
| path_nodes | pathdb · [pathdb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/pathdb/pathdb.go) | open | Nodes of a growth path: point, trio, distance d, links to up to 3 neighbour path nodes. |
| events | spacedb · [spacedb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/spacedb/spacedb.go) | open | An event at a creation time, with a color, growing along one path context. |
| nodes | spacedb · [spacedb.go](https://github.com/freddy33/qsm-go/blob/83f3e19/backend/spacedb/spacedb.go) | open | Event nodes in space-time: a path node placed at a creation time, linked to up to 3 other nodes. |

The Go-era Postgres schema also had a `spaces` table, `events.space_id`, `path_contexts.max_dist`, and `events.max_node_time` (`backend/spacedb/spacedb.go` and `backend/pathdb/pathdb.go` at 83f3e19). Rust does not. A folder is the space, and its limits live in `inputs.json`. The distance a path has reached, and the latest creation time of an event, are the max `d` and `creation_time` of the stored nodes.

## Diagram

![QSM DB schema](qsm-db.svg)

```mermaid
%% Rust Iceberg tables under data/<simulation>/. Static pointdb tables are in memory.
%% The Go-era Postgres schema had a spaces table, events.space_id,
%% path_contexts.max_dist, and events.max_node_time. Rust does not store those.
%% One simulation folder is one space. Solid lines are the logical references.
erDiagram
    %% ---------- backend/pointdb : static reference data (fixed row counts) ----------
    connection_details {
        smallint id PK "50 rows"
        integer x
        integer y
        integer z
        bigint ds "squared length"
    }
    trio_details {
        smallint id PK "200 rows"
        smallint conn1 FK "nullable"
        smallint conn2 FK "nullable"
        smallint conn3 FK "nullable"
    }
    growth_contexts {
        smallint id PK "52 rows"
        smallint ctx_type UK "UK(ctx_type, ctx_index)"
        smallint ctx_index UK
    }
    trio_cubes {
        smallint id PK "5192 rows"
        smallint ctx_id FK "nullable"
        smallint center FK "nullable"
        smallint center_faces_PX_MX_PY_MY_PZ_MZ FK "6 cols, nullable"
        smallint middle_edges_PXPY_to_MYMZ FK "12 cols, nullable"
    }
    path_builders {
        smallint id PK,FK "5192 rows; id = trio_cubes.id"
        smallint ctx_id FK "NOT NULL"
        smallint root FK "NOT NULL"
        smallint inter1_2_3 FK "3 cols, NOT NULL"
        smallint conn_11_12_21_22_31_32 FK "6 cols, NOT NULL"
        smallint last_inter_11_to_32 FK "6 cols, NOT NULL"
        smallint next_main_conn_11_to_32 FK "6 cols, NOT NULL"
        smallint next_inter_conn_11_to_32 FK "6 cols, NOT NULL"
    }
    %% ---------- backend/pathdb : growth paths ----------
    points {
        bigserial id PK
        integer x UK "NOT NULL; UK points_x_y_z_key(x,y,z)"
        integer y UK "NOT NULL"
        integer z UK "NOT NULL"
    }
    path_contexts {
        serial id PK "200 rows expected"
        smallint growth_ctx_id FK,UK "NOT NULL; UK path_ctx_growth_param_key"
        smallint growth_offset UK "NOT NULL"
        smallint path_builders_id FK "NULL until built"
    }
    path_nodes {
        bigserial id PK
        integer path_ctx_id FK,UK "NOT NULL; UK unique_point_per_path_ctx"
        smallint path_builders_id FK "NOT NULL"
        smallint path_builder_idx "NOT NULL"
        smallint trio_id FK "NOT NULL"
        bigint point_id FK,UK "NOT NULL"
        integer d "NOT NULL DEFAULT 0; idx path_nodes_d_index (unused)"
        smallint connection_mask "NOT NULL DEFAULT 0"
        bigint path_node1 FK "nullable self-ref"
        bigint path_node2 FK "nullable self-ref"
        bigint path_node3 FK "nullable self-ref"
    }
    %% ---------- backend/spacedb : simulation runs ----------
    events {
        serial id PK
        integer path_ctx_id FK "NOT NULL"
        integer creation_time "NOT NULL"
        smallint color "NOT NULL"
        integer end_time "NOT NULL"
    }
    nodes {
        bigserial id PK
        integer event_id FK,UK "NOT NULL; UK unique_point_per_event_node"
        bigint path_node_id FK "NOT NULL"
        smallint trio_id FK "NOT NULL"
        bigint point_id FK,UK "NOT NULL"
        integer d "NOT NULL"
        integer creation_time "NOT NULL"
        smallint connection_mask "NOT NULL DEFAULT 0"
        bigint node1 FK "nullable self-ref"
        bigint node2 FK "nullable self-ref"
        bigint node3 FK "nullable self-ref"
    }

    %% pointdb
    connection_details |o--o{ trio_details : "conn1, conn2, conn3"
    growth_contexts |o--o{ trio_cubes : "ctx_id"
    trio_details |o--o{ trio_cubes : "center + 6 faces + 12 edges (19 FKs)"
    trio_cubes ||--o| path_builders : "id (PK = FK, 1:1)"
    growth_contexts ||--o{ path_builders : "ctx_id"
    trio_details ||--o{ path_builders : "root, inter1-3, last_inter* (10 FKs)"
    connection_details ||--o{ path_builders : "conn*, next_main_conn*, next_inter_conn* (18 FKs)"
    %% pathdb
    growth_contexts ||--o{ path_contexts : "growth_ctx_id"
    path_builders |o--o{ path_contexts : "path_builders_id"
    path_contexts ||--o{ path_nodes : "path_ctx_id"
    path_builders ||--o{ path_nodes : "path_builders_id"
    trio_details ||--o{ path_nodes : "trio_id"
    points ||--o{ path_nodes : "point_id"
    path_nodes |o--o{ path_nodes : "path_node1/2/3"
    %% spacedb
    path_contexts ||--o{ events : "path_ctx_id"
    events ||--o{ nodes : "event_id"
    path_nodes ||--o{ nodes : "path_node_id"
    trio_details ||--o{ nodes : "trio_id"
    points ||--o{ nodes : "point_id"
    nodes |o--o{ nodes : "node1/2/3"
```

- `qsm-db.mmd`: Mermaid ER diagram source (GitHub renders it inline below)
- `qsm-db.svg`: rendered diagram
- `qsm-db.puml`: PlantUML version

Grouped rows in `trio_cubes` and `path_builders` (e.g. "6 cols") stand for repeated FK columns; edge labels list them.

## Notes on the historical Postgres schema

1. `TableDefinition.Indexes` is declared for `events` (3), `nodes` (3) and `path_nodes` (1, on `d`) but `m3table.go` never executes it, so those indexes are never created.
2. Denormalized data with no DB consistency check: `nodes.trio_id`/`nodes.point_id` duplicate what `nodes.path_node_id -> path_nodes` gives; `path_builders.ctx_id` duplicates `trio_cubes.ctx_id`; nothing checks a node's path node belongs to its event's `path_ctx_id`.
3. Uneven nullability: FKs in `trio_details`/`trio_cubes` (and `connection_details` x/y/z/ds) are nullable while the matching `path_builders` columns are NOT NULL.
4. Mixed-case column names like `center_faces_PX` are unquoted, so Postgres stores them lowercase.
5. Typo: Go function `creatPathNodesTableDef`.

## Iceberg layout

Rust storage uses the `iceberg` crate 0.10 (apache/iceberg-rust) for table metadata and the `arrow` and `parquet` crates 58.4 (arrow-rs) for the data files. There is no catalog service, no Postgres, and no network. Each simulation is a local warehouse at `data/<simulation name>/`.

iceberg-rust 0.10 has no filesystem catalog. The process opens a memory catalog whose warehouse is the absolute path of `data/<simulation name>` (`/data/...` on Unix, `C:/data/...` on Windows) and whose storage is the local filesystem. A `file:///` URI is not used: iceberg's local filesystem turns `file:///C:/...` into `/C:/...`, which Windows cannot create. Older metadata that still contains `file:///...` is still read. Creating a table writes `qsm/<table>/metadata/00000-<uuid>.metadata.json`. A later process does not trust the in-memory index: it scans that metadata directory, picks the highest zero-padded version, and registers that file. The latest `*.metadata.json` is the durable commit pointer.

Namespace: `qsm`. Format version: 2. No table is partitioned. Ids are assigned in Rust (`max(id) + 1`, starting at 1). Iceberg does not enforce primary keys, foreign keys, or unique constraints. Before a row is staged, Rust rejects a non-positive id and any repeat of:

- `points` `(x, y, z)` and `id`
- `path_contexts` `(growth_ctx_id, growth_offset)` and `id`
- `path_nodes` `(path_ctx_id, point_id)` and `id`
- `events` `id`
- `nodes` `(event_id, point_id)` and `id`

Load reads the current snapshot as the whole table. A repeated id in that file is an error. There is no "keep the newest copy of an id" pass.

Rows are immutable except one case. `points`, `path_contexts`, `events`, and event `nodes` are written once. `path_contexts` has no `max_dist` column and `events` has no `max_node_time` column; the Go-era Postgres schema had both. A path node is created with its `From` link set and its other slots `NotSet`. The next shell, which may be a later `qsm` invocation, fills those slots with `Next` or `Blocked`. That final mask is not known at the first write, and the open frontier has to be on disk so a crash can resume, so a later step replaces the link columns of that same id. The natural key `(path_ctx_id, point_id)` does not change, and the rewritten file still contains the id once. Event nodes copy the path-node mask at creation and are not revised.

### One data file per table

Parquet cannot be appended in place, and iceberg-rust 0.10 does not expose an overwrite action: `fast_append` keeps every previous data file, and `TableCommit`'s builder is crate-private. A commit therefore rewrites the table. The rolling writer target is `usize::MAX`, so the new rows and the rows already stored become one Parquet file. Rust then writes a manifest and a manifest list that name only that file, and writes the next `*.metadata.json` with an Iceberg snapshot whose operation is `overwrite` and whose `main` ref points at it. That metadata file is the commit for the table. The memory catalog is pointed at it with `drop_table` plus `register_table`.

A step stages every dirty table and commits them in order: points, path contexts, path nodes, event nodes, events. `qsm/step.json` is renamed into place only after those commits. It records each table's current snapshot id (`null` when the table is still empty). Iceberg has no cross-table transaction, so the journal is the barrier. A crash before it leaves the new snapshots in the metadata but not in the journal. The next open moves `main` back to the journaled snapshot, or clears `main` when the journal is missing or says `null`, and only then deletes Parquet files the restored snapshot does not name. Old snapshots stay in the metadata so that rollback can still see them. A crash after the journal keeps the new step; the next open deletes the previous data files.

The cost is that every dirty table is copied in full at every step. `points`, `path_nodes`, and `nodes` grow with the simulation, so the bytes written across a run are quadratic in the number of rows. A large simulation pays that copy on each time step. The single file is the property this layout is for; a later format can switch to many files if that copy dominates.

### Tables

| Table | Stored | Files | When rows are written |
| --- | --- | --- | --- |
| `connection_details`, `trio_details`, `growth_contexts`, `trio_cubes`, `path_builders` | no | — | Derived. `PointPackData::shared()` rebuilds the 50 / 200 / 52 / 5192 / 5192 rows at startup. |
| `points` | yes | one Parquet file | Appended as new coordinates appear. The whole file is rewritten at the end of the step. |
| `path_contexts` | yes | one Parquet file | One row per growth context and offset, written once when the context is created. |
| `path_nodes` | yes | one Parquet file | New nodes are added at the step that creates them, including the open frontier. A later step may replace link columns of an existing id, then the whole file is rewritten. |
| `events` | yes | one Parquet file | One row per event, written once when the event is created. |
| `nodes` | yes | one Parquet file | New event nodes are added at the end of the step that creates them. A stored event node is not revised. |

Postgres `smallint`, `integer`, and `serial` columns are Iceberg `int` (32-bit). `bigserial` / `bigint` are `long`. `varchar` is `string`. Nullable columns stay optional. The `id` column is field id 1 and is the Iceberg identifier field. Other field ids follow the column order in the diagram above.

Nullable link columns follow `GetLinkIdsForDb`. `NotSet`, `Blocked`, and a `Next` that is not assigned yet are NULL. `From`, and a `Next` that already has an id, store that positive id. The connection state itself lives in `connection_mask`, so a NULL link is not ambiguous on read. Event nodes copy the path-node mask at the time they are created. Go's `getNodeLinkIds` does not point an event `Next` at the following node, and a new shell has no `Next` yet, so event `node1/2/3` are either NULL or the earlier event node on a `From` link.

`inputs.json` sits next to `qsm/` and is not a table. It holds `max_time_steps`, the space parameters, explicit events, and shortcuts such as `pyramid`. The run grows until every event's max creation time reaches `max_time_steps`. Resume uses the nodes already stored and the max from that file.

After a successful `qsm <name>`, Rust writes `qsm/run-<YYYYMMDDTHHMMSSZ>.json` (UTC). The file records the sha256 of the raw `inputs.json`, the parsed inputs, the resolved space parameters, the event list after shortcut expansion, the `max_time_steps` used, and the last saved creation time. The next run of that name loads the latest `run-*.json` (lexical order; a same-second file is `run-<stamp>_01.json`, which sorts after the unsuffixed name). Space parameters and expanded events must match. Only `max_time_steps` may change, and it may not drop below that file's `creation_time`. A lower value is an error, nothing is written, and no new run file is created. Any other difference is an error and writes nothing. Saved rows with no `run-*.json` are an error too. The `doc` string is not part of that check.

### Folder tree

```text
data/<simulation name>/
├── inputs.json
└── qsm/
    ├── step.json
    ├── run-20261022T121800Z.json
    ├── points/
    │   ├── data/part-<n>-<pid>-<seq>.parquet
    │   └── metadata/
    │       ├── 00000-<uuid>.metadata.json
    │       ├── 00001-<uuid>.metadata.json
    │       └── snap-<id>-0-<suffix>.avro
    ├── path_contexts/
    │   └── data/part-<n>-<pid>-<seq>.parquet
    ├── path_nodes/
    │   └── data/part-<n>-<pid>-<seq>.parquet
    ├── events/
    │   └── data/part-<n>-<pid>-<seq>.parquet
    └── nodes/
        └── data/part-<n>-<pid>-<seq>.parquet
```

`00000-*.metadata.json` is the empty table. Each commit writes the next metadata version, one manifest, and one manifest list. After `step.json` names that snapshot, the previous Parquet file is deleted, so `data/` holds one file. Opening the folder registers the highest metadata version, rolls `main` back when it disagrees with `step.json`, then deletes unreferenced Parquet. `data/` at the repository root is gitignored, and so is `rust/data/` when the CLI is run from `rust/`.
