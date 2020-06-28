package m3util

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log"
	"os"
)

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

type BaseLogger struct {
	log             *log.Logger
	level           LogLevel
	evaluateAssert  bool
	ignoreNextError bool
}

var p = message.NewPrinter(language.English)

func NewLogger(prefix string, level LogLevel) Logger {
	return &BaseLogger{log.New(os.Stdout, prefix+" ", log.LstdFlags|log.Lshortfile), level, level <= DEBUG, false}
}

func NewDataLogger(prefix string, level LogLevel) Logger {
	return &BaseLogger{log.New(os.Stdout, prefix+" ", 0), level, false, false}
}

func NewStatLogger(prefix string, level LogLevel) Logger {
	return &BaseLogger{log.New(os.Stdout, prefix+" ", log.Ltime|log.Lmicroseconds), level, false, false}
}

/***************************************************************/
// General Functions
/***************************************************************/

func ReadVerbose() {
	if len(os.Args) > 1 {
		for i := 1; i<len(os.Args); i++ {
			if os.Args[i] == "-v" {
				// Make all logger debug level
			}
		}
	}
}

func GetLevelName(level LogLevel) string {
	switch level {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "ERROR"
	}
	return fmt.Sprintf("UNK%d", level)
}

func makeMsg(level LogLevel, a ...interface{}) string {
	return p.Sprintln(append([]interface{}{GetLevelName(level)}, a...))
}

func makeMsgFormat(level LogLevel, format string, v ...interface{}) string {
	return p.Sprintln(append([]interface{}{GetLevelName(level)}, p.Sprintf(format, v...)))
}

/***************************************************************/
// Logger Functions
/***************************************************************/

func (l *BaseLogger) print(msg string) {
	err := l.log.Output(3, msg)
	if err != nil {
		log.Println(err)
	}
}

func (l *BaseLogger) GetLevelName() string {
	return GetLevelName(l.level)
}

func (l *BaseLogger) DoAssert() bool {
	return l.evaluateAssert
}

func (l *BaseLogger) SetAssert(enable bool) {
	l.evaluateAssert = enable
}

// Trace Level

func (l *BaseLogger) SetTrace() {
	l.level = TRACE
}

func (l *BaseLogger) IsTrace() bool {
	return l.level <= TRACE
}

func (l *BaseLogger) Trace(a ...interface{}) {
	if l.IsTrace() {
		l.print(makeMsg(TRACE, a...))
	}
}

func (l *BaseLogger) Tracef(format string, v ...interface{}) {
	if l.IsTrace() {
		l.print(makeMsgFormat(TRACE, format, v...))
	}
}

// Debug Level

func (l *BaseLogger) SetDebug() {
	l.level = DEBUG
}

func (l *BaseLogger) IsDebug() bool {
	return l.level <= DEBUG
}

func (l *BaseLogger) Debug(a ...interface{}) {
	if l.IsDebug() {
		l.print(makeMsg(DEBUG, a...))
	}
}

func (l *BaseLogger) Debugf(format string, v ...interface{}) {
	if l.IsDebug() {
		l.print(makeMsgFormat(DEBUG, format, v...))
	}
}

// Info Level

func (l *BaseLogger) SetInfo() {
	l.level = INFO
}

func (l *BaseLogger) IsInfo() bool {
	return l.level <= INFO
}

func (l *BaseLogger) Info(a ...interface{}) {
	if l.IsInfo() {
		l.print(makeMsg(INFO, a...))
	}
}

func (l *BaseLogger) Infof(format string, v ...interface{}) {
	if l.IsInfo() {
		l.print(makeMsgFormat(INFO, format, v...))
	}
}

// Warn Level

func (l *BaseLogger) SetWarn() {
	l.level = WARN
}

func (l *BaseLogger) IsWarn() bool {
	return l.level <= WARN
}

func (l *BaseLogger) Warn(a ...interface{}) {
	if l.IsWarn() {
		l.print(makeMsg(WARN, a...))
	}
}

func (l *BaseLogger) Warnf(format string, v ...interface{}) {
	if l.IsWarn() {
		l.print(makeMsgFormat(WARN, format, v...))
	}
}

// Error Level

func (l *BaseLogger) SetError() {
	l.level = ERROR
}

func (l *BaseLogger) IsError() bool {
	// Always true
	return true
}

func (l *BaseLogger) Error(a ...interface{}) {
	if l.ignoreNextError {
		l.ignoreNextError = false
		return
	}
	msg := makeMsg(ERROR, a...)
	l.print(msg)
	log.Print(msg)
}

func (l *BaseLogger) Errorf(format string, v ...interface{}) {
	if l.ignoreNextError {
		l.ignoreNextError = false
		return
	}
	msg := makeMsgFormat(ERROR, format, v...)
	l.print(msg)
	log.Print(msg)
}

// Fatal panic out
func (l *BaseLogger) Fatal(a ...interface{}) {
	msg := makeMsg(FATAL, a...)
	l.print(msg)
	panic(msg)
}

func (l *BaseLogger) Fatalf(format string, v ...interface{}) {
	msg := makeMsgFormat(FATAL, format, v...)
	l.print(msg)
	panic(msg)
}

func (l *BaseLogger) IgnoreNextError() {
	l.ignoreNextError = true
}

type Logger interface {
	GetLevelName() string

	DoAssert() bool
	SetAssert(enable bool)

	SetTrace()
	IsTrace() bool
	Trace(a ...interface{})
	Tracef(format string, v ...interface{})

	SetDebug()
	IsDebug() bool
	Debug(a ...interface{})
	Debugf(format string, v ...interface{})

	SetInfo()
	IsInfo() bool
	Info(a ...interface{})
	Infof(format string, v ...interface{})

	SetWarn()
	IsWarn() bool
	Warn(a ...interface{})
	Warnf(format string, v ...interface{})

	SetError()
	IsError() bool
	Error(a ...interface{})
	Errorf(format string, v ...interface{})

	Fatal(a ...interface{})
	Fatalf(format string, v ...interface{})

	IgnoreNextError()
}
