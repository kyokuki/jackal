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
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/ortuman/jackal/c2s"
	"github.com/ortuman/jackal/component"
	"github.com/ortuman/jackal/log"
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
	var cfg config
	if err := cfg.FromFile(configFile); err != nil {
		logError(err)
		return
	}
	// initialize logger
	var logFiles []io.WriteCloser
	if len(cfg.Logger.LogPath) > 0 {
		// create logFile intermediate directories.
		if err := os.MkdirAll(filepath.Dir(cfg.Logger.LogPath), os.ModePerm); err != nil {
			logError(err)
			return
		}
		f, err := os.OpenFile(cfg.Logger.LogPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			logError(err)
			return
		}
		logFiles = append(logFiles, f)
	}
	logger, err := log.New(cfg.Logger.Level, os.Stdout, logFiles...)
	if err != nil {
		logError(err)
	}
	log.Set(logger)
	defer log.Unset()

	s, err := storage.New(&cfg.Storage)
	if err != nil {
		log.Fatal(err)
	}
	storage.Set(s)
	defer storage.Unset()

	// initialize router
	r, err := router.New(&cfg.Router)
	if err != nil {
		log.Fatal(err)
	}

	// initialize modules & components...
	mods := module.New(&cfg.Modules, r)
	defer mods.Close()

	comps := component.New(&cfg.Components, mods.DiscoInfo)
	defer comps.Close()

	// create PID file
	if err := createPIDFile(cfg.PIDFile); err != nil {
		log.Warnf("%v", err)
	}
	// start serving...
	for i := range logoStr {
		log.Infof("%s", logoStr[i])
	}
	log.Infof("")
	log.Infof("jackal %v\n", version.ApplicationVersion)

	// initialize debug server...
	if cfg.Debug.Port > 0 {
		go initDebugServer(cfg.Debug.Port)
	}

	// start serving s2s...
	s2s := s2s.New(cfg.S2S, mods, r)
	if s2s.Enabled() {
		r.SetS2SOutProvider(s2s)

		s2s.Start()
		defer s2s.Stop()

	} else {
		log.Infof("s2s disabled")
	}

	// start serving c2s...
	c2s, err := c2s.New(cfg.C2S, mods, comps, r)
	if err != nil {
		log.Fatal(err)
	}
	c2s.Start()
	defer c2s.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

var debugSrv *http.Server

func initDebugServer(port int) {
	debugSrv = &http.Server{}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("%v", err)
	}
	debugSrv.Serve(ln)
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
	defer file.Close()

	currentPid := os.Getpid()
	if _, err := file.WriteString(strconv.FormatInt(int64(currentPid), 10)); err != nil {
		return err
	}
	return nil
}

func logError(err error) {
	fmt.Fprintf(os.Stderr, "jackal: %v\n", err)
}
