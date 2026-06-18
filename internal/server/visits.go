package server

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"time"

	"github.com/brayangomez22/bg01-api/internal/store"
)

// bogota pins visits to Colombia's local day (fixed UTC-5, no DST), so "today"
// in the control center matches the operator's day rather than UTC midnight.
// A fixed zone avoids depending on tzdata being present in the container.
var bogota = time.FixedZone("America/Bogota", -5*60*60)

// stationDay formats t as the station-local calendar day, 'YYYY-MM-DD'.
func stationDay(t time.Time) string {
	return t.In(bogota).Format("2006-01-02")
}

// visitorHash derives a privacy-preserving identifier for a request. It folds
// the day and the server's session secret in as salt, so the result is
// irreversible (no IP recoverable), holds no PII at rest, and rotates every day
// — unique-visitor counting without cookies or any durable identifier (the same
// approach Plausible uses).
func (s *server) visitorHash(r *http.Request, day string) string {
	ip := r.Header.Get("Fly-Client-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		// Strip the ephemeral port so the same client maps to one hash; without
		// this, RemoteAddr's per-connection port would defeat per-day dedupe.
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}
	h := sha256.New()
	h.Write(s.secret)
	h.Write([]byte(day))
	h.Write([]byte(ip))
	h.Write([]byte(r.UserAgent()))
	return hex.EncodeToString(h.Sum(nil))[:32]
}

// handlePulse records an anonymous visit. It is public — the static portfolio
// fires it as a fire-and-forget beacon on load — so it returns 204 with no body
// and swallows every error: analytics must never disturb a page load.
func (s *server) handlePulse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		day := stationDay(time.Now())
		if err := s.q.RecordVisit(r.Context(), day, s.visitorHash(r, day)); err != nil {
			s.logger.Error("record visit failed", "err", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// visitStats is the control center's traffic summary.
type visitStats struct {
	Today     int64               `json:"today"`
	Yesterday int64               `json:"yesterday"`
	Week      int64               `json:"week"`   // sum of the 7-day window
	Total     int64               `json:"total"`  // all-time unique visitor-days
	Series    []store.DailyVisits `json:"series"` // last 7 days, oldest→newest, gaps filled with 0
}

// handleStats returns the unique-visitor summary for the control dashboard:
// today, yesterday, the trailing 7-day total, the all-time total, and a gap-free
// 7-day series for the mini chart.
func (s *server) handleStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		// The 7-day window ending today, as station-local day strings.
		days := make([]string, 7)
		for i := range days {
			days[i] = stationDay(now.AddDate(0, 0, i-6))
		}

		rows, err := s.q.VisitsSince(r.Context(), days[0])
		if err != nil {
			s.logger.Error("visits since failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "db error")
			return
		}
		byDay := make(map[string]int64, len(rows))
		for _, d := range rows {
			byDay[d.Day] = d.Count
		}

		series := make([]store.DailyVisits, 7)
		var week int64
		for i, day := range days {
			c := byDay[day]
			series[i] = store.DailyVisits{Day: day, Count: c}
			week += c
		}

		total, err := s.q.VisitsTotal(r.Context())
		if err != nil {
			s.logger.Error("visits total failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "db error")
			return
		}

		s.writeJSON(w, http.StatusOK, visitStats{
			Today:     series[6].Count,
			Yesterday: series[5].Count,
			Week:      week,
			Total:     total,
			Series:    series,
		})
	}
}
