/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

const logChanBufferSize = 512

const projectFolder = "jackal"

var exitHandler = func() { os.Exit(-1) }

// LogLevel represents logger level type.
type LogLevel int

const (
	// DebugLevel represents DEBUG logger level.
	DebugLevel LogLevel = iota

	// InfoLevel represents INFO logger level.
	InfoLevel

	// WarningLevel represents WARNING logger level.
	WarningLevel

	// ErrorLevel represents ERROR logger level.
	ErrorLevel

	// FatalLevel represents FATAL logger level.
	FatalLevel
)

// Logger interface is used to log specific application component messages.
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Error(err error)
	Fatal(err error)
	Close() error
}

type dummyLogger struct{}

func (_ *dummyLogger) Debugf(format string, args ...interface{}) {}
func (_ *dummyLogger) Infof(format string, args ...interface{})  {}
func (_ *dummyLogger) Warnf(format string, args ...interface{})  {}
func (_ *dummyLogger) Errorf(format string, args ...interface{}) {}
func (_ *dummyLogger) Fatalf(format string, args ...interface{}) {}
func (_ *dummyLogger) Error(err error)                           {}
func (_ *dummyLogger) Fatal(err error)                           {}
func (_ *dummyLogger) Close() error                              { return nil }

var (
	instance unsafe.Pointer
)

func init() {
	Set(&dummyLogger{})
}

func Set(logger Logger) {
	atomic.StorePointer(&instance, unsafe.Pointer(&logger))
}

func get() Logger {
	return *(*Logger)(atomic.LoadPointer(&instance))
}

// Debugf logs a 'debug' message to the logger file and echoes it to the console.
func Debugf(format string, args ...interface{}) { get().Debugf(format, args...) }

// Infof logs an 'info' message to the logger file and echoes it to the console.
func Infof(format string, args ...interface{}) { get().Debugf(format, args...) }

// Warnf logs a 'warning' message to the logger file and echoes it to the console.
func Warnf(format string, args ...interface{}) { get().Debugf(format, args...) }

// Errorf logs an 'error' message to the logger file and echoes it to the console.
func Errorf(format string, args ...interface{}) { get().Debugf(format, args...) }

// Fatalf logs a 'fatal' message to the logger file and echoes it to the console.
// Application should terminate after logging.
func Fatalf(format string, args ...interface{}) { get().Debugf(format, args...) }

// Error logs an error value to the logger file and echoes it to the console.
func Error(err error) { get().Error(err) }

// Fatal logs an error value to the logger file and echoes it to the console.
// Application should terminate after logging.
func Fatal(err error) { get().Fatal(err) }

func Close() error { return get().Close() }

type logger struct {
	level   LogLevel
	writers []io.Writer
	b       strings.Builder
	closed  uint32
	recCh   chan record
	closeCh chan bool
}

func New(level string, writers ...io.Writer) (Logger, error) {
	lvl, err := logLevelFromString(level)
	if err != nil {
		return nil, err
	}
	l := &logger{
		level:   lvl,
		writers: writers,
	}
	l.recCh = make(chan record, logChanBufferSize)
	l.closeCh = make(chan bool)
	go l.loop()
	return l, nil
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		ci := getCallerInfo()
		l.writeLog(ci.pkg, ci.filename, ci.line, format, DebugLevel, true, args...)
	}
}

func (l *logger) Infof(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		ci := getCallerInfo()
		l.writeLog(ci.pkg, ci.filename, ci.line, format, InfoLevel, true, args...)
	}
}

func (l *logger) Warnf(format string, args ...interface{}) {
	if l.level <= WarningLevel {
		ci := getCallerInfo()
		l.writeLog(ci.pkg, ci.filename, ci.line, format, WarningLevel, true, args...)
	}
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		ci := getCallerInfo()
		l.writeLog(ci.pkg, ci.filename, ci.line, format, ErrorLevel, true, args...)
	}
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	ci := getCallerInfo()
	l.writeLog(ci.pkg, ci.filename, ci.line, format, FatalLevel, false, args...)
}

func (l *logger) Error(err error) {
	if l.level <= ErrorLevel {
		ci := getCallerInfo()
		l.writeLog(ci.pkg, ci.filename, ci.line, "%v", ErrorLevel, true, err)
	}
}

func (l *logger) Fatal(err error) {
	ci := getCallerInfo()
	l.writeLog(ci.pkg, ci.filename, ci.line, "%v", FatalLevel, true, err)
}

func (l *logger) Close() error {
	if atomic.CompareAndSwapUint32(&l.closed, 0, 1) {
		close(l.closeCh)
	}
	return nil
}

type callerInfo struct {
	pkg      string
	filename string
	line     int
}

type record struct {
	level      LogLevel
	pkg        string
	file       string
	line       int
	log        string
	continueCh chan struct{}
}

func (l *logger) writeLog(pkg, file string, line int, format string, level LogLevel, async bool, args ...interface{}) {
	entry := record{
		level:      level,
		pkg:        pkg,
		file:       file,
		line:       line,
		log:        fmt.Sprintf(format, args...),
		continueCh: make(chan struct{}),
	}
	select {
	case l.recCh <- entry:
		if !async {
			<-entry.continueCh // wait until done
		}
	default:
		break // avoid blocking...
	}
}

func (l *logger) loop() {
	for {
		select {
		case rec := <-l.recCh:
			l.b.Reset()

			l.b.WriteString(time.Now().Format("2006-01-02 15:04:05"))
			l.b.WriteString(" ")
			l.b.WriteString(logLevelGlyph(rec.level))
			l.b.WriteString(" [")
			l.b.WriteString(logLevelAbbreviation(rec.level))
			l.b.WriteString("] ")

			l.b.WriteString(rec.pkg)
			if len(rec.pkg) > 0 {
				l.b.WriteString("/")
			}
			l.b.WriteString(rec.file)
			l.b.WriteString(":")
			l.b.WriteString(strconv.Itoa(rec.line))
			l.b.WriteString(" - ")
			l.b.WriteString(rec.log)
			l.b.WriteString("\n")

			line := l.b.String()
			for _, w := range l.writers {
				fmt.Fprint(w, line)
			}
			if rec.level == FatalLevel {
				exitHandler()
			}
			close(rec.continueCh)

		case <-l.closeCh:
			return
		}
	}
}

func getCallerInfo() callerInfo {
	ci := callerInfo{}
	_, file, ln, ok := runtime.Caller(2)
	if ok {
		ci.pkg = filepath.Base(path.Dir(file))
		if ci.pkg == projectFolder {
			ci.pkg = ""
		}
		filename := filepath.Base(file)
		ci.filename = strings.TrimSuffix(filename, filepath.Ext(filename))
		ci.line = ln
	} else {
		ci.filename = "???"
		ci.pkg = "???"
	}
	return ci
}

func logLevelAbbreviation(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DBG"
	case InfoLevel:
		return "INF"
	case WarningLevel:
		return "WRN"
	case ErrorLevel:
		return "ERR"
	case FatalLevel:
		return "FTL"
	default:
		// should not be reached
		return ""
	}
}

func logLevelGlyph(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "\U0001f50D"
	case InfoLevel:
		return "\u2139\ufe0f"
	case WarningLevel:
		return "\u26a0\ufe0f"
	case ErrorLevel:
		return "\U0001f4a5"
	case FatalLevel:
		return "\U0001f480"
	default:
		// should not be reached
		return ""
	}
}

func logLevelFromString(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel, nil
	case "", "info": // default logger level
		return InfoLevel, nil
	case "warning":
		return WarningLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return LogLevel(-1), fmt.Errorf("logger: unrecognized logger level: %s", level)
	}
}
