//! Run a named simulation from its `inputs.json` up to `max_time_steps`.

use std::fmt;
use std::path::{Path, PathBuf};
use std::time::{Instant, SystemTime, UNIX_EPOCH};

use crate::engine::{run_simulation, RunStats};
use crate::error::{Error, Result};
use crate::inputs::{default_inputs, load_config, load_inputs, resolve_events, save_inputs};
use crate::runfile::{ensure_compatible, format_utc, latest_run, write_run, RunRecap};
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
    pub points: usize,
    pub steps_computed: i32,
    pub recap: RunRecap,
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
        writeln!(f, "points: {}", self.points)?;
        writeln!(f, "steps_computed: {}", self.steps_computed)?;
        writeln!(f, "elapsed: {}", format_duration(self.recap.elapsed_ms))?;
        writeln!(f, "{}", format_steps(&self.recap.step_ms))?;
        let memory = match self.recap.peak_rss_bytes {
            Some(bytes) => format_memory(bytes),
            None => "unavailable".to_string(),
        };
        writeln!(f, "memory: {memory}")?;
        write!(
            f,
            "created: nodes {}, points {}",
            self.recap.nodes_created, self.recap.points_created
        )
    }
}

impl From<RunStats> for StepReport {
    fn from(stats: RunStats) -> Self {
        let recap = RunRecap {
            elapsed_ms: stats.elapsed_ms,
            step_ms: stats.step_ms,
            peak_rss_bytes: stats.peak_rss_bytes,
            nodes: stats.nodes as u64,
            nodes_created: stats.nodes_created as u64,
            points: stats.points as u64,
            points_created: stats.points_created as u64,
        };
        Self {
            simulation: String::new(),
            folder: PathBuf::new(),
            resumed: stats.resumed,
            max_time_steps: stats.max_time_steps,
            time: stats.time,
            events: stats.events,
            nodes: stats.nodes,
            points: stats.points,
            steps_computed: stats.steps_computed,
            recap,
        }
    }
}

/// `400ms` below one second, otherwise `1.500s`.
pub(crate) fn format_duration(ms: u64) -> String {
    if ms < 1000 {
        format!("{ms}ms")
    } else {
        format!("{:.3}s", ms as f64 / 1000.0)
    }
}

/// `512 KiB` below 1 MiB, otherwise `12.0 MiB`.
pub(crate) fn format_memory(bytes: u64) -> String {
    const MIB: u64 = 1024 * 1024;
    if bytes < MIB {
        let kib = (bytes as f64 / 1024.0).round() as u64;
        format!("{kib} KiB")
    } else {
        format!("{:.1} MiB", bytes as f64 / MIB as f64)
    }
}

pub(crate) fn format_steps(step_ms: &[u64]) -> String {
    if step_ms.is_empty() {
        "steps: 0".to_string()
    } else {
        let parts: Vec<_> = step_ms.iter().copied().map(format_duration).collect();
        format!("steps: {} ({})", step_ms.len(), parts.join(", "))
    }
}

/// Peak resident set size of this process, in bytes.
///
/// Unix uses `getrusage`. Linux and Android report kilobytes; macOS and the
/// other Unix targets report bytes. If that call fails, Linux
/// `/proc/self/status` `VmHWM` is the fallback. Windows uses
/// `GetProcessMemoryInfo` and `PeakWorkingSetSize`, which is already bytes.
/// `None` means the OS did not report a peak.
pub fn peak_rss_bytes() -> Option<u64> {
    platform_peak_rss()
}

#[cfg(unix)]
fn platform_peak_rss() -> Option<u64> {
    let mut usage = std::mem::MaybeUninit::<libc::rusage>::zeroed();
    let rc = unsafe { libc::getrusage(libc::RUSAGE_SELF, usage.as_mut_ptr()) };
    if rc == 0 {
        let usage = unsafe { usage.assume_init() };
        if usage.ru_maxrss > 0 {
            let raw = usage.ru_maxrss as u64;
            let bytes = if cfg!(any(target_os = "linux", target_os = "android")) {
                raw.saturating_mul(1024)
            } else {
                raw
            };
            return Some(bytes);
        }
    }
    vm_hwm_bytes()
}

#[cfg(unix)]
fn vm_hwm_bytes() -> Option<u64> {
    let text = std::fs::read_to_string("/proc/self/status").ok()?;
    parse_vm_hwm(&text)
}

#[cfg(windows)]
fn platform_peak_rss() -> Option<u64> {
    use std::mem::{size_of, MaybeUninit};
    use windows_sys::Win32::System::ProcessStatus::{
        GetProcessMemoryInfo, PROCESS_MEMORY_COUNTERS,
    };
    use windows_sys::Win32::System::Threading::GetCurrentProcess;

    let mut counters = MaybeUninit::<PROCESS_MEMORY_COUNTERS>::zeroed();
    let mut counters = unsafe { counters.assume_init() };
    counters.cb = u32::try_from(size_of::<PROCESS_MEMORY_COUNTERS>()).ok()?;
    let ok = unsafe { GetProcessMemoryInfo(GetCurrentProcess(), &mut counters, counters.cb) };
    if ok == 0 {
        return None;
    }
    let peak = counters.PeakWorkingSetSize as u64;
    (peak > 0).then_some(peak)
}

#[cfg(not(any(unix, windows)))]
fn platform_peak_rss() -> Option<u64> {
    None
}

pub(crate) fn parse_vm_hwm(text: &str) -> Option<u64> {
    for line in text.lines() {
        if let Some(rest) = line.strip_prefix("VmHWM:") {
            let kb: u64 = rest.split_whitespace().next()?.parse().ok()?;
            return Some(kb.saturating_mul(1024));
        }
    }
    None
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
    let started = Instant::now();
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
    let mut stats = run_simulation(&mut store, &inputs).await?;
    stats.elapsed_ms = u64::try_from(started.elapsed().as_millis()).unwrap_or(u64::MAX);
    stats.peak_rss_bytes = peak_rss_bytes();
    let mut report = StepReport::from(stats);
    write_run(
        store.root(),
        &inputs,
        &raw,
        report.time,
        report.recap.clone(),
    )?;
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
    fn recap_lines_use_stable_units() {
        assert_eq!(format_duration(0), "0ms");
        assert_eq!(format_duration(400), "400ms");
        assert_eq!(format_duration(999), "999ms");
        assert_eq!(format_duration(1000), "1.000s");
        assert_eq!(format_duration(1100), "1.100s");
        assert_eq!(format_duration(1500), "1.500s");
        assert_eq!(format_memory(100), "0 KiB");
        assert_eq!(format_memory(512 * 1024), "512 KiB");
        assert_eq!(format_memory(12 * 1024 * 1024), "12.0 MiB");
        assert_eq!(format_memory(1536 * 1024), "1.5 MiB");
        assert_eq!(format_steps(&[]), "steps: 0");
        assert_eq!(format_steps(&[400, 1100]), "steps: 2 (400ms, 1.100s)");
        assert_eq!(
            parse_vm_hwm("Name:\tfoo\nVmHWM:\t  2048 kB\nVmPeak:\t  3000 kB\n"),
            Some(2048 * 1024)
        );
        assert_eq!(parse_vm_hwm("VmPeak:\t  1 kB\n"), None);
        assert!(peak_rss_bytes().is_some_and(|bytes| bytes > 0));

        let report = StepReport {
            simulation: "pyramid".to_string(),
            folder: PathBuf::from("data/pyramid"),
            resumed: false,
            max_time_steps: 2,
            time: 2,
            events: 4,
            nodes: 40,
            points: 20,
            steps_computed: 2,
            recap: RunRecap {
                elapsed_ms: 1500,
                step_ms: vec![400, 1100],
                peak_rss_bytes: Some(12 * 1024 * 1024),
                nodes: 40,
                nodes_created: 4,
                points: 20,
                points_created: 2,
            },
        };
        let text = report.to_string();
        assert!(text.contains("elapsed: 1.500s\n"), "{text}");
        assert!(text.contains("steps: 2 (400ms, 1.100s)\n"), "{text}");
        assert!(text.contains("memory: 12.0 MiB\n"), "{text}");
        assert!(text.ends_with("created: nodes 4, points 2"), "{text}");
        assert!(text.contains("steps_computed: 2\n"), "{text}");
        assert!(text.contains("nodes: 40\n"), "{text}");
        assert!(text.contains("points: 20\n"), "{text}");
    }

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
