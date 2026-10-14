package server

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	loginAttempts = 5
	loginWindow   = time.Minute

	// A real visitor fires one pulse per page load; this only stops floods
	// (e.g. rotating User-Agents to mint fake unique visitors).
	pulseHits   = 10
	pulseWindow = time.Minute
)

// rateLimiter is a fixed-window, per-key limiter: at most limit hits per
// window. State lives in memory, which is fine for a single-machine,
// single-admin API (a restart simply resets the counters).
type rateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	hits   map[string]*rateWindow
}

type rateWindow struct {
	start time.Time
	count int
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:  limit,
		window: window,
		now:    time.Now,
		hits:   make(map[string]*rateWindow),
	}
}

// allow records a hit for key. When the key is over its limit it returns false
// and how long until its window resets.
func (l *rateLimiter) allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	w, ok := l.hits[key]
	if !ok || now.Sub(w.start) >= l.window {
		l.hits[key] = &rateWindow{start: now, count: 1}
		return true, 0
	}
	if w.count >= l.limit {
		return false, w.start.Add(l.window).Sub(now)
	}
	w.count++
	return true, 0
}

// sweep drops expired windows so the map cannot grow without bound.
func (l *rateLimiter) sweep(now time.Time) {
	for k, w := range l.hits {
		if now.Sub(w.start) >= l.window {
			delete(l.hits, k)
		}
	}
}

// wrap guards next, answering 429 with Retry-After once the client IP is over
// the limit.
func (l *rateLimiter) wrap(s *server, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, retry := l.allow(clientIP(r))
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retry.Seconds()))))
			s.errorJSON(w, http.StatusTooManyRequests, "too many attempts")
			return
		}
		next(w, r)
	}
}

// clientIP returns the caller's IP. Behind Fly's proxy that is Fly-Client-IP
// (set by the edge, overwriting any client value); elsewhere it falls back to
// the TCP peer address.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("Fly-Client-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
