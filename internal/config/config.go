package config

import (
	"time"

	"github.com/chilly266futon/spotService/pkg/shared/config"
	"github.com/chilly266futon/spotService/pkg/shared/logger"
)

// Config конфигурация spot-service
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Logger    logger.Config   `yaml:"logger"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Health    HealthConfig    `yaml:"health"`
	Markets   []MarketConfig  `yaml:"markets"`
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// RateLimitConfig конфигурация rate limiting
type RateLimitConfig struct {
	Enabled           bool                             `yaml:"enabled"`
	RequestsPerSecond float64                          `yaml:"requests_per_second"`
	Burst             int                              `yaml:"burst"`
	Methods           map[string]MethodRateLimitConfig `yaml:"methods"`
}

// MethodRateLimitConfig лимит для конкретного метода
type MethodRateLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	Burst             int     `yaml:"burst"`
}

// HealthConfig конфигурация health checks
type HealthConfig struct {
	Enabled bool `yaml:"enabled"`
}

// MarketConfig конфигурация рынка
type MarketConfig struct {
	ID           string   `yaml:"id"`
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Enabled      bool     `yaml:"enabled"`
	AllowedRoles []string `yaml:"allowed_roles"`
}

// Load загружает конфигурацию из файла
func Load(configPath string) (*Config, error) {
	var cfg Config
	if err := config.Load(configPath, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// MustLoad загружает конфигурацию или паникует
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		panic(err)
	}
	return cfg
}
