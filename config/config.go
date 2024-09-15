package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		HTTP `yaml:"http"`
		Log  `yaml:"log"`
		PG   `yaml:"postgres"`
	}

	HTTP struct {
		Port    string `env-required:"true" env:"SERVER_PORT" env-upd:""`
		Address string `env-required:"true" yaml:"name" env:"SERVER_ADDRESS" env-upd:""`
	}

	Log struct {
		Level string `yaml:"level"`
	}

	PG struct {
		URL         string `env-required:"true" env:"POSTGRES_CONN" env-upd:""`
		MaxPoolSize int    `yaml:"max_pool_size"`
	}
)

func NewConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig(configPath, cfg)

	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	err = cleanenv.UpdateEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("error updating env: %w", err)
	}

	return cfg, nil
}
