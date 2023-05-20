package config

import (
	"errors"
	"fmt"
	"os"

	jsonIter "github.com/json-iterator/go"
)

const (
	configPath = "CONFIG_PATH"
)

var errEnvNotSet = errors.New("env not set")

type Config struct {
	ServerOpts ServerOpts `json:"server_opts"`
	MainPort   string     `json:"main_port"`
}

type ServerOpts struct {
	ReadTimeout          int `json:"read_timeout"`
	WriteTimeout         int `json:"write_timeout"`
	IdleTimeout          int `json:"idle_timeout"`
	MaxRequestBodySizeMb int `json:"max_request_body_size_mb"`
}

func ParseConfig() (*Config, error) {
	path := os.Getenv(configPath)
	if path == "" {
		return nil, fmt.Errorf("%w: %s", errEnvNotSet, configPath)
	}

	fileBody, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read config file: %w", err)
	}

	var cfg Config

	err = jsonIter.Unmarshal(fileBody, &cfg)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal config: %w", err)
	}

	return &cfg, nil
}
