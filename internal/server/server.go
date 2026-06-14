// Package server wires HTTP routing and middleware for the BG-01 API.
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/brayangomez22/bg01-api/internal/config"
	"github.com/brayangomez22/bg01-api/internal/store"
)

// server holds shared dependencies for HTTP handlers.
type server struct {
	logger *slog.Logger
	q      *store.Queries
}

// New builds the configured *http.Server. Routes are registered on the stdlib
// ServeMux (Go 1.22+ method-aware routing); a third-party router can be
// introduced later without touching the entrypoint.
func New(cfg config.Config, logger *slog.Logger, q *store.Queries) *http.Server {
	s := &server{logger: logger, q: q}

	mux := http.NewServeMux()
	s.routes(mux)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           logging(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// routes mounts every handler. New resource groups (missions, technologies, …)
// are added here as the API grows.
func (s *server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.handleHealth())
}
