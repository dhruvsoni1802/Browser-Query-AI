package main

import (
	"log/slog"  // library for structured logging
	"os"        // library for os related operations
	"os/signal" // library for signal handling such as Ctrl+C and kill signals
	"syscall"   // library for system call constants

	"github.com/dhruvsoni1802/browser-query-ai/internal/config"
)

// Main entry point of the program
func main() {

	// Setup the logger
	logger := InitializeLogger()
	slog.SetDefault(logger)

	// Load the configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("configuration loaded", "chromium_path", cfg.ChromiumPath, "server_port", cfg.ServerPort, "max_browsers", cfg.MaxBrowsers)

	// Create a channel to receive shutdown signals
	quit := make(chan os.Signal, 1)

	// Notify the channel for SIGINT and SIGTERM signals
	// Ctrl+C is SIGINT, kill signal is SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
	// Log the service is ready and awaiting shutdown signal
	slog.Info("Service ready", "status", "awaiting shutdown signal")
    
	// Wait for a shutdown signal
  sig := <-quit

	// Log the shutdown initiated with the signal
  slog.Info("shutdown initiated", "signal", sig.String())
    
    
	// Log the shutdown complete
  slog.Info("shutdown complete")
}

