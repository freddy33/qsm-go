package m3util

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
)

type Logger struct {
	log   *log.Logger
	Level LogLevel
}

var Log = NewLogger("m3util", INFO)

func NewLogger(prefix string, level LogLevel) *Logger {
	return &Logger{log.New(os.Stdout, prefix + " ", log.LstdFlags|log.Lshortfile), level}
}

func (l *Logger) Trace(a ...interface{}) {
	if l.Level <= TRACE {
		l.log.Print("TRACE ", fmt.Sprintln(a...))
	}
}

func (l *Logger) Tracef(format string, v ...interface{}) {
	if l.Level <= TRACE {
		l.log.Println("TRACE", fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Debug(a ...interface{}) {
	if l.Level <= DEBUG {
		l.log.Print("DEBUG ", fmt.Sprintln(a...))
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.Level <= DEBUG {
		l.log.Println("DEBUG", fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(a ...interface{}) {
	if l.Level <= INFO {
		l.log.Print("INFO  ", fmt.Sprintln(a...))
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.Level <= INFO {
		l.log.Println("INFO ", fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(a ...interface{}) {
	if l.Level <= WARN {
		l.log.Print("WARN  ", fmt.Sprintln(a...))
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.Level <= WARN {
		l.log.Println("WARN ", fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(a ...interface{}) {
	msg := fmt.Sprintln(a...)
	l.log.Print("ERROR ", msg)
	log.Print(msg)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Println("ERROR", msg)
	log.Println(msg)
}

func (l *Logger) Fatal(a ...interface{}) {
	msg := fmt.Sprintln(a...)
	l.log.Print("FATAL ", msg)
	log.Fatal(msg)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Println("FATAL", msg)
	log.Fatal(msg)
}

func ChangeToDocsGeneratedDir() {
	changeToDocsSubdir("generated")
}

func ChangeToDocsDataDir() {
	changeToDocsSubdir("data")
}

func changeToDocsSubdir(subDir string) {
	if _, err := os.Stat("docs"); !os.IsNotExist(err) {
		ExitOnError(os.Chdir("docs"))
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			ExitOnError(os.Mkdir(subDir, os.ModePerm))
		}
		ExitOnError(os.Chdir(subDir))
	}
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