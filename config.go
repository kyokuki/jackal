/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package main

import (
	"bytes"
	"io/ioutil"

	"github.com/ortuman/jackal/c2s"
	"github.com/ortuman/jackal/component"
	"github.com/ortuman/jackal/module"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/s2s"
	"github.com/ortuman/jackal/storage"
	"gopkg.in/yaml.v2"
)

// debugConfig represents debug server configuration.
type debugConfig struct {
	Port int `yaml:"port"`
}

type loggerConfig struct {
	Level   string `yaml:"level"`
	LogPath string `yaml:"log_path"`
}

// config represents a global configuration.
type config struct {
	PIDFile    string           `yaml:"pid_path"`
	Debug      debugConfig      `yaml:"debug"`
	Logger     loggerConfig     `yaml:"logger"`
	Storage    storage.Config   `yaml:"storage"`
	Router     router.Config    `yaml:"router"`
	Modules    module.Config    `yaml:"modules"`
	Components component.Config `yaml:"components"`
	C2S        []c2s.Config     `yaml:"c2s"`
	S2S        *s2s.Config      `yaml:"s2s"`
}

// FromFile loads default global configuration from
// a specified file.
func (cfg *config) FromFile(configFile string) error {
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, cfg)
}

// FromBuffer loads default global configuration from
// a specified byte buffer.
func (cfg *config) FromBuffer(buf *bytes.Buffer) error {
	return yaml.Unmarshal(buf.Bytes(), cfg)
}
