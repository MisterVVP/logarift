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

	"github.com/MisterVVP/logarift/backend/internal/config"
	"github.com/MisterVVP/logarift/backend/internal/database"
	"github.com/MisterVVP/logarift/backend/internal/httpserver"
	"github.com/MisterVVP/logarift/backend/internal/store/mongostore"
	"github.com/MisterVVP/logarift/backend/internal/version"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), cfg.MongoDBConnectTimeout)
	db, err := database.ConnectWithRetry(connectCtx, cfg, 500*time.Millisecond)
	cancelConnect()
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	bootstrapCtx, cancelBootstrap := context.WithTimeout(context.Background(), cfg.MongoDBConnectTimeout)
	if err := db.EnsureIndexes(bootstrapCtx); err != nil {
		cancelBootstrap()
		slog.Error("failed to ensure MongoDB indexes", "error", err)
		os.Exit(1)
	}
	stores := mongostore.New(db)
	if err := mongostore.EnsureDefaultModelConfig(bootstrapCtx, stores.ModelConfigs); err != nil {
		cancelBootstrap()
		slog.Error("failed to ensure default model config", "error", err)
		os.Exit(1)
	}
	cancelBootstrap()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := db.Close(shutdownCtx); err != nil {
			slog.Error("failed to close MongoDB connection", "error", err)
		}
	}()

	api := httpserver.New(cfg, db, version.Current())
	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("starting Logarift API", "address", cfg.Address(), "database", cfg.MongoDBDatabase)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
			stop()
		}
	}()

	<-serverCtx.Done()
	slog.Info("shutting down Logarift API")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown failed", "error", err)
		os.Exit(1)
	}
}
