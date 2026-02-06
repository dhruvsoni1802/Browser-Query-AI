package main

import (
	"log/slog"
	"os"
	"time"
)

// Function to initialize the logger
func InitializeLogger() *slog.Logger {
	var handler slog.Handler

	if os.Getenv("ENV") == "production" {

		// Initialize JSON handler for production environment
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ Level: slog.LevelInfo })
	} else {

		// Initialize Text handler for development environment with better formatting
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ 
			Level: slog.LevelDebug,
			AddSource: false,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Format timestamp to be more readable
				if a.Key == slog.TimeKey {
					t := a.Value.Time()
					return slog.String("time", t.Format(time.DateTime))
				}
				return a
			},
		})
	}

	// Create a new logger with the initialized handler
	return slog.New(handler)
}