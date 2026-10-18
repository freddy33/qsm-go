# QSM in Rust

Rust port of the Quantum State Machine: a 3D integer-lattice cellular automaton. The Go tree is unchanged. This directory is a Cargo workspace with one crate per Go area.

Lattice coordinates, connection ids, trio indexes, growth contexts, cube keys, and path steps stay integers, as in the Go code. The one formula that uses `float32` in Go, the predicted path size, is kept as `f32` and truncated toward zero the same way.

Derived counts match the Go tests:

| data | count |
| --- | ---: |
| connections | 50 |
| trios | 200 |
| growth contexts | 52 |
| trio cubes | 5192 |
| path builders | 5192 |

## Crates

| crate | Go source | what it does |
| --- | --- | --- |
| `m3util` | `m3util` | environment ids `0..16`, positive moduli, bit count, errors |
| `m3point` | `model/m3point`, `backend/pointdb` | points, connections, base trios, valid-next and mod-4/mod-8 permutations, growth contexts, trio cubes, path builders, murmur point map |
| `m3path` | `model/m3path`, `backend/pathdb` | 4-bit connection state packed in a `u16`, event link ids, predicted size |
| `m3space` | `model/m3space`, `backend/spacedb` | event colors, active-link threshold, three-id groups, index rows, pyramids |
| `m3db` | `backend/m3db`, `backend/pointdb` | PostgreSQL via sqlx: schema `qsm{env}` and inserts for connections, trios, and growth contexts |
| `m3gl` | `ui/m3gl` | drawing filters, colors, dimmers, and object types |
| `playgl` | `ui/playgl` | process entry point |

`PointPackData::shared()` builds the pack once. `playgl` prints the counts above.

## Build and run

[Rustup](https://rustup.rs/) is required. `rust-toolchain.toml` selects Rust 1.99.1, because the sqlx 0.8 dependency tree uses Rust edition 2024.

```bash
cd rust
cargo build --workspace
cargo test --workspace
cargo run -p playgl
```

`cargo test` does not need PostgreSQL. The live round trip runs only when `QSM_DATABASE_URL` is set:

```bash
QSM_DATABASE_URL=postgres://user:password@localhost:5432/qsm cargo test -p m3db -- --nocapture
```

That test creates schema `qsm_rust_test`, writes the pack, checks the three row counts, and drops the schema. Connection strings also follow the Go form (`host= port= user= password= dbname= sslmode=disable`) on `DbConnDetails`.

## Ported

- Connection vectors and the 50 connection details, ordered as in Go.
- The 8 base trios, the 12 valid next-trio pairs, and the 12 mod-4 and 12 mod-8 permutations, including the original fill rules.
- All 52 growth contexts and the trio index selected from a main point.
- The 5192 trio cubes and 5192 path builders, including the forward and back walk used by the Go path-builder test.
- Connection-state nibbles, event link ids, and the predicted-size formula.
- Event color masks, active links, three-id grouping, index generation, and pyramid size.
- sqlx persistence of connections, trios, and growth contexts.
- Viewer filter, color, and dimmer rules, covered by unit tests.

## Not ported

- **OpenGL window.** `ui/m3gl` and `ui/playgl` open a GLFW window with `go-gl`. `m3gl` has the drawing rules only. `playgl` prints the pack counts and exits. A window should use `glow` + `glutin` or `wgpu` and consume `m3gl::SpaceDrawingElement`.
- **HTTP server** (`backend/m3server`) and the **protobuf API** (`model/m3api`) plus the Go client.
- **Database-backed space-time stepper**: `spacedb` next/populate and path-node growth that reads and writes nodes.
- **Inserts for `trio_cubes` and `path_builders`.** The table-name constants exist. `save_point_pack` writes connections, trios, and growth contexts. Cubes and path builders are rebuilt in memory.
