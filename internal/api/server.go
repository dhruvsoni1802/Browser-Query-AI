package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dhruvsoni1802/browser-query-ai/internal/session"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Server represents the HTTP API server
type Server struct {
	router  *chi.Mux
	server  *http.Server
	manager *session.Manager
}

// NewServer creates a new HTTP server
// 
// NOTE: browserPort is temporary for single-browser setup.
// TODO: Replace with LoadBalancer when implementing process pool.
func NewServer(port string, manager *session.Manager, browserPort int) *Server {
	// Create chi router
	router := chi.NewRouter()

	// Add middleware (order matters!)
	router.Use(RecoveryMiddleware)   // 1. Catch panics
	router.Use(LoggingMiddleware)    // 2. Log requests
	router.Use(middleware.RequestID) // 3. Add request IDs

	// CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins (dev only)
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Cache preflight for 5 minutes
	}))

	// Create handlers
	handlers := NewHandlers(manager, browserPort)

	// Register routes
	router.Route("/sessions", func(r chi.Router) {
		// Session CRUD
		r.Post("/", handlers.CreateSession)   // POST /sessions
		r.Get("/", handlers.ListSessions)     // GET /sessions

		// Individual session routes
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.GetSession)       // GET /sessions/{id}
			r.Delete("/", handlers.DestroySession) // DELETE /sessions/{id}

			// Session operations (pageID in request body)
			r.Post("/navigate", handlers.Navigate)           // POST /sessions/{id}/navigate
			r.Post("/execute", handlers.ExecuteJS)           // POST /sessions/{id}/execute
			r.Post("/screenshot", handlers.CaptureScreenshot) // POST /sessions/{id}/screenshot

			// Page-specific routes (pageID in URL)
			r.Route("/pages/{pageId}", func(r chi.Router) {
				r.Get("/content", handlers.GetPageContent) // GET /sessions/{id}/pages/{pageId}/content
				r.Delete("/", handlers.ClosePage)          // DELETE /sessions/{id}/pages/{pageId}
			})
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		router:  router,
		server:  server,
		manager: manager,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	slog.Info("starting HTTP server", "addr", s.server.Addr)

	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down HTTP server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	slog.Info("HTTP server stopped")
	return nil
}