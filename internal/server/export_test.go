package server

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

// /export is the build-time contract with the portfolio (src/types/domain.ts).
// These tests read the raw JSON, not the Go structs, because what matters is
// the wire shape the frontend parses: arrays must be [] (never null) and
// optional fields must be absent rather than null.

func TestExportEmptyDatabase(t *testing.T) {
	h := newTestHandler(t)
	got := getExport(t, h)

	if _, ok := got["pilot"].(map[string]any); !ok {
		t.Errorf("pilot = %#v, want the seeded object", got["pilot"])
	}
	for _, key := range []string{"missions", "technologies", "experience", "training", "frequencies"} {
		assertEmptyArray(t, key, got[key])
	}
	archive, _ := got["archive"].(map[string]any)
	assertEmptyArray(t, "archive.sections", archive["sections"])
	assertEmptyArray(t, "archive.records", archive["records"])
	if sc, ok := got["siteCopy"].(map[string]any); !ok || len(sc) != 0 {
		t.Errorf("siteCopy = %#v, want {}", got["siteCopy"])
	}
}

func TestExportDerivedFields(t *testing.T) {
	h := newTestHandler(t)
	auth := withCookie(login(t, h))

	seed := []struct{ path, body string }{
		{"/admin/technologies", `{"id":"go","name":"Go","order":1}`},
		{"/admin/technologies", `{"id":"react","name":"React","order":2}`},
		{"/admin/technologies", `{"id":"docker","name":"Docker","order":3}`},
		// Inserted out of order: usedInMissions must follow mission order, not
		// insertion order. "rust" has no technology row and must be ignored.
		{"/admin/missions", `{"id":"beta","code":"M-02","order":2,"technologies":["go","rust"]}`},
		{"/admin/missions", `{"id":"alpha","code":"M-01","order":1,"technologies":["react","go"]}`},
		{"/admin/site-copy", `{"key":"hero.title","value":"Hola"}`},
		{"/admin/site-copy", `{"key":"hero.subtitle","value":"Estación BG-01"}`},
	}
	for _, s := range seed {
		if rec := do(h, http.MethodPut, s.path, s.body, auth); rec.Code != http.StatusOK {
			t.Fatalf("seed %s %s: status %d, body %s", s.path, s.body, rec.Code, rec.Body)
		}
	}

	got := getExport(t, h)

	techs := byID(t, got["technologies"])
	tests := []struct {
		tech string
		want any // nil means the key must be absent
	}{
		{"go", []any{"alpha", "beta"}},
		{"react", []any{"alpha"}},
		{"docker", nil},
	}
	for _, tt := range tests {
		t.Run("usedInMissions/"+tt.tech, func(t *testing.T) {
			used, present := techs[tt.tech]["usedInMissions"]
			if tt.want == nil {
				if present {
					t.Errorf("usedInMissions = %#v, want key absent", used)
				}
				return
			}
			if !reflect.DeepEqual(used, tt.want) {
				t.Errorf("usedInMissions = %#v, want %#v", used, tt.want)
			}
		})
	}

	t.Run("mission arrays are never null", func(t *testing.T) {
		alpha := byID(t, got["missions"])["alpha"]
		for _, key := range []string{"highlights", "challenges"} {
			assertEmptyArray(t, "alpha."+key, alpha[key])
		}
	})

	t.Run("siteCopy is a key/value map", func(t *testing.T) {
		want := map[string]any{"hero.title": "Hola", "hero.subtitle": "Estación BG-01"}
		if !reflect.DeepEqual(got["siteCopy"], want) {
			t.Errorf("siteCopy = %#v, want %#v", got["siteCopy"], want)
		}
	})
}

func TestExportIsPublic(t *testing.T) {
	h := newTestHandler(t)
	if rec := do(h, http.MethodGet, "/export", ""); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d without a session", rec.Code, http.StatusOK)
	}
}

// getExport fetches /export and decodes it generically.
func getExport(t *testing.T, h http.Handler) map[string]any {
	t.Helper()
	rec := do(h, http.MethodGet, "/export", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("export: status %d, body %s", rec.Code, rec.Body)
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// byID indexes a decoded JSON array of objects by their "id" field.
func byID(t *testing.T, v any) map[string]map[string]any {
	t.Helper()
	list, ok := v.([]any)
	if !ok {
		t.Fatalf("want array, got %#v", v)
	}
	out := make(map[string]map[string]any, len(list))
	for _, item := range list {
		obj := item.(map[string]any)
		out[obj["id"].(string)] = obj
	}
	return out
}

func assertEmptyArray(t *testing.T, name string, v any) {
	t.Helper()
	if arr, ok := v.([]any); !ok || len(arr) != 0 {
		t.Errorf("%s = %#v, want []", name, v)
	}
}
