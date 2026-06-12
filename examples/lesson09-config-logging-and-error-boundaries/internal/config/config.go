package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port           string
	RequestTimeout time.Duration
	LogLevel       slog.Level
}

func (c Config) Addr() string {
	return ":" + c.Port
}

func Load() (Config, error) {
	return LoadFromEnv(os.Getenv)
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	port := valueOrDefault(getenv("PORT"), "8080")

	timeout, err := time.ParseDuration(valueOrDefault(getenv("REQUEST_TIMEOUT"), "2s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse REQUEST_TIMEOUT: %w", err)
	}
	if timeout <= 0 {
		return Config{}, fmt.Errorf("REQUEST_TIMEOUT must be positive")
	}

	level, err := parseLogLevel(valueOrDefault(getenv("LOG_LEVEL"), "info"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:           port,
		RequestTimeout: timeout,
		LogLevel:       level,
	}, nil
}

func valueOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported LOG_LEVEL %q", value)
	}
}
