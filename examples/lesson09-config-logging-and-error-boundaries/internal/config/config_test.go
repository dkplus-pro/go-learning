package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadFromEnvDefaults(t *testing.T) {
	cfg, err := LoadFromEnv(func(string) string { return "" })
	if err != nil {
		t.Fatalf("LoadFromEnv() unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.RequestTimeout != 2*time.Second {
		t.Fatalf("RequestTimeout = %s, want 2s", cfg.RequestTimeout)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("LogLevel = %v, want info", cfg.LogLevel)
	}
}

func TestLoadFromEnvCustomValues(t *testing.T) {
	values := map[string]string{
		"PORT":            "9090",
		"REQUEST_TIMEOUT": "1500ms",
		"LOG_LEVEL":       "debug",
	}

	cfg, err := LoadFromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("LoadFromEnv() unexpected error: %v", err)
	}

	if cfg.Addr() != ":9090" {
		t.Fatalf("Addr() = %q, want :9090", cfg.Addr())
	}
	if cfg.RequestTimeout != 1500*time.Millisecond {
		t.Fatalf("RequestTimeout = %s, want 1500ms", cfg.RequestTimeout)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("LogLevel = %v, want debug", cfg.LogLevel)
	}
}

func TestLoadFromEnvRejectsInvalidTimeout(t *testing.T) {
	_, err := LoadFromEnv(func(key string) string {
		if key == "REQUEST_TIMEOUT" {
			return "soon"
		}
		return ""
	})

	if err == nil {
		t.Fatal("LoadFromEnv() error = nil, want error")
	}
}

func TestLoadFromEnvRejectsInvalidLogLevel(t *testing.T) {
	_, err := LoadFromEnv(func(key string) string {
		if key == "LOG_LEVEL" {
			return "trace"
		}
		return ""
	})

	if err == nil {
		t.Fatal("LoadFromEnv() error = nil, want error")
	}
}
