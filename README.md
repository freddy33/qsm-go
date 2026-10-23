# Quantum State Machine

A 3D integer-lattice cellular automaton. Space is a graph whose nodes each connect to three others. Time is an integer. Events grow along that graph.

The implementation is a Cargo workspace in [`rust/`](rust/README.md). Saved simulations are local Iceberg tables of Parquet files. The column reference is [docs/db/README.md](docs/db/README.md).

## Build, test, run, view

[Rustup](https://rustup.rs/) is required. `rust/rust-toolchain.toml` selects the pinned toolchain. These commands work from any directory. `QSM_DATA_ROOT` overrides the data directory; when it is unset the scripts use `data/` at the repository root.

```bash
./run build
./run test
./run run pyramid-demo
./run view pyramid-demo
```

The same cargo commands, run from `rust/`:

```bash
cargo build --workspace
cargo test --workspace
cargo run -p qsm -- pyramid-demo
cargo run -p qsm-view --features view -- pyramid-demo
```

`qsm <name>` creates or resumes `data/<name>/` and grows it to `max_time_steps` from `inputs.json`. `qsm-view <name>` draws that folder. Bevy is the `view` feature, so `cargo test --workspace` does not compile the window. `./run view <name> --screenshot out.png` saves one frame.

![Default pyramid at creation time 12](docs/qsm-view-pyramid.png)

Details, controls, and the inputs format: [rust/README.md](rust/README.md).

## History

Originally written in Go (freddy33/qsm-go, 2018); the project started in 2014 in [freddy33/qsm](https://github.com/freddy33/qsm) (Java).
