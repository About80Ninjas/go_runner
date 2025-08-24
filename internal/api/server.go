// internal/api/server.go
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go_runner/internal/config"
	"go_runner/internal/executor"
	"go_runner/internal/repository"
	"go_runner/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Server represents the API server
type Server struct {
	config   config.ServerConfig
	router   *chi.Mux
	server   *http.Server
	storage  storage.Storage
	git      *repository.GitManager
	executor *executor.Executor
}

// NewServer creates a new API server
func NewServer(cfg config.ServerConfig, storage storage.Storage, git *repository.GitManager, exec *executor.Executor) *Server {
	s := &Server{
		config:   cfg,
		storage:  storage,
		git:      git,
		executor: exec,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public endpoints
		r.Get("/health", s.healthHandler)
		r.Get("/docs", s.swaggerUIHandler)
		r.Get("/openapi.json", s.openAPIHandler)

		// Binary management endpoints (requires auth)
		r.Route("/binaries", func(r chi.Router) {
			r.Use(s.authMiddleware)
			r.Get("/", s.listBinariesHandler)
			r.Post("/", s.createBinaryHandler)
			r.Get("/{id}", s.getBinaryHandler)
			r.Put("/{id}", s.updateBinaryHandler)
			r.Delete("/{id}", s.deleteBinaryHandler)
			r.Post("/{id}/build", s.buildBinaryHandler)
		})

		// Execution endpoints
		r.Route("/execute", func(r chi.Router) {
			r.Use(s.apiKeyMiddleware)
			r.Post("/", s.executeBinaryHandler)
			r.Get("/{id}", s.getExecutionHandler)
			r.Delete("/{id}", s.stopExecutionHandler)
		})
	})

	// Admin UI
	r.Route("/admin", func(r chi.Router) {
		r.Use(s.authMiddleware)
		r.Get("/*", s.adminUIHandler)
	})

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	s.router = r
	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}
}

// Start starts the server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
