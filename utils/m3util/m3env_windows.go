// +build windows

package m3util

import (
	"fmt"
	"os/exec"
	"strings"
)

func WslPath(path string) string {
	wp := strings.ReplaceAll(path, "\\", "/")
	wp = strings.ReplaceAll(wp, "C:", "/mnt/c")
	wp = strings.ReplaceAll(wp, " ", "\\ ")
	return wp
}

func CopyFile(src, dest string) {
	cmd := exec.Command("cmd.exe", "/C", fmt.Sprintf("copy \"%s\" \"%s\"", src, dest))
	err := cmd.Run()
	ExitOnError(err)
}

func RunQsm(id QsmEnvID, params ...string) {
	rootDir := GetGitRootDir()
	command := fmt.Sprintf("%s -env %s", WslPath(filepath.Join(rootDir, "qsm")), id.String())
	for _, p := range params {
		command += " " + p
	}
	cmd := exec.Command("bash.exe", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Errorf("failed to run %v in env %d due to %v with output: ***\n%s\n***", params, id, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("run %v in env %d output: ***\n%s\n***", params, id, string(out))
		}
	}
}
