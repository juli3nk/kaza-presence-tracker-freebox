package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	Mqtt MqttConfig `hcl:"mqtt,block"`
}

type MqttConfig struct {
	Host string `hcl:"host"`
	Port uint16 `hcl:"port"`
	Username string `hcl:"username"`
	Password string `hcl:"password"`
}

func NewConfig(filepath string) (*Config, error) {
	config := new(Config)

	if err := hclsimple.DecodeFile(filepath, nil, config); err != nil {
		return nil, fmt.Errorf("Failed to load configuration: %s", err)
	}

	return config, nil
}
