package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	Webhooks      []string `yaml:"webhooks"`
	WebSocketPort int      `yaml:"websocket_port"`
	HTTPPort      int      `yaml:"http_port"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
