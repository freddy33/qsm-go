use std::fs;
use std::path::PathBuf;
use std::process::Command;
use std::time::{SystemTime, UNIX_EPOCH};

struct TempRoot(PathBuf);

impl TempRoot {
    fn new() -> Self {
        let nanos = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_nanos();
        let path = std::env::temp_dir().join(format!("qsm-cli-{nanos}-{}", std::process::id()));
        fs::create_dir_all(&path).unwrap();
        Self(path)
    }
}

impl Drop for TempRoot {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.0);
    }
}

#[test]
fn qsm_writes_the_default_pyramid_and_resumes() {
    let tmp = TempRoot::new();
    let bin = env!("CARGO_BIN_EXE_qsm");

    let first = Command::new(bin)
        .arg("test-sim")
        .env("QSM_DATA_ROOT", &tmp.0)
        .output()
        .unwrap();
    assert!(
        first.status.success(),
        "stderr {}",
        String::from_utf8_lossy(&first.stderr)
    );
    let stdout = String::from_utf8(first.stdout).unwrap();
    assert!(stdout.contains("simulation: test-sim"), "{stdout}");
    assert!(stdout.contains("resumed: no"), "{stdout}");
    assert!(stdout.contains("time: 3"), "{stdout}");
    assert!(stdout.contains("events: 4"), "{stdout}");
    assert!(stdout.contains("steps_computed: 3"), "{stdout}");
    assert!(tmp.0.join("test-sim").join("qsm").join("points").is_dir());
    let inputs = fs::read_to_string(tmp.0.join("test-sim").join("inputs.json")).unwrap();
    assert!(inputs.contains("\"pyramid\""), "{inputs}");
    assert!(inputs.contains("\"max_time_steps\": 3"), "{inputs}");

    let second = Command::new(bin)
        .arg("test-sim")
        .env("QSM_DATA_ROOT", &tmp.0)
        .output()
        .unwrap();
    assert!(
        second.status.success(),
        "stderr {}",
        String::from_utf8_lossy(&second.stderr)
    );
    let stdout = String::from_utf8(second.stdout).unwrap();
    assert!(stdout.contains("resumed: yes"), "{stdout}");
    assert!(stdout.contains("time: 3"), "{stdout}");
    assert!(stdout.contains("steps_computed: 0"), "{stdout}");

    let missing = Command::new(bin).output().unwrap();
    assert_eq!(missing.status.code(), Some(2));
}

#[test]
fn qsm_new_archives_the_old_folder_and_runs_the_config() {
    let tmp = TempRoot::new();
    let configs = TempRoot::new();
    let bin = env!("CARGO_BIN_EXE_qsm");
    fs::create_dir_all(tmp.0.join("demo")).unwrap();
    fs::write(tmp.0.join("demo").join("marker.txt"), b"old").unwrap();
    fs::write(
        configs.0.join("demo-config.json"),
        br#"{ "max_time_steps": 1, "shortcuts": [ { "pyramid": {} } ] }"#,
    )
    .unwrap();

    let run = Command::new(bin)
        .args(["demo", "--new"])
        .env("QSM_DATA_ROOT", &tmp.0)
        .env("QSM_CONFIG_ROOT", &configs.0)
        .output()
        .unwrap();
    assert!(
        run.status.success(),
        "stderr {}",
        String::from_utf8_lossy(&run.stderr)
    );
    let stdout = String::from_utf8(run.stdout).unwrap();
    assert!(stdout.contains("simulation: demo"), "{stdout}");
    assert!(stdout.contains("resumed: no"), "{stdout}");
    assert!(stdout.contains("time: 1"), "{stdout}");
    assert!(stdout.contains("events: 4"), "{stdout}");

    let inputs = fs::read_to_string(tmp.0.join("demo").join("inputs.json")).unwrap();
    assert!(inputs.contains("\"pyramid\""), "{inputs}");
    assert!(inputs.contains("\"max_time_steps\": 1"), "{inputs}");
    assert!(inputs.contains("\"size\": 4"), "{inputs}");
    assert!(inputs.contains("\"max_coord\": 18"), "{inputs}");
    assert!(!tmp.0.join("demo").join("marker.txt").exists());

    let mut archives = Vec::new();
    for entry in fs::read_dir(&tmp.0).unwrap() {
        let entry = entry.unwrap();
        let name = entry.file_name();
        let name = name.to_string_lossy();
        if name.starts_with("demo-") {
            archives.push(entry.path());
        }
    }
    assert_eq!(archives.len(), 1, "{archives:?}");
    assert_eq!(fs::read(archives[0].join("marker.txt")).unwrap(), b"old");

    let kept = TempRoot::new();
    fs::create_dir_all(kept.0.join("demo")).unwrap();
    fs::write(kept.0.join("demo").join("marker.txt"), b"keep").unwrap();
    let missing = Command::new(bin)
        .args(["demo", "--new"])
        .env("QSM_DATA_ROOT", &kept.0)
        .env("QSM_CONFIG_ROOT", &kept.0)
        .output()
        .unwrap();
    assert!(!missing.status.success());
    let stderr = String::from_utf8(missing.stderr).unwrap();
    assert!(stderr.contains("no simulation config"), "{stderr}");
    assert_eq!(
        fs::read(kept.0.join("demo").join("marker.txt")).unwrap(),
        b"keep"
    );
    let extra = Command::new(bin)
        .args(["demo", "--new", "extra"])
        .output()
        .unwrap();
    assert_eq!(extra.status.code(), Some(2));
}
