//! `data/<simulation>/qsm/run-<utc>.json`, written after a successful `qsm` run.
//!
//! A later run of the same name loads the latest file and requires the resolved
//! space parameters and the expanded event list to match. Only `max_time_steps`
//! may change, and it may not drop below the saved creation time. The file is
//! not written when the run fails that check or fails while growing.

use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};

use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

use crate::error::{Error, Result};
use crate::inputs::{expand_events, EventInput, SimulationInputs, SpaceParams};

/// Space state captured when a run finished.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct RunRecord {
    /// SHA-256 of the raw `inputs.json` bytes used for this run.
    pub inputs_sha256: String,
    /// Parsed `inputs.json`, with serde defaults filled in.
    pub inputs: SimulationInputs,
    /// Space parameters after defaults. One simulation is one space.
    pub space: SpaceParams,
    /// Explicit events, then shortcut expansion, in creation order.
    pub events: Vec<EventInput>,
    /// `max_time_steps` from `inputs.json` for this run.
    pub max_time_steps: i32,
    /// Highest `nodes.creation_time` saved when this run finished.
    pub creation_time: i32,
    /// How long the run took and how much memory it used.
    ///
    /// Absent on run files written before this field existed. It is not part of
    /// [`ensure_compatible`]: two runs with different recaps still match.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub recap: Option<RunRecap>,
}

/// Timing and memory recorded at the end of a successful `qsm` run.
///
/// This is a report of that invocation. It is not an input, and a later run
/// does not have to match it.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct RunRecap {
    pub elapsed_ms: u64,
    #[serde(default)]
    pub step_ms: Vec<u64>,
    /// Peak resident memory. Absent when the OS did not report one.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub peak_rss_bytes: Option<u64>,
    pub nodes: u64,
    pub nodes_created: u64,
    pub points: u64,
    pub points_created: u64,
}

/// Newest `qsm/run-*.json` under the simulation folder, by file name.
pub fn latest_run(root: &Path) -> Result<Option<(PathBuf, RunRecord)>> {
    let mut files = run_files(root)?;
    files.sort();
    let Some(path) = files.pop() else {
        return Ok(None);
    };
    let text = std::fs::read_to_string(&path)?;
    let record: RunRecord = serde_json::from_str(&text).map_err(|err| {
        Error::new(format!(
            "cannot read {}: {err}",
            path.file_name()
                .and_then(|name| name.to_str())
                .unwrap_or("run file")
        ))
    })?;
    Ok(Some((path, record)))
}

pub fn run_files(root: &Path) -> Result<Vec<PathBuf>> {
    let dir = root.join("qsm");
    let Ok(entries) = std::fs::read_dir(&dir) else {
        return Ok(Vec::new());
    };
    let mut files = Vec::new();
    for entry in entries {
        let path = entry?.path();
        let Some(name) = path.file_name().and_then(|name| name.to_str()) else {
            continue;
        };
        if path.is_file() && name.starts_with("run-") && name.ends_with(".json") {
            files.push(path);
        }
    }
    Ok(files)
}

/// Reject a run whose space parameters or expanded events differ from `previous`.
pub fn ensure_compatible(
    run_path: &Path,
    previous: &RunRecord,
    inputs: &SimulationInputs,
) -> Result<()> {
    let expanded = expand_events(inputs)?;
    let mut lines = Vec::new();
    diff_space(&previous.space, &inputs.space, &mut lines);
    diff_events(&previous.events, &expanded, &mut lines);
    if lines.is_empty() {
        return Ok(());
    }
    let name = run_path
        .file_name()
        .and_then(|name| name.to_str())
        .unwrap_or("run file");
    let mut message = format!(
        "inputs.json does not match the last run ({name}); only max_time_steps may change:"
    );
    for line in lines {
        message.push('\n');
        message.push_str("  ");
        message.push_str(&line);
    }
    Err(Error::new(message))
}

pub fn write_run(
    root: &Path,
    inputs: &SimulationInputs,
    raw: &[u8],
    creation_time: i32,
    recap: RunRecap,
) -> Result<PathBuf> {
    let record = RunRecord {
        inputs_sha256: sha256_hex(raw),
        inputs: inputs.clone(),
        space: inputs.space.clone(),
        events: expand_events(inputs)?,
        max_time_steps: inputs.max_time_steps,
        creation_time,
        recap: Some(recap),
    };
    let dir = root.join("qsm");
    std::fs::create_dir_all(&dir)?;
    let secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|err| Error::new(err.to_string()))?
        .as_secs();
    let path = next_run_path(&dir, &format_utc(secs))?;
    let text = serde_json::to_string_pretty(&record)?;
    std::fs::write(&path, text + "\n")?;
    Ok(path)
}

fn diff_space(previous: &SpaceParams, current: &SpaceParams, lines: &mut Vec<String>) {
    push_i32(
        lines,
        "space.active_path_node_threshold",
        previous.active_path_node_threshold,
        current.active_path_node_threshold,
    );
    push_i32(
        lines,
        "space.max_trios_per_point",
        previous.max_trios_per_point,
        current.max_trios_per_point,
    );
    push_i32(
        lines,
        "space.max_path_nodes_per_point",
        previous.max_path_nodes_per_point,
        current.max_path_nodes_per_point,
    );
    push_i32(
        lines,
        "space.max_coord",
        previous.max_coord,
        current.max_coord,
    );
}

fn diff_events(previous: &[EventInput], current: &[EventInput], lines: &mut Vec<String>) {
    if previous.len() != current.len() {
        lines.push(format!(
            "events: last run has {} events, inputs expand to {}",
            previous.len(),
            current.len()
        ));
    }
    for (index, (prev, next)) in previous.iter().zip(current.iter()).enumerate() {
        push_i32(lines, &format!("events[{index}].x"), prev.x, next.x);
        push_i32(lines, &format!("events[{index}].y"), prev.y, next.y);
        push_i32(lines, &format!("events[{index}].z"), prev.z, next.z);
        push_i32(
            lines,
            &format!("events[{index}].creation_time"),
            prev.creation_time,
            next.creation_time,
        );
        if prev.color != next.color {
            lines.push(format!(
                "events[{index}].color: last run {}, inputs {}",
                prev.color, next.color
            ));
        }
        if prev.growth_type != next.growth_type {
            lines.push(format!(
                "events[{index}].growth_type: last run {}, inputs {}",
                prev.growth_type, next.growth_type
            ));
        }
        push_i32(
            lines,
            &format!("events[{index}].growth_index"),
            prev.growth_index,
            next.growth_index,
        );
        push_i32(
            lines,
            &format!("events[{index}].growth_offset"),
            prev.growth_offset,
            next.growth_offset,
        );
    }
}

fn push_i32(lines: &mut Vec<String>, name: &str, previous: i32, current: i32) {
    if previous != current {
        lines.push(format!("{name}: last run {previous}, inputs {current}"));
    }
}

pub(crate) fn sha256_hex(bytes: &[u8]) -> String {
    const HEX: &[u8; 16] = b"0123456789abcdef";
    let digest = Sha256::digest(bytes);
    let mut out = String::with_capacity(digest.len() * 2);
    for byte in digest {
        out.push(HEX[(byte >> 4) as usize] as char);
        out.push(HEX[(byte & 0xf) as usize] as char);
    }
    out
}

/// `YYYYMMDDTHHMMSSZ` in UTC. A same-second collision uses `run-<stamp>_NN.json`,
/// which sorts after the unsuffixed name and before the next second.
pub fn format_utc(secs: u64) -> String {
    let days = (secs / 86_400) as i64;
    let tod = secs % 86_400;
    let hour = tod / 3_600;
    let min = (tod % 3_600) / 60;
    let sec = tod % 60;
    let (year, month, day) = civil_from_days(days);
    format!("{year:04}{month:02}{day:02}T{hour:02}{min:02}{sec:02}Z")
}

fn next_run_path(dir: &Path, stamp: &str) -> Result<PathBuf> {
    let plain = dir.join(format!("run-{stamp}.json"));
    if !plain.exists() {
        return Ok(plain);
    }
    for n in 1..1000 {
        let path = dir.join(format!("run-{stamp}_{n:02}.json"));
        if !path.exists() {
            return Ok(path);
        }
    }
    Err(Error::new(format!(
        "too many run files already named run-{stamp}"
    )))
}

/// Howard Hinnant's `civil_from_days`. `z` is days since 1970-01-01.
fn civil_from_days(z: i64) -> (i32, u32, u32) {
    let z = z + 719_468;
    let era = if z >= 0 { z } else { z - 146_096 } / 146_097;
    let doe = (z - era * 146_097) as u64;
    let yoe = (doe - doe / 1460 + doe / 36524 - doe / 146096) / 365;
    let y = yoe as i64 + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let y = if m <= 2 { y + 1 } else { y };
    (y as i32, m as u32, d as u32)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn recap_is_not_part_of_the_lock_and_old_files_load() {
        let inputs = crate::inputs::default_inputs();
        let events = expand_events(&inputs).unwrap();
        let recap = RunRecap {
            elapsed_ms: 12,
            step_ms: vec![4, 8],
            peak_rss_bytes: Some(4096),
            nodes: 4,
            nodes_created: 4,
            points: 1,
            points_created: 1,
        };
        let previous = RunRecord {
            inputs_sha256: String::new(),
            inputs: inputs.clone(),
            space: inputs.space.clone(),
            events: events.clone(),
            max_time_steps: inputs.max_time_steps,
            creation_time: 0,
            recap: Some(recap.clone()),
        };
        ensure_compatible(Path::new("run-test.json"), &previous, &inputs).unwrap();
        let changed = RunRecord {
            recap: Some(RunRecap {
                elapsed_ms: 9_000,
                nodes: 100,
                nodes_created: 0,
                peak_rss_bytes: None,
                ..recap
            }),
            ..previous
        };
        ensure_compatible(Path::new("run-test.json"), &changed, &inputs).unwrap();

        let old = r#"{
            "inputs_sha256": "ab",
            "inputs": { "max_time_steps": 0 },
            "space": {},
            "events": [],
            "max_time_steps": 0,
            "creation_time": 0
        }"#;
        let record: RunRecord = serde_json::from_str(old).unwrap();
        assert!(record.recap.is_none());
        assert_eq!(record.space.active_path_node_threshold, 3);
    }

    #[test]
    fn utc_stamp_is_sortable() {
        assert_eq!(format_utc(0), "19700101T000000Z");
        assert_eq!(format_utc(1_792_671_480), "20261022T121800Z");
        let plain = "run-20261022T121800Z.json";
        let suffix = "run-20261022T121800Z_01.json";
        let later = "run-20261022T121801Z.json";
        assert!(plain < suffix);
        assert!(suffix < later);
    }
}
