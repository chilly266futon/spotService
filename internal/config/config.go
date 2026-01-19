package config

import (
	"time"

	"spotService/pkg/shared/config"
	"spotService/pkg/shared/logger"
)

// Config конфигурация spot-service
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Logger    logger.Config   `mapstructure:"logger"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	Health    HealthConfig    `mapstructure:"health"`
	Markets   []MarketConfig  `mapstructure:"markets"`
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// RateLimitConfig конфигурация rate limiting
type RateLimitConfig struct {
	Enabled           bool                             `mapstructure:"enabled"`
	RequestsPerSecond float64                          `mapstructure:"requests_per_second"`
	Burst             int                              `mapstructure:"burst"`
	Methods           map[string]MethodRateLimitConfig `mapstructure:"methods"`
}

// MethodRateLimitConfig лимит для конкретного метода
type MethodRateLimitConfig struct {
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	Burst             int     `mapstructure:"burst"`
}

// HealthConfig конфигурация health checks
type HealthConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// MarketConfig конфигурация рынка
type MarketConfig struct {
	ID           string   `mapstructure:"id"`
	Name         string   `mapstructure:"name"`
	Description  string   `mapstructure:"description"`
	Enabled      bool     `mapstructure:"enabled"`
	AllowedRoles []string `mapstructure:"allowed_roles"`
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
