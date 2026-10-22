//! Load and save QSM simulation rows as local Iceberg tables of Parquet files.
//!
//! A run lives in `data/<simulation name>/`. Static reference data
//! (connections, trios, growth contexts, trio cubes, path builders) is
//! [`static_pack`] and is not written. The five open tables are Iceberg tables
//! in the `qsm` namespace. Space parameters live in `inputs.json`. Each
//! successful `qsm` execution also writes `qsm/run-<utc>.json`. See
//! `docs/db/README.md` for the column list and the partition layout.

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
pub use m3point::PointPackData;
pub use rows::{EventRow, NodeRow, PathContextRow, PathNodeRow, PointRow};
pub use runfile::RunRecord;
pub use schema::{TableKind, BUCKET_WIDTH, NAMESPACE};
pub use sim::{default_data_root, run_named, run_named_async, StepReport};
pub use store::SimulationStore;

/// Static reference tables, built once in memory. They are not stored.
pub fn static_pack() -> &'static PointPackData {
    PointPackData::shared()
}

#[cfg(test)]
mod tests;
