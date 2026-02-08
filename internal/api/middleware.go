package api

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record start time
		startTime := time.Now()

		// Log request start
		slog.Info("request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
		)

		// Serve the request
		next.ServeHTTP(w, r)

		// Log request completion with duration
		slog.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(startTime),
		)
	})
}

// RecoveryMiddleware recovers from panics in the handlers
// A panic is like an exception. It stops normal execution of the handler and returns a 500 error to the client.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				//Logging error with stack trace
				slog.Error("panic in handler", "error", err, "stack", string(debug.Stack()))
				
				writeError(w, http.StatusInternalServerError, ErrCodeInternalError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}