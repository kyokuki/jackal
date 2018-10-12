package pubsub

import (
	"errors"
)

type Config struct {
	Host string
}

type configProxy struct {
	Host string `yaml:"host"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := configProxy{}
	if err := unmarshal(&p); err != nil {
		return err
	}
	// mandatory fields
	if len(p.Host) == 0 {
		return errors.New("pubsub.Config: host value must be set")
	}

	c.Host = p.Host

	return nil
}
