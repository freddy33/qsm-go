//! Run a named simulation from its `inputs.json` up to `max_time_steps`.

use std::fmt;
use std::path::{Path, PathBuf};

use crate::engine::{run_simulation, RunStats};
use crate::error::Result;
use crate::inputs::{default_inputs, load_inputs, save_inputs};
use crate::store::SimulationStore;

/// What the last saved step was after this invocation.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct StepReport {
    pub simulation: String,
    pub folder: PathBuf,
    pub resumed: bool,
    pub max_time_steps: i32,
    pub time: i32,
    pub events: usize,
    pub nodes: usize,
    pub steps_computed: i32,
}

impl fmt::Display for StepReport {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        writeln!(f, "simulation: {}", self.simulation)?;
        writeln!(f, "folder: {}", self.folder.display())?;
        writeln!(f, "resumed: {}", if self.resumed { "yes" } else { "no" })?;
        writeln!(f, "max_time_steps: {}", self.max_time_steps)?;
        writeln!(f, "time: {}", self.time)?;
        writeln!(f, "events: {}", self.events)?;
        writeln!(f, "nodes: {}", self.nodes)?;
        write!(f, "steps_computed: {}", self.steps_computed)
    }
}

impl From<RunStats> for StepReport {
    fn from(stats: RunStats) -> Self {
        Self {
            simulation: String::new(),
            folder: PathBuf::new(),
            resumed: stats.resumed,
            max_time_steps: stats.max_time_steps,
            time: stats.time,
            events: stats.events,
            nodes: stats.nodes,
            steps_computed: stats.steps_computed,
        }
    }
}

/// Data root for the CLI: `QSM_DATA_ROOT` when set, otherwise `data`.
pub fn default_data_root() -> PathBuf {
    match std::env::var("QSM_DATA_ROOT") {
        Ok(value) => PathBuf::from(value),
        Err(_) => PathBuf::from("data"),
    }
}

/// Open `data_root/<simulation>`, read or create `inputs.json`, and grow to its max time.
pub fn run_named(data_root: impl AsRef<Path>, simulation: &str) -> Result<StepReport> {
    let runtime = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()?;
    runtime.block_on(run_named_async(data_root.as_ref(), simulation))
}

pub async fn run_named_async(data_root: &Path, simulation: &str) -> Result<StepReport> {
    let mut store = SimulationStore::open(data_root, simulation).await?;
    let inputs_path = store.root().join("inputs.json");
    let inputs = if inputs_path.is_file() {
        load_inputs(&inputs_path)?
    } else if store.nodes().is_empty() && store.events().is_empty() && store.points().is_empty() {
        let inputs = default_inputs();
        save_inputs(&inputs_path, &inputs)?;
        inputs
    } else {
        return Err(crate::error::Error::new(format!(
            "simulation {simulation} has saved rows but {} is missing",
            inputs_path.display()
        )));
    };
    let stats = run_simulation(&mut store, &inputs).await?;
    let mut report = StepReport::from(stats);
    report.simulation = simulation.to_string();
    report.folder = store.root().to_path_buf();
    Ok(report)
}
