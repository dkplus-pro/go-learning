package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dkplus-pro/go-learning/examples/lesson09-config-logging-and-error-boundaries/internal/config"
	"github.com/dkplus-pro/go-learning/examples/lesson09-config-logging-and-error-boundaries/internal/task"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	level := new(slog.LevelVar)
	level.Set(cfg.LogLevel)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           task.NewHandler(task.NewService(), logger, cfg.RequestTimeout),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("server starting", "addr", cfg.Addr(), "request_timeout", cfg.RequestTimeout.String())
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
