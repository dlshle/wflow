package config

import (
	"context"
	"os"

	"github.com/dlshle/gommon/logging"
	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Database string `yaml:"database"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"database"`

	TCPPort  int `yaml:"tcp_port"`
	HTTPPort int `yaml:"http_port"`
}

func LoadConfig(path string) (ServerConfig, error) {
	var config ServerConfig
	err := loadYAML(path, &config)
	if err != nil {
		return ServerConfig{}, err
	}
	logging.GlobalLogger.Infof(context.Background(), "config %v loaded", config)
	return config, nil
}

func loadYAML(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(v)
}
