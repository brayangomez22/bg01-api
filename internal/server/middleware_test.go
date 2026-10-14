package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brayangomez22/bg01-api/internal/config"
)

func TestCORS(t *testing.T) {
	const origin = "https://brayangomez.dev"

	tests := []struct {
		name        string
		allowed     string
		method      string
		status      int
		wantHeaders bool
		wantNext    bool
	}{
		{"configured GET", origin, http.MethodGet, http.StatusOK, true, true},
		{"configured preflight", origin, http.MethodOptions, http.StatusNoContent, true, false},
		{"unconfigured GET", "", http.MethodGet, http.StatusOK, false, true},
		{"unconfigured preflight", "", http.MethodOptions, http.StatusNoContent, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calledNext bool
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calledNext = true })
			h := cors(config.Config{AllowedOrigin: tt.allowed}, next)

			req := httptest.NewRequest(tt.method, "/admin/session", nil)
			req.Header.Set("Origin", origin)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			if calledNext != tt.wantNext {
				t.Errorf("next called = %v, want %v", calledNext, tt.wantNext)
			}

			want := map[string]string{
				"Access-Control-Allow-Origin":      origin,
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Allow-Methods":     "GET, PUT, POST, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type",
				"Vary":                             "Origin",
			}
			for k, v := range want {
				got := rec.Header().Get(k)
				if tt.wantHeaders && got != v {
					t.Errorf("%s = %q, want %q", k, got, v)
				}
				if !tt.wantHeaders && got != "" {
					t.Errorf("%s = %q, want it absent", k, got)
				}
			}
		})
	}
}
