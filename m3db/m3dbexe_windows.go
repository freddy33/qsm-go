// +build windows

package m3db

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"os/exec"
	"path/filepath"
)

func CopyFile(src, dest string) {
	cmd := exec.Command("cmd.exe", "/C", fmt.Sprintf("copy \"%s\" \"%s\"", src, dest))
	err := cmd.Run()
	m3util.ExitOnError(err)
}

func DbDrop(envNumber string) {
	rootDir := m3util.GetGitRootDir()
	cmd := exec.Command("bash.exe", "-c", fmt.Sprintf("%s -env %s db drop", filepath.Join(rootDir, "qsm"), envNumber))
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Errorf("failed to destroy environment %s at OS level due to %v with output: ***\n%s\n***", envNumber, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("destroy environment %s at OS level output: ***\n%s\n***", envNumber, string(out))
		}
	}
}

func checkOsEnv(envNumber string) {
	rootDir := m3util.GetGitRootDir()
	cmd := exec.Command("bash.exe", "-c", fmt.Sprintf("%s -env %s db check", filepath.Join(rootDir, "qsm"), envNumber))
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
	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s -env %s run filldb", filepath.Join(rootDir, "qsm"), envNumber))
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Fatalf("failed to fill db for test environment %s at OS level due to %v with output: ***\n%s\n***", envNumber, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("check environment %s at OS output: ***\n%s\n***", envNumber, string(out))
		}
	}
}
