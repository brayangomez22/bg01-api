package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/brayangomez22/bg01-api/internal/domain"
)

// The generic resource handlers are exercised through /admin/technologies;
// every other content type shares the same code path.

const goTech = `{"id":"go","name":"Go","category":"language","proficiency":80,"since":"2023-01-01",` +
	`"description":"Backend.","planet":{"color":"#00ADD8","size":"md","orbit":2},"featured":true,"order":1}`

func TestResourceCRUD(t *testing.T) {
	h := newTestHandler(t)
	auth := withCookie(login(t, h))

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
		check  func(t *testing.T, body []byte)
	}{
		{"list starts empty", "GET", "/admin/technologies", "", http.StatusOK, wantTechs()},
		{"create", "PUT", "/admin/technologies", goTech, http.StatusOK, func(t *testing.T, b []byte) {
			var got domain.Technology
			mustDecode(t, b, &got)
			if got.ID != "go" || got.Planet.Color != "#00ADD8" || !got.Featured || got.Order != 1 {
				t.Errorf("created = %+v", got)
			}
		}},
		{"list has it", "GET", "/admin/technologies", "", http.StatusOK, wantTechs("go:80")},
		{"update same id", "PUT", "/admin/technologies", strings.Replace(goTech, `"proficiency":80`, `"proficiency":95`, 1), http.StatusOK, nil},
		{"list reflects update", "GET", "/admin/technologies", "", http.StatusOK, wantTechs("go:95")},
		{"malformed json", "PUT", "/admin/technologies", `{"id":`, http.StatusBadRequest, nil},
		{"unknown field", "PUT", "/admin/technologies", `{"id":"x","bogus":1}`, http.StatusBadRequest, nil},
		{"body over 1 MB", "PUT", "/admin/technologies", `{"description":"` + strings.Repeat("a", maxBodyBytes) + `"}`, http.StatusRequestEntityTooLarge, nil},
		{"delete", "DELETE", "/admin/technologies/go", "", http.StatusNoContent, nil},
		{"list is empty again", "GET", "/admin/technologies", "", http.StatusOK, wantTechs()},
	}
	// Steps share one database and run in order.
	for _, tt := range tests {
		rec := do(h, tt.method, tt.path, tt.body, auth)
		if rec.Code != tt.status {
			t.Fatalf("%s: status = %d, want %d (body %s)", tt.name, rec.Code, tt.status, rec.Body)
		}
		if tt.check != nil {
			t.Run(tt.name, func(t *testing.T) { tt.check(t, rec.Body.Bytes()) })
		}
	}
}

func TestResourceRequiresAuth(t *testing.T) {
	h := newTestHandler(t)
	tests := []struct {
		method, path, body string
	}{
		{"GET", "/admin/technologies", ""},
		{"PUT", "/admin/technologies", goTech},
		{"DELETE", "/admin/technologies/go", ""},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			if rec := do(h, tt.method, tt.path, tt.body); rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

// wantTechs checks a list response against "id:proficiency" pairs, in order.
func wantTechs(want ...string) func(t *testing.T, body []byte) {
	return func(t *testing.T, body []byte) {
		var list []domain.Technology
		mustDecode(t, body, &list)
		got := make([]string, len(list))
		for i, tech := range list {
			got[i] = tech.ID + ":" + strconv.Itoa(tech.Proficiency)
		}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("list = %v, want %v", got, want)
		}
	}
}

func mustDecode(t *testing.T, b []byte, dst any) {
	t.Helper()
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatalf("decode %s: %v", b, err)
	}
}

func TestResourceKeyValidation(t *testing.T) {
	h := newTestHandler(t)
	auth := withCookie(login(t, h))

	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"slug", "/admin/technologies", `{"id":"go-lang_2"}`, http.StatusOK},
		{"mixed-case repo name", "/admin/training", `{"repo":"TeamOne"}`, http.StatusOK},
		{"dotted site-copy key", "/admin/site-copy", `{"key":"hero.title","value":"Hi"}`, http.StatusOK},
		{"max length", "/admin/technologies", `{"id":"` + strings.Repeat("a", 100) + `"}`, http.StatusOK},
		{"too long", "/admin/technologies", `{"id":"` + strings.Repeat("a", 101) + `"}`, http.StatusBadRequest},
		{"space", "/admin/technologies", `{"id":"go lang"}`, http.StatusBadRequest},
		{"slash", "/admin/technologies", `{"id":"go/lang"}`, http.StatusBadRequest},
		{"leading dot", "/admin/technologies", `{"id":".hidden"}`, http.StatusBadRequest},
		{"non-ascii", "/admin/technologies", `{"id":"canción"}`, http.StatusBadRequest},
		// Every resource must reject a missing key (proves key is wired for each).
		{"empty technologies", "/admin/technologies", `{}`, http.StatusBadRequest},
		{"empty missions", "/admin/missions", `{}`, http.StatusBadRequest},
		{"empty experiences", "/admin/experiences", `{}`, http.StatusBadRequest},
		{"empty training", "/admin/training", `{}`, http.StatusBadRequest},
		{"empty archive sections", "/admin/archive/sections", `{}`, http.StatusBadRequest},
		{"empty archive records", "/admin/archive/records", `{}`, http.StatusBadRequest},
		{"empty frequencies", "/admin/frequencies", `{}`, http.StatusBadRequest},
		{"empty site-copy", "/admin/site-copy", `{}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(h, http.MethodPut, tt.path, tt.body, auth); rec.Code != tt.status {
				t.Errorf("status = %d, want %d (body %s)", rec.Code, tt.status, rec.Body)
			}
		})
	}
}
