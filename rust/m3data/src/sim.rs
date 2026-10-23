//! Run a named simulation from its `inputs.json` up to `max_time_steps`.

use std::fmt;
use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};

use crate::engine::{run_simulation, RunStats};
use crate::error::{Error, Result};
use crate::inputs::{default_inputs, load_config, load_inputs, resolve_events, save_inputs};
use crate::runfile::{ensure_compatible, format_utc, latest_run, write_run};
use crate::store::{validate_name, SimulationStore};

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
        Ok(value) if !value.is_empty() => PathBuf::from(value),
        _ => PathBuf::from("data"),
    }
}

/// Directory of `<name>-config.json` files.
///
/// `QSM_CONFIG_ROOT` wins. Otherwise `QSM_HOME/simulation-configs`. Otherwise
/// `./simulation-configs`, or `../simulation-configs` when that is missing
/// (cargo is started from `rust/`).
pub fn config_root() -> PathBuf {
    if let Some(path) = nonempty_env("QSM_CONFIG_ROOT") {
        return path;
    }
    if let Some(home) = nonempty_env("QSM_HOME") {
        return home.join("simulation-configs");
    }
    let here = PathBuf::from("simulation-configs");
    if here.is_dir() {
        return here;
    }
    PathBuf::from("../simulation-configs")
}

fn nonempty_env(key: &str) -> Option<PathBuf> {
    std::env::var(key)
        .ok()
        .filter(|value| !value.is_empty())
        .map(PathBuf::from)
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
    // Empty table metadata from `open` is not saved data. Rows without a run
    // file mean a previous execution did not finish cleanly.
    if store.has_rows() && latest_run(store.root())?.is_none() {
        return Err(Error::new(format!(
            "simulation {simulation} has saved rows but no qsm/run-*.json; refusing to continue an inconsistent folder"
        )));
    }
    let (inputs, raw) = if inputs_path.is_file() {
        let raw = std::fs::read(&inputs_path)?;
        (load_inputs(&inputs_path)?, raw)
    } else if !store.has_rows() {
        let inputs = default_inputs();
        save_inputs(&inputs_path, &inputs)?;
        let raw = std::fs::read(&inputs_path)?;
        (inputs, raw)
    } else {
        return Err(Error::new(format!(
            "simulation {simulation} has saved rows but {} is missing",
            inputs_path.display()
        )));
    };
    if let Some((path, record)) = latest_run(store.root())? {
        ensure_compatible(&path, &record, &inputs)?;
        if inputs.max_time_steps < record.creation_time {
            let name = path
                .file_name()
                .and_then(|name| name.to_str())
                .unwrap_or("run file");
            return Err(Error::new(format!(
                "max_time_steps {} is below the last saved creation time {} ({name}); refusing to shorten the simulation",
                inputs.max_time_steps, record.creation_time
            )));
        }
    }
    let stats = run_simulation(&mut store, &inputs).await?;
    write_run(store.root(), &inputs, &raw, stats.time)?;
    let mut report = StepReport::from(stats);
    report.simulation = simulation.to_string();
    report.folder = store.root().to_path_buf();
    Ok(report)
}

/// Archive `data/<name>/` when it exists, write a new `inputs.json` from
/// `config_root/<name>-config.json`, and run that simulation.
///
/// A missing or invalid config is an error and leaves the existing folder
/// where it is. The archive name is `data/<name>-<YYYYMMDDTHHMMSSZ>/` in UTC.
/// A same-second collision uses `_<NN>`, same as run files.
pub fn run_new(
    data_root: impl AsRef<Path>,
    simulation: &str,
    config_root: impl AsRef<Path>,
) -> Result<StepReport> {
    let runtime = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()?;
    runtime.block_on(run_new_async(
        data_root.as_ref(),
        simulation,
        config_root.as_ref(),
    ))
}

pub async fn run_new_async(
    data_root: &Path,
    simulation: &str,
    config_root: &Path,
) -> Result<StepReport> {
    validate_name(simulation)?;
    let config_path = config_root.join(format!("{simulation}-config.json"));
    if !config_path.is_file() {
        return Err(Error::new(format!(
            "no simulation config at {}; expected simulation-configs/{simulation}-config.json",
            config_path.display()
        )));
    }
    let inputs = load_config(&config_path)?;
    if resolve_events(&inputs)?.is_empty() {
        return Err(Error::new(format!(
            "simulation config {} expands to no events",
            config_path.display()
        )));
    }
    let folder = data_root.join(simulation);
    if folder.exists() {
        let secs = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map_err(|err| Error::new(err.to_string()))?
            .as_secs();
        let archived = archive_folder(data_root, simulation, &format_utc(secs))?;
        std::fs::rename(&folder, &archived)?;
    }
    save_inputs(&folder.join("inputs.json"), &inputs)?;
    run_named_async(data_root, simulation).await
}

fn archive_folder(data_root: &Path, simulation: &str, stamp: &str) -> Result<PathBuf> {
    let plain = data_root.join(format!("{simulation}-{stamp}"));
    if !plain.exists() {
        return Ok(plain);
    }
    for n in 1..1000 {
        let path = data_root.join(format!("{simulation}-{stamp}_{n:02}"));
        if !path.exists() {
            return Ok(path);
        }
    }
    Err(Error::new(format!(
        "too many archived folders already named {simulation}-{stamp}"
    )))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn archive_names_use_the_utc_stamp_and_a_suffix() {
        let nanos = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_nanos();
        let root = std::env::temp_dir().join(format!("qsm-archive-{nanos}"));
        std::fs::create_dir_all(&root).unwrap();
        let plain = archive_folder(&root, "pyramid", "20261023T120000Z").unwrap();
        assert_eq!(
            plain.file_name().and_then(|name| name.to_str()),
            Some("pyramid-20261023T120000Z")
        );
        std::fs::create_dir_all(&plain).unwrap();
        let next = archive_folder(&root, "pyramid", "20261023T120000Z").unwrap();
        assert_eq!(
            next.file_name().and_then(|name| name.to_str()),
            Some("pyramid-20261023T120000Z_01")
        );
        let _ = std::fs::remove_dir_all(&root);
    }
}
