/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ortuman/jackal/c2s"
	"github.com/ortuman/jackal/component"
	"github.com/ortuman/jackal/host"
	"github.com/ortuman/jackal/logger"
	"github.com/ortuman/jackal/module"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/s2s"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/version"
)

var logoStr = []string{
	`        __               __            __   `,
	`       |__|____    ____ |  | _______  |  |  `,
	`       |  \__  \ _/ ___\|  |/ /\__  \ |  |  `,
	`       |  |/ __ \\  \___|    <  / __ \|  |__`,
	`   /\__|  (____  /\___  >__|_ \(____  /____/`,
	`   \______|    \/     \/     \/     \/      `,
}

const usageStr = `
Usage: jackal [options]

Server Options:
    -c, --config <file>    Configuration file path
Common Options:
    -h, --help             Show this message
    -v, --version          Show version
`

func main() {
	var configFile string
	var showVersion bool
	var showUsage bool

	flag.BoolVar(&showUsage, "help", false, "Show this message")
	flag.BoolVar(&showUsage, "h", false, "Show this message")
	flag.BoolVar(&showVersion, "version", false, "Print version information.")
	flag.BoolVar(&showVersion, "v", false, "Print version information.")
	flag.StringVar(&configFile, "config", "/etc/jackal/jackal.yml", "Configuration file path.")
	flag.StringVar(&configFile, "c", "/etc/jackal/jackal.yml", "Configuration file path.")
	flag.Usage = func() {
		for i := range logoStr {
			fmt.Fprintf(os.Stdout, "%s\n", logoStr[i])
		}
		fmt.Fprintf(os.Stdout, "%s\n", usageStr)
	}
	flag.Parse()

	// print usage
	if showUsage {
		flag.Usage()
		return
	}

	// print version
	if showVersion {
		fmt.Fprintf(os.Stdout, "jackal version: %v\n", version.ApplicationVersion)
		return
	}
	// load configuration
	var cfg Config
	if err := cfg.FromFile(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "jackal: %v\n", err)
		return
	}
	// initialize subsystems... (order matters)
	initLogger(&cfg.Logger)

	storage.Initialize(&cfg.Storage)

	host.Initialize(cfg.Hosts)

	router.Initialize(&router.Config{GetS2SOut: s2s.GetS2SOut})

	// initialize modules & components...
	module.Initialize(&cfg.Modules)
	component.Initialize(&cfg.Components)

	// create PID file
	if err := createPIDFile(cfg.PIDFile); err != nil {
		logger.Warnf("%v", err)
	}
	// start serving...
	for i := range logoStr {
		logger.Infof("%s", logoStr[i])
	}
	logger.Infof("")
	logger.Infof("jackal %v\n", version.ApplicationVersion)

	// initialize debug server...
	if cfg.Debug.Port > 0 {
		go initDebugServer(cfg.Debug.Port)
	}

	// start serving s2s...
	s2s.Initialize(cfg.S2S)

	// start serving c2s...
	c2s.Initialize(cfg.C2S)
}

var debugSrv *http.Server

func initDebugServer(port int) {
	debugSrv = &http.Server{}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatalf("%v", err)
	}
	debugSrv.Serve(ln)
}

func initLogger(cfg *LoggerConfig) error {
	writers := []io.Writer{os.Stdout}
	if len(cfg.LogPath) > 0 {
		// create logFile intermediate directories.
		if err := os.MkdirAll(filepath.Dir(cfg.LogPath), os.ModePerm); err != nil {
			return err
		}
		f, err := os.OpenFile(cfg.LogPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		writers = append(writers, f)
	}
	l, err := logger.New(cfg.Level, writers...)
	if err != nil {
		return err
	}
	logger.Set(l)
	return nil
}

func createPIDFile(pidFile string) error {
	if len(pidFile) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(pidFile), os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(pidFile)
	if err != nil {
		return err
	}
	currentPid := os.Getpid()
	if _, err := file.WriteString(strconv.FormatInt(int64(currentPid), 10)); err != nil {
		return err
	}
	return nil
}
