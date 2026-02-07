package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// writeJSON writes a JSON success response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	// Set Content-Type header to tell client it's JSON
	w.Header().Set("Content-Type", "application/json")
	
	// Set HTTP status code (200, 201, etc.)
	w.WriteHeader(statusCode)
	
	// Encode data to JSON and write to response body
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
		return err
	}
	
	return nil
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	// Set Content-Type header
	w.Header().Set("Content-Type", "application/json")
	
	// Set HTTP status code (400, 404, 500, etc.)
	w.WriteHeader(statusCode)
	
	// Build error response
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	
	// Encode and write
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't even write the error response, log it
		slog.Error("failed to encode error response", "error", err)
	}
}