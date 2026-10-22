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
| `m3path` | `model/m3path`, `backend/pathdb` | 4-bit connection state packed in a `u16`, event link ids, predicted size |
| `m3space` | `model/m3space`, `backend/spacedb` | event colors, active-link threshold, three-id groups, index rows, pyramids |
| `m3data` | `backend/m3db`, `backend/pathdb`, `backend/spacedb` | load and save simulation rows as local Iceberg tables of Parquet files |
| `qsm` | `backend` entry point | `qsm <simulation name>` seeds a folder or resumes it by one step |
| `m3gl` | `ui/m3gl` | drawing filters, colors, dimmers, and object types |
| `playgl` | `ui/playgl` | process entry point |

`PointPackData::shared()` builds the static reference pack once (connections, trios, growth contexts, trio cubes, path builders). Those tables are not written. `playgl` prints the counts above.

## Simulation data

`qsm <simulation name>` creates `data/<simulation name>/` under the current directory, or resumes that folder when it already has rows. `QSM_DATA_ROOT` overrides the `data` directory, which is how tests use a temp dir.

```bash
cd rust
cargo run -p qsm -- my-sim
QSM_DATA_ROOT=/tmp/qsm cargo run -p qsm -- my-sim
```

A new name seeds one space, one growth-type-1 path context, the origin point, one red event, and the root path node and space node at distance 0. The same name again advances one main point by +3 on x and appends the matching path node and node. The printed distance and point are the last saved step.

`SimulationStore::open` loads every stored table into memory. A later viewer can call that with the simulation name and read `points()`, `path_nodes()`, `events()`, and `nodes()`. There is no API server.

Each append is one Iceberg snapshot for that table (one batch, not one row). Iceberg does not enforce primary keys, so ids are assigned in Rust and uniqueness of `(x, y, z)`, `(path_ctx_id, point_id)`, and `(event_id, point_id)` is checked before the commit.

## Build and run

[Rustup](https://rustup.rs/) is required. `rust-toolchain.toml` selects Rust 1.99.1.

```bash
cd rust
cargo build --workspace
cargo test --workspace
cargo run -p playgl
cargo run -p qsm -- test-sim
```

`cargo test` does not need PostgreSQL or a network. The `qsm` integration test runs `qsm test-sim` twice in a temp directory and checks that the second run resumes at distance 1.

## Ported

- Connection vectors and the 50 connection details, ordered as in Go.
- The 8 base trios, the 12 valid next-trio pairs, and the 12 mod-4 and 12 mod-8 permutations, including the original fill rules.
- All 52 growth contexts and the trio index selected from a main point.
- The 5192 trio cubes and 5192 path builders, including the forward and back walk used by the Go path-builder test.
- Connection-state nibbles, event link ids, and the predicted-size formula.
- Event color masks, active links, three-id grouping, index generation, and pyramid size.
- Local Iceberg/Parquet load and save for points, path contexts, path nodes, spaces, events, and nodes, plus a one-step resume.
- Viewer filter, color, and dimmer rules, covered by unit tests.

## Not ported

- **OpenGL window.** `ui/m3gl` and `ui/playgl` open a GLFW window with `go-gl`. `m3gl` has the drawing rules only. `playgl` prints the pack counts and exits. A window should use `glow` + `glutin` or `wgpu`, load a simulation by name through `m3data`, and consume `m3gl::SpaceDrawingElement`.
- **HTTP server** (`backend/m3server`) and the **protobuf API** (`model/m3api`) plus the Go client. The Rust entry point is the `qsm` CLI.
- **Full space-time stepper.** `qsm` stores one root node per step. The Go path-growth and `spacedb` populate loops are not ported.
- The Go packages still contain their PostgreSQL code. The Rust workspace does not link it.
