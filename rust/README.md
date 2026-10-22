# QSM in Rust

Rust port of the Quantum State Machine: a 3D integer-lattice cellular automaton. The Go tree stays, except the okteto launch scripts. This directory is a Cargo workspace with one crate per Go area.

Lattice coordinates, connection ids, trio indexes, growth contexts, cube keys, and path steps stay integers, as in the Go code. The one formula that uses `float32` in Go, the predicted path size, is kept as `f32` and truncated toward zero the same way.

Derived counts match the Go tests:

| data | count |
| --- | ---: |
| connections | 50 |
| trios | 200 |
| growth contexts | 52 |
| trio cubes | 5192 |
| path builders | 5192 |

Storage layout: [docs/db/README.md](../docs/db/README.md).

## Crates

| crate | Go source | what it does |
| --- | --- | --- |
| `m3util` | `m3util` | environment ids `0..16`, positive moduli, bit count, errors |
| `m3point` | `model/m3point`, `backend/pointdb` | points, connections, base trios, valid-next and mod-4/mod-8 permutations, growth contexts, trio cubes, path builders, murmur point map |
| `m3path` | `model/m3path`, `backend/pathdb` | 4-bit connection state, path-context growth, event link ids, predicted size |
| `m3space` | `model/m3space`, `backend/spacedb` | event colors, active-link threshold, three-id groups, index rows, pyramids |
| `m3data` | `backend/m3db`, `backend/pathdb`, `backend/spacedb` | `inputs.json`, event growth, and local Iceberg tables of Parquet files |
| `qsm` | `backend` entry point | `qsm <simulation name>` reads `inputs.json` and grows up to its max time |
| `m3gl` | `ui/m3gl` | drawing filters, colors, dimmers, and object types |
| `playgl` | `ui/playgl` | process entry point |

`PointPackData::shared()` builds the static reference pack once (connections, trios, growth contexts, trio cubes, path builders). Those tables are not written. `playgl` prints the counts above.

## Simulation data

`qsm <simulation name>` creates `data/<simulation name>/` under the current directory, or resumes that folder. `QSM_DATA_ROOT` overrides the `data` directory.

```bash
cd rust
cargo run -p qsm -- my-sim
QSM_DATA_ROOT=/tmp/qsm cargo run -p qsm -- my-sim
```

The initial state is `data/<simulation name>/inputs.json`. A new folder with no file and no saved rows gets the default below, then the run starts. A later invocation reads the same file and grows from the last saved creation time until `max_time_steps`. It does not rewrite `inputs.json`.

One simulation name is one space. There is no `spaces` table. After a successful run, `qsm` writes `data/<name>/qsm/run-<YYYYMMDDTHHMMSSZ>.json` (UTC, so the names sort in time order). The next run of that name loads the latest `run-*.json` and compares it with the current `inputs.json`. Resolved space parameters and the event list (shortcuts and explicit events, compared after expansion) must be identical. Only `max_time_steps` may change, and a value below the saved creation time is an error: nothing is written and no new run file is created. Any other difference is an error and writes nothing. Saved rows with no `run-*.json` are an error as well. A same-second collision is `run-<stamp>_01.json`, which sorts after the unsuffixed name and before the next second. The `doc` field is a note and is not part of the check. Raising `max_time_steps` resumes from the last saved creation time.

`max_time_steps` is the last creation time to store. Events created at time 0 contribute their root nodes at time 0, then one shell per later time. `3` therefore stores times 0, 1, 2, and 3.

```json
{
  "doc": "Default is the playgl pyramid.",
  "max_time_steps": 3,
  "space": {
    "active_path_node_threshold": 3,
    "max_trios_per_point": 2,
    "max_path_nodes_per_point": 4,
    "max_coord": 18
  },
  "events": [],
  "shortcuts": [
    {
      "pyramid": {
        "size": 4,
        "growth_type": 8,
        "growth_index": 0,
        "offsets": [0, 2, 4, 6],
        "creation_time": 0
      }
    }
  ]
}
```

An explicit event uses the Go `CreateEvent` fields. `color` is 1 red, 2 green, 4 blue, 8 yellow:

```json
{
  "x": 0,
  "y": 0,
  "z": 0,
  "creation_time": 0,
  "color": 1,
  "growth_type": 8,
  "growth_index": 0,
  "growth_offset": 0
}
```

`events` are created as written, then each shortcut is expanded. The only shortcut today is `pyramid`, matching `CreatePyramidWithParams` in `ui/m3gl/m3gl.go` (and the same four calls in `backend/spacedb/pyramid.go`). playgl calls it with size 4, growth type 8, index 0, and offsets 0, 2, 4, 6 when the world has no events. The four centers are `{3,0,3}`, `{-3,3,3}`, `{-3,-3,3}`, `{0,0,-3}` multiplied by `size`, in the order red, green, blue, yellow. A new shortcut is another variant on that enum, serialized as its own key.

The default file omits the sample `events` entry and keeps only the pyramid. Space defaults are playgl `MakeWorld`: active threshold 3, max trios per point 2, max path nodes per point 4, max coord 18. Growth does not consult them, matching the Go TODO in `m3spacetimenode.go`.

A finished default run writes `qsm/run-20261022T121800Z.json`. `inputs_sha256` is the SHA-256 of the raw `inputs.json` bytes. `events` is the pyramid after expansion. `creation_time` is the highest saved node time:

```json
{
  "inputs_sha256": "3adb4bb7fb389c14f853f3ed72280b0495149c46dd2220207e54a2b6e7d867af",
  "inputs": {
    "doc": "Default is the playgl pyramid: CreatePyramidWithParams(space, 4, growth types 8, indexes 0, offsets 0,2,4,6). max_time_steps is the last creation time stored. Time 0 is the four root nodes.",
    "max_time_steps": 3,
    "space": {
      "active_path_node_threshold": 3,
      "max_trios_per_point": 2,
      "max_path_nodes_per_point": 4,
      "max_coord": 18
    },
    "events": [],
    "shortcuts": [
      {
        "pyramid": {
          "size": 4,
          "growth_type": 8,
          "growth_index": 0,
          "offsets": [0, 2, 4, 6],
          "creation_time": 0
        }
      }
    ]
  },
  "space": {
    "active_path_node_threshold": 3,
    "max_trios_per_point": 2,
    "max_path_nodes_per_point": 4,
    "max_coord": 18
  },
  "events": [
    { "x": 12, "y": 0, "z": 12, "creation_time": 0, "color": 1, "growth_type": 8, "growth_index": 0, "growth_offset": 0 },
    { "x": -12, "y": 12, "z": 12, "creation_time": 0, "color": 2, "growth_type": 8, "growth_index": 0, "growth_offset": 2 },
    { "x": -12, "y": -12, "z": 12, "creation_time": 0, "color": 4, "growth_type": 8, "growth_index": 0, "growth_offset": 4 },
    { "x": 0, "y": 0, "z": -12, "creation_time": 0, "color": 8, "growth_type": 8, "growth_index": 0, "growth_offset": 6 }
  ],
  "max_time_steps": 3,
  "creation_time": 3
}
```

Changing `space.max_coord` from 18 to 21 and running again fails before any write:

```text
qsm: inputs.json does not match the last run (run-20261022T121800Z.json); only max_time_steps may change:
  space.max_coord: last run 18, inputs 21
```

Each table is one Parquet file under `qsm/<table>/data/`. Parquet cannot be appended in place, and iceberg-rust 0.10 has no public overwrite action (`fast_append` keeps the old files). At the end of a successful step Rust rewrites each dirty table into a new single file, commits an Iceberg snapshot whose operation is `overwrite` and whose manifest names only that file, then renames `qsm/step.json` to record the snapshot ids. A crash before that journal leaves the previous step in place. Old Parquet is deleted only after the journal names the new snapshot.

`points`, `path_contexts`, `events`, and event `nodes` are written once. `path_contexts` does not store `max_dist`, and `events` does not store `max_node_time`. A path node cannot be fully immutable: it is saved when it is created, including the open frontier, and the next shell fills `Next` and `Blocked` on that same id. The rewrite keeps one row per id. The cost is a full copy of each dirty table every step, so bytes written grow with the square of the row count for `path_nodes` and `nodes`.

Uniqueness of `(x, y, z)`, `(path_ctx_id, point_id)`, and `(event_id, point_id)` is still checked in Rust before the write. A second row with the same id is rejected.

`SimulationStore::open` loads every stored table into memory. A later viewer can call that with the simulation name and read `points()`, `path_nodes()`, `events()`, and `nodes()`. There is no API server.

## Build and run

[Rustup](https://rustup.rs/) is required. `rust-toolchain.toml` selects Rust 1.99.1.

```bash
cd rust
cargo build --workspace
cargo test --workspace
cargo run -p playgl
cargo run -p qsm -- test-sim
```

`cargo test` does not need PostgreSQL or a network. The `qsm` integration test runs `qsm test-sim` twice in a temp directory: the first run writes the default pyramid and grows to time 3, and the second run finds that max already reached.

## Ported

- Connection vectors and the 50 connection details, ordered as in Go.
- The 8 base trios, the 12 valid next-trio pairs, and the 12 mod-4 and 12 mod-8 permutations, including the original fill rules.
- All 52 growth contexts and the trio index selected from a main point.
- The 5192 trio cubes and 5192 path builders, including the forward and back walk used by the Go path-builder test.
- Connection-state nibbles, event link ids, and the predicted-size formula.
- Path-context growth (`calculateNextMaxDist` / `makeNewNodes`) and event growth (`increaseMaxNodeTime`), including neighbour links.
- Event color masks, active links, three-id grouping, index generation, and pyramid size.
- Local Iceberg/Parquet load and save for points, path contexts, path nodes, events, and nodes. `qsm` grows from `inputs.json`, writes `qsm/run-<utc>.json` after a successful run, and resumes until `max_time_steps`.
- Viewer filter, color, and dimmer rules, covered by unit tests.

## Not ported

- **OpenGL window.** `ui/m3gl` and `ui/playgl` open a GLFW window with `go-gl`. `m3gl` has the drawing rules only. `playgl` prints the pack counts and exits. A window should use `glow` + `glutin` or `wgpu`, load a simulation by name through `m3data`, and consume `m3gl::SpaceDrawingElement`.
- **HTTP server** (`backend/m3server`) and the **protobuf API** (`model/m3api`) plus the Go client. The Rust entry point is the `qsm` CLI.
- **Space-time populate map.** Growth writes one row per event node. It does not build the deduped `SpaceTimeNode` map `Populate` uses for drawing, and the active-threshold window does not filter what is saved.
- The Go packages still contain their PostgreSQL code. The Rust workspace does not link it.
