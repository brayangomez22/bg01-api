package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// handlePublish triggers the portfolio's GitHub Actions deploy via a
// repository_dispatch event, so the static site rebuilds and republishes with
// the latest content from the database. The GitHub token stays server-side; the
// admin panel only ever calls this endpoint, never GitHub directly.
func (s *server) handlePublish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.GithubToken == "" || s.cfg.PublishRepo == "" {
			s.errorJSON(w, http.StatusServiceUnavailable, "publishing not configured")
			return
		}

		payload, _ := json.Marshal(map[string]string{"event_type": "content-updated"})
		url := fmt.Sprintf("https://api.github.com/repos/%s/dispatches", s.cfg.PublishRepo)

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			s.errorJSON(w, http.StatusInternalServerError, "build request failed")
			return
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+s.cfg.GithubToken)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s.logger.Error("publish dispatch failed", "err", err)
			s.errorJSON(w, http.StatusBadGateway, "could not reach GitHub")
			return
		}
		defer resp.Body.Close()

		// GitHub returns 204 No Content when the dispatch is accepted.
		if resp.StatusCode != http.StatusNoContent {
			s.logger.Error("publish dispatch rejected", "status", resp.StatusCode)
			s.errorJSON(w, http.StatusBadGateway, "GitHub rejected the dispatch")
			return
		}
		s.writeJSON(w, http.StatusAccepted, map[string]bool{"triggered": true})
	}
}
