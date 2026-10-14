package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/brayangomez22/bg01-api/internal/domain"
)

// Each content type goes domain → params → SQLite → row → domain and must come
// back identical. Fixtures populate every field (assertFilled enforces it), so
// a field added to domain.go but forgotten in a mapper, a query or the schema
// makes this test fail instead of silently dropping data from /export.

func TestMapperRoundTrip(t *testing.T) {
	q := newTestStore(t)
	media := domain.MediaRef{Src: "/img/a.webp", Alt: "alt", Width: 640, Height: 480, Placeholder: "data:blur"}

	tests := []struct {
		name string
		run  func(t *testing.T) (in, out any)
	}{
		{"pilot", func(t *testing.T) (any, any) {
			in := domain.Pilot{
				Name: "Brayan", Callsign: "BG-01", Role: "Backend", Available: true, Location: "Bogotá",
				Bio: "bio", Manifesto: "manifesto", Stats: []domain.LabelledValue{{Label: "Años", Value: "5"}},
				Avatar: media, ResumeURL: "/cv.pdf",
			}
			return in, roundTrip(t, in, pilotFromDomain, q.UpdatePilot, pilotToDomain)
		}},
		{"technology", func(t *testing.T) (any, any) {
			in := domain.Technology{
				ID: "go", Name: "Go", Category: "language", Proficiency: 90, Since: "2021-01-01", Description: "d",
				Planet: domain.TechnologyPlanet{Color: "#00ADD8", Size: "m", Orbit: 2}, Featured: true, Order: 1,
			}
			return in, roundTrip(t, in, technologyFromDomain, q.UpsertTechnology, technologyToDomain)
		}},
		{"mission", func(t *testing.T) (any, any) {
			in := domain.Mission{
				ID: "m1", Code: "M-01", Title: "t", Summary: "s", Description: "d", Status: "completed", Role: "lead",
				DurationLabel: "6 meses", Period: domain.Period{Start: "2023-01", End: "2023-06"},
				Technologies: []string{"go"}, Highlights: []string{"h"}, Challenges: []string{"c"},
				Metrics: []domain.LabelledValue{{Label: "p99", Value: "40ms"}},
				Links:   domain.MissionLinks{Live: "https://x.dev", Repo: "https://github.com/x", CaseStudy: "/case"},
				Cover:   media, Gallery: []domain.MediaRef{media}, Featured: true, Order: 2,
			}
			return in, roundTrip(t, in, missionFromDomain, q.UpsertMission, missionToDomain)
		}},
		{"experience", func(t *testing.T) (any, any) {
			in := domain.Experience{
				ID: "log-1", Period: domain.Period{Start: "2022-01", End: "present"}, Company: "Co", Role: "Dev",
				Location: "Remote", Summary: "s", Responsibilities: []string{"r"}, Achievements: []string{"a"},
				Technologies: []string{"go"}, Type: "full-time", Order: 3,
			}
			return in, roundTrip(t, in, experienceFromDomain, q.UpsertExperience, experienceToDomain)
		}},
		{"training sim", func(t *testing.T) (any, any) {
			in := domain.TrainingSim{
				Repo: "TeamOne", Code: "TS-01", Title: "t", Summary: "s", Stack: []string{"go"}, Year: 2024,
				RepoURL: "https://github.com/x/TeamOne", Order: 4,
			}
			return in, roundTrip(t, in, trainingFromDomain, q.UpsertTrainingSim, trainingToDomain)
		}},
		{"archive section", func(t *testing.T) (any, any) {
			in := domain.ArchiveSection{ID: "ops", Code: "OPS", Label: "Operaciones", Order: 5}
			return in, roundTrip(t, in, archiveSectionFromDomain, q.UpsertArchiveSection, archiveSectionToDomain)
		}},
		{"archive record", func(t *testing.T) (any, any) {
			in := domain.ArchiveRecord{
				ID: "rec-1", Code: "AR-01", Title: "t", Abstract: "a", Section: "ops", Tags: []string{"go"},
				ArchivedAt: "2025-03-01", ReadingMinutes: 7,
				// One segment carrying every optional field keeps assertFilled honest.
				Body: []domain.ArchiveSegment{{Kind: "code", Text: "x", Items: []string{"i"}, Lang: "go", Code: "fmt.Println()"}},
				Refs: []string{"https://go.dev"}, Order: 6,
			}
			return in, roundTrip(t, in, archiveRecordFromDomain, q.UpsertArchiveRecord, archiveRecordToDomain)
		}},
		{"frequency", func(t *testing.T) (any, any) {
			in := domain.Frequency{ID: "github", Label: "GitHub", Handle: "@x", URL: "https://github.com/x", Icon: "github", Primary: true, Order: 7}
			return in, roundTrip(t, in, frequencyFromDomain, q.UpsertFrequency, frequencyToDomain)
		}},
		{"site copy", func(t *testing.T) (any, any) {
			in := domain.SiteCopyEntry{Key: "hero.title", Value: "Hola"}
			return in, roundTrip(t, in, siteCopyFromDomain, q.UpsertSiteCopy, siteCopyToDomain)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, out := tt.run(t)
			assertFilled(t, reflect.ValueOf(in), tt.name)
			if !reflect.DeepEqual(in, out) {
				t.Errorf("round trip changed the value\n got: %+v\nwant: %+v", out, in)
			}
		})
	}
}

// roundTrip saves in through its mapper and store method and maps it back.
func roundTrip[Dom, Params, Row any](
	t *testing.T,
	in Dom,
	fromDom func(Dom) (Params, error),
	save func(context.Context, Params) (Row, error),
	toDom func(Row) (Dom, error),
) Dom {
	t.Helper()
	params, err := fromDom(in)
	if err != nil {
		t.Fatalf("fromDomain: %v", err)
	}
	row, err := save(context.Background(), params)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	out, err := toDom(row)
	if err != nil {
		t.Fatalf("toDomain: %v", err)
	}
	return out
}

// derivedFields are computed at export time, never stored, so fixtures leave
// them empty.
var derivedFields = map[string]bool{"Technology.UsedInMissions": true}

// assertFilled fails if any field of v (recursively, including slice elements)
// is a zero value, so fixtures cannot fall behind the domain types.
func assertFilled(t *testing.T, v reflect.Value, path string) {
	t.Helper()
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			f := v.Type().Field(i)
			if derivedFields[v.Type().Name()+"."+f.Name] {
				continue
			}
			assertFilled(t, v.Field(i), path+"."+f.Name)
		}
	case reflect.Slice:
		if v.Len() == 0 {
			t.Errorf("fixture field %s is empty; populate it so the round trip covers it", path)
		}
		for i := range v.Len() {
			assertFilled(t, v.Index(i), path)
		}
	case reflect.Bool:
		// false is a legitimate value but would hide a dropped field.
		if !v.Bool() {
			t.Errorf("fixture field %s is false; set it to true so the round trip covers it", path)
		}
	default:
		if v.IsZero() {
			t.Errorf("fixture field %s is zero; populate it so the round trip covers it", path)
		}
	}
}
