package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	StatePath string     `hcl:"state_path"`
	Mqtt      MqttConfig `hcl:"mqtt,block"`
}

type MqttConfig struct {
	Enabled  bool   `hcl:"enabled"`
	Host     string `hcl:"host,optional"`
	Port     uint16 `hcl:"port,optional"`
	Username string `hcl:"username,optional"`
	Password string `hcl:"password,optional"`
}

func New(filepath string) (*Config, error) {
	config := new(Config)

	if err := hclsimple.DecodeFile(filepath, nil, config); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %s", err)
	}

	return config, nil
}
