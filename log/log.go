/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package log

import (
	"errors"
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

var (
	errLoggerAlreadyInitialized = errors.New("log: logger already initialized")
	errLoggerNotInitialized     = errors.New("log: logger not initialized")
)

const logChanBufferSize = 512

const projectFolder = "jackal"

var exitHandler = func() { os.Exit(-1) }

// singleton interface
var (
	inst unsafe.Pointer
)

// Level represents log level type.
type Level int

const (
	// DebugLevel represents DEBUG log level.
	DebugLevel Level = iota

	// InfoLevel represents INFO log level.
	InfoLevel

	// WarningLevel represents WARNING log level.
	WarningLevel

	// ErrorLevel represents ERROR log level.
	ErrorLevel

	// FatalLevel represents FATAL log level.
	FatalLevel
)

type Logger interface {
	Level() Level
	Log(format string, args []interface{}, pkg string, file string, line int, level Level, async bool)
	Close()
}

func Init(logger Logger) {
	if !atomic.CompareAndSwapPointer(&inst, unsafe.Pointer(nil), unsafe.Pointer(&logger)) {
		panic(errLoggerAlreadyInitialized)
	}
}

func Close() {
	ptr := atomic.SwapPointer(&inst, unsafe.Pointer(nil))
	if ptr == nil {
		panic(errLoggerNotInitialized)
	}
	(*(*Logger)(ptr)).Close()
}

// Debugf writes a 'debug' message to configured logger.
func Debugf(format string, args ...interface{}) {
	if inst := instance(); inst.Level() <= DebugLevel {
		ci := getCallerInfo()
		inst.Log(format, args, ci.pkg, ci.filename, ci.line, DebugLevel, true)
	}
}

// Infof writes a 'info' message to configured logger.
func Infof(format string, args ...interface{}) {
	if inst := instance(); inst.Level() <= InfoLevel {
		ci := getCallerInfo()
		inst.Log(format, args, ci.pkg, ci.filename, ci.line, InfoLevel, true)
	}
}

// Warnf writes a 'warning' message to configured logger.
func Warnf(format string, args ...interface{}) {
	if inst := instance(); inst.Level() <= WarningLevel {
		ci := getCallerInfo()
		inst.Log(format, args, ci.pkg, ci.filename, ci.line, WarningLevel, true)
	}
}

// Errorf writes an 'error' message to configured logger.
func Errorf(format string, args ...interface{}) {
	if inst := instance(); inst.Level() <= ErrorLevel {
		ci := getCallerInfo()
		inst.Log(format, args, ci.pkg, ci.filename, ci.line, ErrorLevel, true)
	}
}

// Fatalf writes a 'fatal' message to configured logger.
// Application should terminate after logging.
func Fatalf(format string, args ...interface{}) {
	ci := getCallerInfo()
	instance().Log(format, args, ci.pkg, ci.filename, ci.line, FatalLevel, false)
}

// Error writes an error value to configured logger.
func Error(err error) {
	if inst := instance(); inst.Level() <= ErrorLevel {
		ci := getCallerInfo()
		inst.Log("%v", []interface{}{err}, ci.pkg, ci.filename, ci.line, ErrorLevel, true)
	}
}

// Fatal writes an error value to configured logger.
// Application should terminate after logging.
func Fatal(err error) {
	ci := getCallerInfo()
	instance().Log("%v", []interface{}{err}, ci.pkg, ci.filename, ci.line, FatalLevel, false)
}

func instance() Logger {
	ptr := atomic.LoadPointer(&inst)
	if ptr == nil {
		panic(errLoggerNotInitialized)
	}
	return *(*Logger)(ptr)
}

type callerInfo struct {
	pkg      string
	filename string
	line     int
}

type record struct {
	level      Level
	pkg        string
	file       string
	line       int
	log        string
	continueCh chan struct{}
}

type logger struct {
	level   Level
	writers []io.Writer
	b       strings.Builder
	recCh   chan record
	closeCh chan bool
}

func New(level string, writers ...io.Writer) (Logger, error) {
	lvl, err := levelFromString(level)
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

func (l *logger) Level() Level {
	return l.level
}

func (l *logger) Log(format string, args []interface{}, pkg string, file string, line int, level Level, async bool) {
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

func (l *logger) Close() {
	close(l.closeCh)
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
				fmt.Fprintf(w, line)
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

func logLevelAbbreviation(level Level) string {
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

func logLevelGlyph(level Level) string {
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

func levelFromString(level string) (Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel, nil
	case "", "info":
		return InfoLevel, nil
	case "warning":
		return WarningLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	}
	return Level(-1), fmt.Errorf("log: unrecognized level: %s", level)
}
