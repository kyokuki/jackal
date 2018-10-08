/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package logger

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testLogWriter struct {
	C chan string
}

func newTestLogWriter() *testLogWriter {
	return &testLogWriter{C: make(chan string)}
}

func (tw *testLogWriter) Write(p []byte) (int, error) {
	tw.C <- string(p)
	return len(p), nil
}

func TestDebugLog(t *testing.T) {
	Initialize(&Config{Level: DebugLevel})
	defer Shutdown()

	lw := newTestLogWriter()
	instance().outWriter = lw

	continueCh := make(chan struct{})

	Debugf("test debug logger!")
	go func() {
		select {
		case l := <-lw.C:
			require.True(t, strings.Contains(l, "[DBG]"))
			require.True(t, strings.Contains(l, "\U0001f50D"))
			require.True(t, strings.Contains(l, "test debug logger!"))

		case <-time.After(time.Millisecond * 200):
			require.Fail(t, "logger fetch timeout")
		}
		close(continueCh)
	}()
	<-continueCh
}

func TestInfoLog(t *testing.T) {
	Initialize(&Config{Level: InfoLevel})
	defer Shutdown()

	lw := newTestLogWriter()
	instance().outWriter = lw

	continueCh := make(chan struct{})

	Infof("test info logger!")
	go func() {
		select {
		case l := <-lw.C:
			require.True(t, strings.Contains(l, "[INF]"))
			require.True(t, strings.Contains(l, "\u2139\ufe0f"))
			require.True(t, strings.Contains(l, "test info logger!"))

		case <-time.After(time.Millisecond * 200):
			require.Fail(t, "logger fetch timeout")
		}
		close(continueCh)
	}()
	<-continueCh
}

func TestWarningLog(t *testing.T) {
	Initialize(&Config{Level: WarningLevel})
	defer Shutdown()

	lw := newTestLogWriter()
	instance().outWriter = lw

	continueCh := make(chan struct{})

	Warnf("test warning logger!")
	go func() {
		select {
		case l := <-lw.C:
			require.True(t, strings.Contains(l, "[WRN]"))
			require.True(t, strings.Contains(l, "\u26a0\ufe0f"))
			require.True(t, strings.Contains(l, "test warning logger!"))

		case <-time.After(time.Millisecond * 200):
			require.Fail(t, "logger fetch timeout")
		}
		close(continueCh)
	}()
	<-continueCh
}

func TestErrorLog(t *testing.T) {
	Initialize(&Config{Level: ErrorLevel})
	defer Shutdown()

	lw := newTestLogWriter()
	instance().outWriter = lw

	continueCh1 := make(chan struct{})

	Errorf("test error logger!")
	go func() {
		select {
		case l := <-lw.C:
			require.True(t, strings.Contains(l, "[ERR]"))
			require.True(t, strings.Contains(l, "\U0001f4a5"))
			require.True(t, strings.Contains(l, "test error logger!"))

		case <-time.After(time.Millisecond * 200):
			require.Fail(t, "logger fetch timeout")
		}
		close(continueCh1)
	}()
	<-continueCh1

	continueCh2 := make(chan struct{})
	err := errors.New("some error string")
	Error(err)
	go func() {
		select {
		case l := <-lw.C:
			require.True(t, strings.Contains(l, "some error string"))

		case <-time.After(time.Millisecond * 200):
			require.Fail(t, "logger fetch timeout")
		}
		close(continueCh2)
	}()
	<-continueCh2
}

func TestFatalLog(t *testing.T) {
	Initialize(&Config{Level: FatalLevel})
	defer Shutdown()

	lw := newTestLogWriter()
	instance().outWriter = lw
	exitHandler = func() {}

	continueCh := make(chan struct{})

	go func() {
		select {
		case l := <-lw.C:
			require.True(t, strings.Contains(l, "[FTL]"))
			require.True(t, strings.Contains(l, "\U0001f480"))
			require.True(t, strings.Contains(l, "test fatal logger!"))

		case <-time.After(time.Millisecond * 200):
			require.Fail(t, "logger fetch timeout")
		}
		close(continueCh)
	}()
	Fatalf("test fatal logger!")
	<-continueCh
}

func TestLogFile(t *testing.T) {
	logPath := "../testdata/log_file.logger"

	Initialize(&Config{Level: DebugLevel, LogPath: logPath})
	defer Shutdown()
	defer os.Remove(logPath)

	lw := newTestLogWriter()
	instance().outWriter = lw

	continueCh := make(chan struct{})

	Debugf("test file logger!")
	go func() {
		select {
		case <-lw.C:
			b, _ := ioutil.ReadFile(logPath)
			l := string(b)
			require.True(t, strings.Contains(l, "[DBG]"))
			require.True(t, strings.Contains(l, "\U0001f50D"))
			require.True(t, strings.Contains(l, "test file logger!"))

		case <-time.After(time.Millisecond * 200):
			require.Fail(t, "logger fetch timeout")
		}
		close(continueCh)
	}()
	<-continueCh
}
