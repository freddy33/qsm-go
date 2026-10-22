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
fn qsm_test_sim_resumes_in_a_temp_dir() {
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
    assert!(stdout.contains("distance: 0"), "{stdout}");
    assert!(stdout.contains("point: 0,0,0"), "{stdout}");
    assert!(tmp.0.join("test-sim").join("qsm").join("points").is_dir());

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
    assert!(stdout.contains("distance: 1"), "{stdout}");
    assert!(stdout.contains("creation_time: 1"), "{stdout}");
    assert!(stdout.contains("point: 3,0,0"), "{stdout}");

    let missing = Command::new(bin).output().unwrap();
    assert_eq!(missing.status.code(), Some(2));
}
