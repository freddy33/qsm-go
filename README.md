# Quantum State Machine

A 3D integer-lattice cellular automaton. Space is a graph whose nodes each connect to three others. Time is an integer. Events grow along that graph.

The implementation is a Cargo workspace in [`rust/`](rust/README.md). Saved simulations are local Iceberg tables of Parquet files. The column reference is [docs/db/README.md](docs/db/README.md).

## Build, test, run, view

[Rustup](https://rustup.rs/) is required. `rust/rust-toolchain.toml` selects the pinned toolchain. These commands work from any directory. `QSM_DATA_ROOT` overrides the data directory; when it is unset the scripts use `data/` at the repository root.

```bash
./run build
./run test
./run run pyramid --new
./run view pyramid
```

The same cargo commands, run from `rust/`:

```bash
cargo build --workspace
cargo test --workspace
cargo run -p qsm -- pyramid --new
cargo run -p qsm-view --features view -- pyramid
```

`qsm <name>` creates or resumes `data/<name>/` and grows it to `max_time_steps` from `inputs.json`. When it finishes it prints the elapsed time, the time per step, the peak resident memory, and the nodes and points created. `./run run <name> --new` archives an existing `data/<name>/` and writes a new `inputs.json` from [`simulation-configs/<name>-config.json`](simulation-configs/README.md), with the defaults filled in. `qsm-view <name>` draws that folder. It opens with **Display active nodes** on, using `active_path_node_threshold` from the run, and event roots stay visible under every filter. Bevy is the `view` feature, so `cargo test --workspace` does not compile the window. `./run view <name> --screenshot out.png` saves one frame.

![Default pyramid at creation time 6, active nodes on](docs/qsm-view-pyramid.png)

Details, controls, and the inputs format: [rust/README.md](rust/README.md).

## History

Originally written in Go (freddy33/qsm-go, 2018); the project started in 2014 in [freddy33/qsm](https://github.com/freddy33/qsm) (Java).
