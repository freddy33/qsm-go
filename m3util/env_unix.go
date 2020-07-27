// +build !windows

package m3util

import (
	"os"
	"os/exec"
	"path/filepath"
)

func CopyFile(src, dest string) {
	cmd := exec.Command("cp", src, dest)
	err := cmd.Run()
	ExitOnError(err)
}

func osStartBackend(id QsmEnvID, params ...string) *os.Process {
	rootDir := GetGitRootDir()
	args := make([]string, 3+len(params))
	name := filepath.Join(rootDir, "backend/backend")
	args[0] = "server"
	args[1] = "-env"
	args[2] = id.String()
	for idx, p := range params {
		args[idx+3] = p
	}
	Log.Infof("Running start backend command %v", args)
	cmd := exec.Command(name, args...)
	err := cmd.Start()
	if err != nil {
		Log.Errorf("failed to start %v in env %d due to %v", params, id, err)
	}
	return cmd.Process
}

func osQsmCmd(id QsmEnvID, params ...string) {
	rootDir := GetGitRootDir()
	args := make([]string, 3+len(params))
	args[0] = filepath.Join(rootDir, "qsm")
	args[1] = "-env"
	args[2] = id.String()
	for idx, p := range params {
		args[idx+3] = p
	}
	Log.Infof("Running qsm command %v", args)
	cmd := exec.Command("bash", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Errorf("failed to run %v in env %d due to %v with output: ***\n%s\n***", params, id, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("run %v in env %d output: ***\n%s\n***", params, id, string(out))
		}
	}
}
