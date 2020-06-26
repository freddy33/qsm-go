// +build !windows

package m3db

import (
	"github.com/freddy33/qsm-go/utils/m3util"
	"os/exec"
	"path/filepath"
)

func CopyFile(src, dest string) {
	cmd := exec.Command("cp", src, dest)
	err := cmd.Run()
	m3util.ExitOnError(err)
}

func DbDrop(envNumber string) {
	rootDir := m3util.GetGitRootDir()
	cmd := exec.Command("bash", filepath.Join(rootDir, "qsm"), "-env", envNumber, "db", "drop")
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Errorf("failed to destroy environment %d at OS level due to %v with output: ***\n%s\n***", envNumber, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("destroy environment %d at OS level output: ***\n%s\n***", envNumber, string(out))
		}
	}
}

func checkOsEnv(envNumber string) {
	rootDir := m3util.GetGitRootDir()
	cmd := exec.Command("bash", filepath.Join(rootDir, "qsm"), "-env", envNumber, "db", "check")
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Fatalf("failed to check environment %s at OS level due to %v with output: ***\n%s\n***", envNumber, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("check environment %s at OS output: ***\n%s\n***", envNumber, string(out))
		}
	}
}

func FillDb(envNumber string) {
	rootDir := m3util.GetGitRootDir()
	cmd := exec.Command("bash", filepath.Join(rootDir, "qsm"), "-env", envNumber, "run", "filldb")
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Fatalf("failed to fill db for test environment %s at OS level due to %v with output: ***\n%s\n***", envNumber, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("check environment %s at OS output: ***\n%s\n***", envNumber, string(out))
		}
	}
}
