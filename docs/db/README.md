# QSM database tables (from `m3db.TableDefinition` usages)

Source: freddy33/qsm-go, branch `prod` @ e2be02a (the `rust` branch has the same definitions). Registry: [backend/m3db/m3table.go](https://github.com/freddy33/qsm-go/blob/prod/backend/m3db/m3table.go). Each env creates the tables in its own Postgres schema `qsm<envId>` (see [m3db.go](https://github.com/freddy33/qsm-go/blob/prod/backend/m3db/m3db.go)).

| Table | Go package / file | Rows | Purpose |
|---|---|---|---|
| connection_details | pointdb · [m3conndb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/pointdb/m3conndb.go) | 50 | The 50 unit connection vectors (x,y,z) with squared length ds. |
| trio_details | pointdb · [m3conndb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/pointdb/m3conndb.go) | 200 | The 200 trios, each made of 3 connections. |
| growth_contexts | pointdb · [m3growthctxdb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/pointdb/m3growthctxdb.go) | 52 | Growth contexts identified by (ctx_type, ctx_index). |
| trio_cubes | pointdb · [m3triocubedb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/pointdb/m3triocubedb.go) | 5192 | Per-context cube of trios: center, 6 face centers, 12 edge middles. |
| path_builders | pointdb · [m3pathbuilderdb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/pointdb/m3pathbuilderdb.go) | 5192 | Path-builder rules for each trio cube (1:1 with trio_cubes): root trio, intermediate trios, connections. |
| points | pathdb · [pathdb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/pathdb/pathdb.go) | open | Unique 3D integer lattice points. |
| path_contexts | pathdb · [pathdb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/pathdb/pathdb.go) | 200 | A growth context + offset, bound to its path builder, with max distance reached. |
| path_nodes | pathdb · [pathdb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/pathdb/pathdb.go) | open | Nodes of a growth path: point, trio, distance d, links to up to 3 neighbour path nodes. |
| spaces | spacedb · [spacedb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/spacedb/spacedb.go) | open | A named simulation space and its limits/thresholds. |
| events | spacedb · [spacedb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/spacedb/spacedb.go) | open | An event in a space, at a creation time, with a color, growing along one path context. |
| nodes | spacedb · [spacedb.go](https://github.com/freddy33/qsm-go/blob/prod/backend/spacedb/spacedb.go) | open | Event nodes in space-time: a path node placed at a creation time, linked to up to 3 other nodes. |

## Diagram

![QSM DB schema](qsm-db.svg)

```mermaid
%% QSM PostgreSQL schema, from every m3db.TableDefinition in freddy33/qsm-go (branch prod @ e2be02a)
%% Each env gets its own schema "qsm<envId>" (e.g. qsm1); all tables + FKs live inside that schema.
%% Solid lines = declared REFERENCES. Grouped edges list every FK column they stand for.
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
        integer max_dist "NOT NULL DEFAULT 0"
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
    spaces {
        serial id PK
        varchar_175 name UK "NOT NULL UNIQUE"
        smallint active_path_node_threshold "NOT NULL"
        smallint max_trios_per_point "NOT NULL"
        smallint max_path_nodes_per_point "NOT NULL"
        integer max_coord "NOT NULL"
        integer max_node_time "NOT NULL"
    }
    events {
        serial id PK
        integer space_id FK "NOT NULL"
        integer path_ctx_id FK "NOT NULL"
        integer creation_time "NOT NULL"
        smallint color "NOT NULL"
        integer end_time "NOT NULL"
        integer max_node_time "NOT NULL"
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
    spaces ||--o{ events : "space_id"
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

## Notes for the Rust migration

1. `TableDefinition.Indexes` is declared for `events` (3), `nodes` (3) and `path_nodes` (1, on `d`) but `m3table.go` never executes it, so those indexes are never created.
2. Denormalized data with no DB consistency check: `nodes.trio_id`/`nodes.point_id` duplicate what `nodes.path_node_id -> path_nodes` gives; `path_builders.ctx_id` duplicates `trio_cubes.ctx_id`; nothing checks a node's path node belongs to its event's `path_ctx_id`.
3. Uneven nullability: FKs in `trio_details`/`trio_cubes` (and `connection_details` x/y/z/ds) are nullable while the matching `path_builders` columns are NOT NULL.
4. Mixed-case column names like `center_faces_PX` are unquoted, so Postgres stores them lowercase.
5. Typo: Go function `creatPathNodesTableDef`.
