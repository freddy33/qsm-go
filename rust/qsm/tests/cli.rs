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
