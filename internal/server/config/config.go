package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
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
