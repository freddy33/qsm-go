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

type Logger struct {
	log            *log.Logger
	level          LogLevel
	evaluateAssert bool
}

var p = message.NewPrinter(language.English)

func NewLogger(prefix string, level LogLevel) *Logger {
	return &Logger{log.New(os.Stdout, prefix+" ", log.LstdFlags|log.Lshortfile), level, level <= DEBUG}
}

func NewDataLogger(prefix string, level LogLevel) *Logger {
	return &Logger{log.New(os.Stdout, prefix+" ", 0), level, false}
}

func NewStatLogger(prefix string, level LogLevel) *Logger {
	return &Logger{log.New(os.Stdout, prefix+" ", log.Ltime|log.Lmicroseconds), level, false}
}

/***************************************************************/
// General Functions
/***************************************************************/

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

func (l *Logger) print(msg string) {
	err := l.log.Output(3, msg)
	if err != nil {
		log.Println(err)
	}
}

func (l *Logger) GetLevelName() string {
	return GetLevelName(l.level)
}

func (l *Logger) DoAssert() bool {
	return l.evaluateAssert
}

func (l *Logger) SetAssert(enable bool) {
	l.evaluateAssert = enable
}

// Trace Level

func (l *Logger) SetTrace() {
	l.level = TRACE
}

func (l *Logger) IsTrace() bool {
	return l.level <= TRACE
}

func (l *Logger) Trace(a ...interface{}) {
	if l.IsTrace() {
		l.print(makeMsg(TRACE, a...))
	}
}

func (l *Logger) Tracef(format string, v ...interface{}) {
	if l.IsTrace() {
		l.print(makeMsgFormat(TRACE, format, v...))
	}
}

// Debug Level

func (l *Logger) SetDebug() {
	l.level = DEBUG
}

func (l *Logger) IsDebug() bool {
	return l.level <= DEBUG
}

func (l *Logger) Debug(a ...interface{}) {
	if l.IsDebug() {
		l.print(makeMsg(DEBUG, a...))
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.IsDebug() {
		l.print(makeMsgFormat(DEBUG, format, v...))
	}
}

// Info Level

func (l *Logger) SetInfo() {
	l.level = INFO
}

func (l *Logger) IsInfo() bool {
	return l.level <= INFO
}

func (l *Logger) Info(a ...interface{}) {
	if l.IsInfo() {
		l.print(makeMsg(INFO, a...))
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.IsInfo() {
		l.print(makeMsgFormat(INFO, format, v...))
	}
}

// Warn Level

func (l *Logger) SetWarn() {
	l.level = WARN
}

func (l *Logger) IsWarn() bool {
	return l.level <= WARN
}

func (l *Logger) Warn(a ...interface{}) {
	if l.IsWarn() {
		l.print(makeMsg(WARN, a...))
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.IsWarn() {
		l.print(makeMsgFormat(WARN, format, v...))
	}
}

// Error Level

func (l *Logger) SetError() {
	l.level = ERROR
}

func (l *Logger) IsError() bool {
	// Always true
	return true
}

func (l *Logger) Error(a ...interface{}) {
	msg := makeMsg(ERROR, a...)
	l.print(msg)
	log.Print(msg)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	msg := makeMsgFormat(ERROR, format, v...)
	l.print(msg)
	log.Print(msg)
}

// Fatal panic out

func (l *Logger) Fatal(a ...interface{}) {
	msg := makeMsg(FATAL, a...)
	l.print(msg)
	panic(msg)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	msg := makeMsgFormat(FATAL, format, v...)
	l.print(msg)
	panic(msg)
}

