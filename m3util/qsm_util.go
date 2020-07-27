package m3util

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
)

var Log = NewLogger("m3util", INFO)

func DirExists(dir string, subPath string) (bool, string) {
	p := filepath.Join(dir, subPath)
	fi, err := os.Stat(p)
	if os.IsNotExist(err) {
		return false, p
	}
	if err != nil {
		Log.Errorf("searching for %s folder in %s returned unknown error %v", subPath, dir, err)
		return false, p
	}
	return fi != nil && fi.IsDir(), p
}

func AbsPath(dir string) string {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		Log.Fatalf("could not extract absolute path returned unknown error %v", err)
		return ""
	}
	return absPath
}

func GetGitRootDir() string {
	absPath := AbsPath(".")
	p := absPath
	// Check first if we are below the checkout dir
	if b, p := DirExists(p, "qsm-go"); b {
		if b, _ = DirExists(p, ".git"); b {
			return p
		} else {
			Log.Fatalf("found qsm-go sub folder at %s which not a git checkout", p)
			return ""
		}
	}
	for {
		if p == "." || p == "/" {
			Log.Fatalf("did not find path with git under %s", absPath)
			return ""
		}
		if b, _ := DirExists(p, ".git"); b {
			return p
		}
		p = filepath.Dir(p)
	}
}

func getOrCreateBuildSubDir(subPath string) string {
	buildDir := GetBuildDir()
	b, p := DirExists(buildDir, subPath)
	if !b {
		err := os.MkdirAll(p, 0755)
		if err != nil {
			Log.Fatalf("could not create sub build dir %s due to error %v", p, err)
			return ""
		}
	}
	return p
}

func GetBuildDir() string {
	gitRootDir := GetGitRootDir()
	b, p := DirExists(gitRootDir, "build")
	if !b {
		err := os.MkdirAll(p, 0755)
		if err != nil {
			Log.Fatalf("could not create build dir %s due to error %v", p, err)
			return ""
		}
	}
	return p
}

func GetConfDir() string {
	b, p := DirExists(GetGitRootDir(), "backend/conf")
	if !b {
		Log.Fatalf("conf dir %s does not exists!", p)
		return ""
	}
	return p
}

func CreateFile(dir, fileName string) *os.File {
	p := filepath.Join(dir, fileName)
	f, err := os.Create(p)
	if err != nil {
		Log.Fatalf("could not create file %s due to %v", p, err)
		return nil
	}
	return f
}

func GetGenDocDir() string {
	return getOrCreateBuildSubDir("gendoc")
}

func ExitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

/*func writeNextBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	ExitOnError(err)
}
*/

func CloseFile(file *os.File) {
	ExitOnError(file.Close())
}

func CloseBody(body io.ReadCloser) {
	ExitOnError(body.Close())
}

func WriteNextString(file *os.File, text string) {
	_, err := file.WriteString(text)
	ExitOnError(err)
}

func WriteAll(writer *csv.Writer, records [][]string) {
	ExitOnError(writer.WriteAll(records))
}

func Write(writer *csv.Writer, records []string) {
	ExitOnError(writer.Write(records))
}

func SetEnvQuietly(key, value string) {
	ExitOnError(os.Setenv(key, value))
}

func PosMod2(i uint64) uint64 {
	return i & 0x0000000000000001
}

func PosMod4(i uint64) uint64 {
	return i & 0x0000000000000003
}

func PosMod8(i uint64) uint64 {
	return i & 0x0000000000000007
}

