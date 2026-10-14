package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brayangomez22/bg01-api/internal/domain"
)

const pilotBody = `{"name":"Brayan","callsign":"BG-01","role":"Backend","available":false,"location":"Bogotá",` +
	`"bio":"b","manifesto":"m","stats":[{"label":"Años","value":"5"}],` +
	`"avatar":{"src":"/a.webp","alt":"a","width":64,"height":64},"resumeUrl":"/cv.pdf"}`

func TestPilot(t *testing.T) {
	h := newTestHandler(t)
	auth := withCookie(login(t, h))

	tests := []struct {
		name   string
		method string
		body   string
		status int
		check  func(t *testing.T, p domain.Pilot)
	}{
		{"seeded singleton", "GET", "", http.StatusOK, func(t *testing.T, p domain.Pilot) {
			if p.Callsign != "BG-01" {
				t.Errorf("callsign = %q, want seeded BG-01", p.Callsign)
			}
		}},
		{"update", "PUT", pilotBody, http.StatusOK, assertPilot},
		{"update persisted", "GET", "", http.StatusOK, assertPilot},
		{"unknown field", "PUT", `{"name":"x","admin":true}`, http.StatusBadRequest, nil},
		{"still unchanged after bad update", "GET", "", http.StatusOK, assertPilot},
	}
	// Steps share one database and run in order.
	for _, tt := range tests {
		rec := do(h, tt.method, "/admin/pilot", tt.body, auth)
		if rec.Code != tt.status {
			t.Fatalf("%s: status = %d, want %d (body %s)", tt.name, rec.Code, tt.status, rec.Body)
		}
		if tt.check != nil {
			var p domain.Pilot
			mustDecode(t, rec.Body.Bytes(), &p)
			t.Run(tt.name, func(t *testing.T) { tt.check(t, p) })
		}
	}
}

func TestPilotRequiresAuth(t *testing.T) {
	h := newTestHandler(t)
	for _, method := range []string{"GET", "PUT"} {
		if rec := do(h, method, "/admin/pilot", pilotBody); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want %d", method, rec.Code, http.StatusUnauthorized)
		}
	}
}

func assertPilot(t *testing.T, got domain.Pilot) {
	t.Helper()
	var want domain.Pilot
	if err := json.Unmarshal([]byte(pilotBody), &want); err != nil {
		t.Fatal(err)
	}
	if got.Name != want.Name || got.Available != want.Available || len(got.Stats) != 1 ||
		got.Avatar != want.Avatar || got.ResumeURL != want.ResumeURL {
		t.Errorf("pilot = %+v, want %+v", got, want)
	}
}
