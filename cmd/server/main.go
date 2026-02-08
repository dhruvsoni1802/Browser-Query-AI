package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dhruvsoni1802/browser-query-ai/internal/api"
	"github.com/dhruvsoni1802/browser-query-ai/internal/config"
	"github.com/dhruvsoni1802/browser-query-ai/internal/pool"
	"github.com/dhruvsoni1802/browser-query-ai/internal/session"
)

func main() {
	// Setup logger
	logger := InitializeLogger()
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("configuration loaded",
		"chromium_path", cfg.ChromiumPath,
		"server_port", cfg.ServerPort,
		"max_browsers", cfg.MaxBrowsers,
	)

	// Create process pool
	processPool, err := pool.NewProcessPool(cfg.ChromiumPath, cfg.MaxBrowsers)
	if err != nil {
		slog.Error("failed to create process pool", "error", err)
		os.Exit(1)
	}
	defer processPool.Shutdown()

	slog.Info("process pool created", "size", cfg.MaxBrowsers)

	// Create load balancer
	loadBalancer := pool.NewLoadBalancer(processPool)
	slog.Info("load balancer initialized")

	// Create session manager
	manager := session.NewManager()
	defer manager.Close()

	slog.Info("session manager initialized")

	// Create and start HTTP API server
	apiServer := api.NewServer(cfg.ServerPort, manager, loadBalancer)

	// Start HTTP server in goroutine
	go func() {
		if err := apiServer.Start(); err != nil {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("HTTP API server started", "port", cfg.ServerPort)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("service ready",
		"http_port", cfg.ServerPort,
		"browser_processes", cfg.MaxBrowsers,
		"status", "press Ctrl+C to shutdown",
	)

	// Wait for shutdown signal
	sig := <-quit
	slog.Info("shutdown initiated", "signal", sig.String())

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	// Close session manager
	if err := manager.Close(); err != nil {
		slog.Error("session manager close error", "error", err)
	}

	// Shutdown process pool
	if err := processPool.Shutdown(); err != nil {
		slog.Error("process pool shutdown error", "error", err)
	}

	slog.Info("shutdown complete")
}