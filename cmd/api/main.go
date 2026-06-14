// Command api is the entrypoint for the BG-01 station backend.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brayangomez22/bg01-api/internal/config"
	"github.com/brayangomez22/bg01-api/internal/server"
	"github.com/brayangomez22/bg01-api/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		logger.Error("database unavailable", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := store.Migrate(db); err != nil {
		logger.Error("migrations failed", "err", err)
		os.Exit(1)
	}
	logger.Info("database ready", "path", cfg.DBPath)

	srv := server.New(cfg, logger, store.New(db))

	// Start the server in the background so we can listen for shutdown signals.
	go func() {
		logger.Info("station online", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for SIGINT/SIGTERM, then drain in-flight requests gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("station powering down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("station offline")
}
