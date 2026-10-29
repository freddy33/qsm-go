//! Load and save QSM simulation rows as local Iceberg tables of Parquet files.
//!
//! A run lives in `data/<simulation name>/`. Static reference data
//! (connections, trios, growth contexts, trio cubes, path builders) is
//! [`static_pack`] and is not written. The five open tables are Iceberg tables
//! in the `qsm` namespace. Space parameters live in `inputs.json`. Each
//! successful `qsm` execution also writes `qsm/run-<utc>.json`. Each table is
//! one Parquet file. See `docs/db/README.md`.

mod batch;
mod engine;
mod error;
mod inputs;
mod rows;
mod runfile;
mod schema;
mod sim;
mod store;

pub use error::{Error, Result};
pub use inputs::{expand_config, load_config, load_inputs, SimulationInputs};
pub use m3point::PointPackData;
pub use rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow};
pub use runfile::{latest_run, RunRecord};
pub use schema::{TableKind, NAMESPACE};
pub use sim::{
    config_root, default_data_root, run_named, run_named_async, run_new, run_new_async, StepReport,
};
pub use store::{validate_name, SimulationStore};

/// Static reference tables, built once in memory. They are not stored.
pub fn static_pack() -> &'static PointPackData {
    PointPackData::shared()
}

#[cfg(test)]
mod tests;
