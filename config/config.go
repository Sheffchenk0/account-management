package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App  App
	HTTP HTTP
	PG   PG
	RMQ  RMQ
}

type App struct {
	Name    string `env-required:"true" env:"APP_NAME"`
	Version string `env-required:"true" env:"APP_VERSION"`
}

type HTTP struct {
	Port string `env-required:"true" env:"HTTP_PORT"`
}

type PG struct {
	PoolMax int    `env-required:"true" env:"PG_POOL_MAX"`
	URL     string `env-required:"true" env:"PG_URL"`
}

type RMQ struct {
	Url          string `env-required:"true" env:"RMQ_URL"`
	ExchangeName string `env-required:"true" env:"RMQ_EXCHANGE_NAME"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
