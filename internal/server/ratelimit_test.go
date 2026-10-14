package server

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l := newRateLimiter(2, time.Minute)
	l.now = func() time.Time { return clock }

	tests := []struct {
		name      string
		advance   time.Duration
		key       string
		want      bool
		wantRetry time.Duration
	}{
		{"first hit", 0, "a", true, 0},
		{"second hit", 10 * time.Second, "a", true, 0},
		{"over limit", 10 * time.Second, "a", false, 40 * time.Second},
		{"other key has its own budget", 0, "b", true, 0},
		{"still blocked just before reset", 39 * time.Second, "a", false, time.Second},
		{"window resets", time.Second, "a", true, 0},
		{"later hit sweeps b's expired window", 20 * time.Second, "a", true, 0},
	}
	// Steps run in order against one limiter and a shared fake clock.
	for _, tt := range tests {
		clock = clock.Add(tt.advance)
		got, retry := l.allow(tt.key)
		if got != tt.want || retry != tt.wantRetry {
			t.Fatalf("%s: allow() = (%v, %v), want (%v, %v)", tt.name, got, retry, tt.want, tt.wantRetry)
		}
	}
	if _, ok := l.hits["b"]; ok {
		t.Error("expired window for key b was not swept")
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		flyIP      string
		remoteAddr string
		want       string
	}{
		{"fly header wins", "203.0.113.9", "10.0.0.1:5555", "203.0.113.9"},
		{"falls back to remote addr", "", "10.0.0.1:5555", "10.0.0.1"},
		{"ipv6 remote addr", "", "[2001:db8::1]:443", "2001:db8::1"},
		{"remote addr without port", "", "10.0.0.1", "10.0.0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/admin/login", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.flyIP != "" {
				r.Header.Set("Fly-Client-IP", tt.flyIP)
			}
			if got := clientIP(r); got != tt.want {
				t.Errorf("clientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
