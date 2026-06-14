package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// handleHealth is a liveness probe used by the host platform and uptime checks.
func handleHealth(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			logger.Error("health encode failed", "err", err)
		}
	}
}
