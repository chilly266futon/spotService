package interceptors

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level       string `mapstructure:"level"`
	Development bool   `mapstructure:"development"`
	Encoding    string `mapstructure:"encoding"`
}

func New(cfg Config) (*zap.Logger, error) {
	var config zap.Config

	if cfg.Development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// Set encoding
	if config.Encoding != "" {
		config.Encoding = cfg.Encoding
	}

	return config.Build()
}

func NewDefault() (*zap.Logger, error) {
	return New(Config{
		Level:       "info",
		Development: false,
		Encoding:    "json",
	})
}

func NewDevelopment() (*zap.Logger, error) {
	return New(Config{
		Level:       "debug",
		Development: true,
		Encoding:    "console",
	})
}
