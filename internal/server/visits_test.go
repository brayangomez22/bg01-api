package server

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/brayangomez22/bg01-api/internal/store"
)

func TestStationDay(t *testing.T) {
	tests := []struct {
		name string
		at   time.Time
		want string
	}{
		{"just before Bogotá midnight", time.Date(2026, 3, 10, 4, 59, 59, 0, time.UTC), "2026-03-09"},
		{"Bogotá midnight", time.Date(2026, 3, 10, 5, 0, 0, 0, time.UTC), "2026-03-10"},
		{"UTC midnight is still the previous day", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), "2026-03-09"},
		{"year boundary", time.Date(2027, 1, 1, 4, 0, 0, 0, time.UTC), "2026-12-31"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stationDay(tt.at); got != tt.want {
				t.Errorf("stationDay(%s) = %s, want %s", tt.at, got, tt.want)
			}
		})
	}
}

func TestStats(t *testing.T) {
	// Noon in Bogotá on 2026-03-10.
	today := time.Date(2026, 3, 10, 17, 0, 0, 0, time.UTC)
	clock := today
	h := newTestServer(t, func(s *server) { s.now = func() time.Time { return clock } }).handler()

	pulses := []struct {
		daysAgo int
		ip      string
	}{
		{8, "198.51.100.1"}, // outside the 7-day window: only in total
		{6, "198.51.100.1"}, // oldest day in the window
		{1, "198.51.100.1"}, // same visitor, new day: counts again
		{1, "198.51.100.2"},
		{0, "198.51.100.1"},
		{0, "198.51.100.1"}, // repeat within the day: deduplicated
	}
	for _, p := range pulses {
		clock = today.AddDate(0, 0, -p.daysAgo)
		if rec := do(h, http.MethodPost, "/pulse", "", fromIP(p.ip)); rec.Code != http.StatusNoContent {
			t.Fatalf("pulse: status %d", rec.Code)
		}
	}
	clock = today

	rec := do(h, http.MethodGet, "/admin/stats", "", withCookie(login(t, h)))
	if rec.Code != http.StatusOK {
		t.Fatalf("stats: status %d, body %s", rec.Code, rec.Body)
	}
	var got visitStats
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	want := visitStats{
		Today: 1, Yesterday: 2, Week: 4, Total: 5,
		Series: []store.DailyVisits{
			{Day: "2026-03-04", Count: 1},
			{Day: "2026-03-05", Count: 0},
			{Day: "2026-03-06", Count: 0},
			{Day: "2026-03-07", Count: 0},
			{Day: "2026-03-08", Count: 0},
			{Day: "2026-03-09", Count: 2},
			{Day: "2026-03-10", Count: 1},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("stats\n got: %+v\nwant: %+v", got, want)
	}
}

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
