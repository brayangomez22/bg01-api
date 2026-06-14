// Package server wires HTTP routing and middleware for the BG-01 API.
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/brayangomez22/bg01-api/internal/config"
)

// New builds the configured *http.Server. Routes are registered on the stdlib
// ServeMux (Go 1.22+ method-aware routing); a third-party router can be
// introduced later without touching the entrypoint.
func New(cfg config.Config, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	registerRoutes(mux, logger)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           logging(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// registerRoutes mounts every handler. New resource groups are added here as
// the API grows (missions, technologies, …).
func registerRoutes(mux *http.ServeMux, logger *slog.Logger) {
	mux.HandleFunc("GET /health", handleHealth(logger))
}
