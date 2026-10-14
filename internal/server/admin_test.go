package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/brayangomez22/bg01-api/internal/auth"
)

func TestRequireAuth(t *testing.T) {
	h := newTestHandler(t)
	secret := []byte(testSecret)

	tests := []struct {
		name   string
		opts   []func(*http.Request)
		status int
	}{
		{"no cookie", nil, http.StatusUnauthorized},
		{"garbage cookie", []func(*http.Request){withCookie("not-a-token")}, http.StatusUnauthorized},
		{"expired token", []func(*http.Request){withCookie(auth.NewToken(secret, -time.Minute))}, http.StatusUnauthorized},
		{"token from another secret", []func(*http.Request){withCookie(auth.NewToken([]byte("ffffffffffffffffffffffffffffffff"), time.Hour))}, http.StatusUnauthorized},
		{"token forged with empty secret", []func(*http.Request){withCookie(auth.NewToken(nil, time.Hour))}, http.StatusUnauthorized},
		{"valid token", []func(*http.Request){withCookie(auth.NewToken(secret, time.Hour))}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(h, http.MethodGet, "/admin/session", "", tt.opts...)
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d (body %s)", rec.Code, tt.status, rec.Body)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		status     int
		wantCookie bool
	}{
		{"correct password", `{"password":"` + testPassword + `"}`, http.StatusOK, true},
		{"wrong password", `{"password":"nope"}`, http.StatusUnauthorized, false},
		{"malformed json", `{"password":`, http.StatusBadRequest, false},
		{"unknown field", `{"password":"x","admin":true}`, http.StatusBadRequest, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(t)
			rec := do(h, http.MethodPost, "/admin/login", tt.body)
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.status, rec.Body)
			}
			var token string
			for _, c := range rec.Result().Cookies() {
				if c.Name == sessionCookie {
					token = c.Value
				}
			}
			if gotCookie := token != ""; gotCookie != tt.wantCookie {
				t.Fatalf("session cookie set = %v, want %v", gotCookie, tt.wantCookie)
			}
			if tt.wantCookie {
				if rec := do(h, http.MethodGet, "/admin/session", "", withCookie(token)); rec.Code != http.StatusOK {
					t.Errorf("issued cookie rejected: status %d", rec.Code)
				}
			}
		})
	}
}

func TestLoginRateLimit(t *testing.T) {
	h := newTestHandler(t)
	wrong := `{"password":"nope"}`

	tests := []struct {
		name   string
		ip     string
		body   string
		status int
	}{
		{"attempt 1", "203.0.113.1", wrong, http.StatusUnauthorized},
		{"attempt 2", "203.0.113.1", wrong, http.StatusUnauthorized},
		{"attempt 3", "203.0.113.1", wrong, http.StatusUnauthorized},
		{"attempt 4", "203.0.113.1", wrong, http.StatusUnauthorized},
		{"attempt 5", "203.0.113.1", wrong, http.StatusUnauthorized},
		{"attempt 6 is throttled", "203.0.113.1", wrong, http.StatusTooManyRequests},
		{"even the right password is throttled", "203.0.113.1", `{"password":"` + testPassword + `"}`, http.StatusTooManyRequests},
		{"another IP is unaffected", "198.51.100.7", `{"password":"` + testPassword + `"}`, http.StatusOK},
	}
	// Steps share one handler and run in order: the limiter state carries over.
	for _, tt := range tests {
		rec := do(h, http.MethodPost, "/admin/login", tt.body, fromIP(tt.ip))
		if rec.Code != tt.status {
			t.Fatalf("%s: status = %d, want %d (body %s)", tt.name, rec.Code, tt.status, rec.Body)
		}
		if tt.status == http.StatusTooManyRequests && rec.Header().Get("Retry-After") == "" {
			t.Errorf("%s: missing Retry-After header", tt.name)
		}
	}
}
