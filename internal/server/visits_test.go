package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestPulseRateLimit(t *testing.T) {
	h := newTestHandler(t)

	// One IP rotating User-Agents: each would be a new "unique visitor", so
	// only the limiter keeps the flood out of the database.
	for i := range pulseHits + 5 {
		ua := func(r *http.Request) { r.Header.Set("User-Agent", "bot/"+strconv.Itoa(i)) }
		rec := do(h, http.MethodPost, "/pulse", "", fromIP("203.0.113.5"), ua)

		want := http.StatusNoContent
		if i >= pulseHits {
			want = http.StatusTooManyRequests
		}
		if rec.Code != want {
			t.Fatalf("pulse %d: status = %d, want %d", i+1, rec.Code, want)
		}
	}
	if rec := do(h, http.MethodPost, "/pulse", "", fromIP("198.51.100.7")); rec.Code != http.StatusNoContent {
		t.Fatalf("other IP: status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if got := statsToday(t, h); got != pulseHits+1 {
		t.Errorf("today = %d, want %d", got, pulseHits+1)
	}
}

func TestPulseIgnoresForwardedFor(t *testing.T) {
	h := newTestHandler(t)

	// Same client (Fly-Client-IP + UA) spoofing X-Forwarded-For must still
	// count once.
	for _, xff := range []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"} {
		xffOpt := func(r *http.Request) { r.Header.Set("X-Forwarded-For", xff) }
		if rec := do(h, http.MethodPost, "/pulse", "", fromIP("203.0.113.5"), xffOpt); rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	}
	if got := statsToday(t, h); got != 1 {
		t.Errorf("today = %d, want 1", got)
	}
}

// statsToday reads today's unique-visitor count from /admin/stats.
func statsToday(t *testing.T, h http.Handler) int64 {
	t.Helper()
	rec := do(h, http.MethodGet, "/admin/stats", "", withCookie(login(t, h)))
	if rec.Code != http.StatusOK {
		t.Fatalf("stats: status %d, body %s", rec.Code, rec.Body)
	}
	var stats visitStats
	if err := json.Unmarshal(rec.Body.Bytes(), &stats); err != nil {
		t.Fatal(err)
	}
	return stats.Today
}
