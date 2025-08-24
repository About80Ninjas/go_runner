// cmd/binary-executor/main.go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_runner/internal/api"
	"go_runner/internal/config"
	"go_runner/internal/executor"
	"go_runner/internal/repository"
	"go_runner/internal/storage"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize storage
	store := storage.NewFileStorage(cfg.Storage.Path)
	if err := store.Init(); err != nil {
		logger.Error("Failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize services
	gitManager := repository.NewGitManager(cfg.Storage.RepoPath)
	binaryExecutor := executor.NewExecutor(cfg.Storage.BinaryPath, cfg.Executor)

	// Initialize API server
	apiServer := api.NewServer(cfg.Server, store, gitManager, binaryExecutor)

	// Start server in goroutine
	go func() {
		logger.Info("Starting server",
			slog.String("host", cfg.Server.Host),
			slog.Int("port", cfg.Server.Port))

		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}

	logger.Info("Server exited")
}
