package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MisterVVP/logarift/llm-adapter/internal/adapter"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	cfg, err := adapter.LoadConfig()
	if err != nil {
		logger.Error("failed to load adapter configuration", "error", err)
		os.Exit(1)
	}
	svc := adapter.NewService(cfg, nil, logger)
	server := &http.Server{Addr: cfg.Address(), Handler: svc.Router(), ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		logger.Info("starting local llm adapter", "address", cfg.Address(), "model", cfg.Model, "runtime_url", cfg.RuntimeURL)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("adapter HTTP server failed", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("adapter HTTP server shutdown failed", "error", err)
		os.Exit(1)
	}
}
