package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublish(t *testing.T) {
	const token = "ghp_test"

	tests := []struct {
		name         string
		githubStatus int  // status the fake GitHub answers with
		unreachable  bool // point at a closed server instead
		configured   bool
		status       int
		wantCall     bool
	}{
		{"dispatch accepted", http.StatusNoContent, false, true, http.StatusAccepted, true},
		{"bad token", http.StatusUnauthorized, false, true, http.StatusBadGateway, true},
		{"repo not found", http.StatusNotFound, false, true, http.StatusBadGateway, true},
		{"github unreachable", 0, true, true, http.StatusBadGateway, false},
		{"not configured", http.StatusNoContent, false, false, http.StatusServiceUnavailable, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			github := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				assertDispatch(t, r, token)
				w.WriteHeader(tt.githubStatus)
			}))
			defer github.Close()
			if tt.unreachable {
				github.Close()
			}

			s := newTestServer(t, func(s *server) {
				s.githubAPI = github.URL
				s.httpClient = github.Client()
				if tt.configured {
					s.cfg.PublishRepo = "brayangomez22/bg01-station"
					s.cfg.GithubToken = token
				}
			})
			h := s.handler()

			rec := do(h, http.MethodPost, "/admin/publish", "", withCookie(login(t, h)))
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d (body %s)", rec.Code, tt.status, rec.Body)
			}
			if called != tt.wantCall {
				t.Errorf("GitHub called = %v, want %v", called, tt.wantCall)
			}
		})
	}
}

func TestPublishRequiresAuth(t *testing.T) {
	h := newTestHandler(t)
	if rec := do(h, http.MethodPost, "/admin/publish", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// assertDispatch checks the request matches GitHub's repository_dispatch API.
func assertDispatch(t *testing.T, r *http.Request, token string) {
	t.Helper()
	if r.Method != http.MethodPost || r.URL.Path != "/repos/brayangomez22/bg01-station/dispatches" {
		t.Errorf("request = %s %s", r.Method, r.URL.Path)
	}
	headers := map[string]string{
		"Authorization":        "Bearer " + token,
		"Accept":               "application/vnd.github+json",
		"X-GitHub-Api-Version": "2022-11-28",
		"Content-Type":         "application/json",
	}
	for k, want := range headers {
		if got := r.Header.Get(k); got != want {
			t.Errorf("header %s = %q, want %q", k, got, want)
		}
	}
	body, _ := io.ReadAll(r.Body)
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil || payload["event_type"] != "content-updated" {
		t.Errorf("body = %s, want event_type content-updated", body)
	}
}
