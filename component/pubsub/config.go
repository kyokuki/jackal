package pubsub

import (
	"errors"
)

type Config struct {
	Host string
	Mysql string
}

type configProxy struct {
	Host string `yaml:"host"`
	Mysql string `yaml:"mysql"`
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

	// mandatory fields
	if len(p.Mysql) == 0 {
		return errors.New("pubsub.Mysql: Mysql value must be set")
	}

	c.Host = p.Host
	c.Mysql = p.Mysql

	return nil
}
