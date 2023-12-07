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
	Scheduler struct {
		ExecutorPoolSize   int `yaml:"executor_pool_size,omitempty"`
		ExecutorMaxJobSize int `yaml:"executor_max_job_size,omitempty"`
	} `yaml:"scheduler,omitempty"`

	TCPPort      int `yaml:"tcp_port"`
	HTTPPort     int `yaml:"http_port"`
	HouseKeeping struct {
		CleanJobCron      string `yaml:"clean_job_cron"`
		KeepIntervalHours int    `yaml:"keep_interval_hours"`
	} `yaml:"house_keeping"`
}

func LoadConfig(path string) (ServerConfig, error) {
	var config ServerConfig
	err := loadYAML(path, &config)
	if err != nil {
		return ServerConfig{}, err
	}
	config.handleEmptyValues()
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

func (c *ServerConfig) handleEmptyValues() {
	if c.Scheduler.ExecutorPoolSize == 0 {
		c.Scheduler.ExecutorPoolSize = 256
	}
	if c.Scheduler.ExecutorMaxJobSize == 0 {
		c.Scheduler.ExecutorMaxJobSize = 512
	}
}
