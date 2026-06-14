package server

import (
	"fmt"

	"github.com/brayangomez22/bg01-api/internal/domain"
	"github.com/brayangomez22/bg01-api/internal/store"
)

// These mappers convert between the flat store rows (JSON columns as strings,
// bools as int64) and the typed domain contract. Read mappers (…ToDomain) may
// fail on corrupt JSON columns; write mappers (…FromDomain) may fail if a
// nested field can't be marshalled.

// --- Pilot ---

func pilotToDomain(p store.Pilot) (domain.Pilot, error) {
	d := domain.Pilot{
		Name:      p.Name,
		Callsign:  p.Callsign,
		Role:      p.Role,
		Available: i2b(p.Available),
		Location:  p.Location,
		Bio:       p.Bio,
		Manifesto: p.Manifesto,
		ResumeURL: p.ResumeUrl,
	}
	if err := decodeCol(p.Stats, &d.Stats); err != nil {
		return d, fmt.Errorf("pilot stats: %w", err)
	}
	if err := decodeCol(p.Avatar, &d.Avatar); err != nil {
		return d, fmt.Errorf("pilot avatar: %w", err)
	}
	return d, nil
}

func pilotFromDomain(d domain.Pilot) (store.UpdatePilotParams, error) {
	var err error
	js := jsonEncoder(&err)
	p := store.UpdatePilotParams{
		Name:      d.Name,
		Callsign:  d.Callsign,
		Role:      d.Role,
		Available: b2i(d.Available),
		Location:  d.Location,
		Bio:       d.Bio,
		Manifesto: d.Manifesto,
		Stats:     js(d.Stats, "[]"),
		Avatar:    js(d.Avatar, "{}"),
		ResumeUrl: d.ResumeURL,
	}
	return p, err
}

// --- Technology ---

func technologyToDomain(t store.Technology) (domain.Technology, error) {
	d := domain.Technology{
		ID:          t.ID,
		Name:        t.Name,
		Category:    t.Category,
		Proficiency: int(t.Proficiency),
		Since:       t.Since,
		Description: t.Description,
		Featured:    i2b(t.Featured),
		Order:       int(t.SortOrder),
	}
	if err := decodeCol(t.Planet, &d.Planet); err != nil {
		return d, fmt.Errorf("technology %s planet: %w", t.ID, err)
	}
	return d, nil
}

func technologyFromDomain(d domain.Technology) (store.UpsertTechnologyParams, error) {
	var err error
	js := jsonEncoder(&err)
	return store.UpsertTechnologyParams{
		ID:          d.ID,
		Name:        d.Name,
		Category:    d.Category,
		Proficiency: int64(d.Proficiency),
		Since:       d.Since,
		Description: d.Description,
		Planet:      js(d.Planet, "{}"),
		Featured:    b2i(d.Featured),
		SortOrder:   int64(d.Order),
	}, err
}

// --- Mission ---

func missionToDomain(m store.Mission) (domain.Mission, error) {
	d := domain.Mission{
		ID:            m.ID,
		Code:          m.Code,
		Title:         m.Title,
		Summary:       m.Summary,
		Description:   m.Description,
		Status:        m.Status,
		Role:          m.Role,
		DurationLabel: m.DurationLabel,
		Period:        domain.Period{Start: m.PeriodStart, End: nullToStr(m.PeriodEnd)},
		Featured:      i2b(m.Featured),
		Order:         int(m.SortOrder),
	}
	for _, f := range []struct {
		name string
		raw  string
		dst  any
	}{
		{"technologies", m.Technologies, &d.Technologies},
		{"highlights", m.Highlights, &d.Highlights},
		{"challenges", m.Challenges, &d.Challenges},
		{"metrics", m.Metrics, &d.Metrics},
		{"links", m.Links, &d.Links},
		{"cover", m.Cover, &d.Cover},
		{"gallery", m.Gallery, &d.Gallery},
	} {
		if err := decodeCol(f.raw, f.dst); err != nil {
			return d, fmt.Errorf("mission %s %s: %w", m.ID, f.name, err)
		}
	}
	return d, nil
}

func missionFromDomain(d domain.Mission) (store.UpsertMissionParams, error) {
	var err error
	js := jsonEncoder(&err)
	return store.UpsertMissionParams{
		ID:            d.ID,
		Code:          d.Code,
		Title:         d.Title,
		Summary:       d.Summary,
		Description:   d.Description,
		Status:        d.Status,
		Role:          d.Role,
		DurationLabel: d.DurationLabel,
		PeriodStart:   d.Period.Start,
		PeriodEnd:     nullStr(d.Period.End),
		Technologies:  js(d.Technologies, "[]"),
		Highlights:    js(d.Highlights, "[]"),
		Challenges:    js(d.Challenges, "[]"),
		Metrics:       js(d.Metrics, "[]"),
		Links:         js(d.Links, "{}"),
		Cover:         js(d.Cover, "{}"),
		Gallery:       js(d.Gallery, "[]"),
		Featured:      b2i(d.Featured),
		SortOrder:     int64(d.Order),
	}, err
}

// --- Experience ---

func experienceToDomain(e store.Experience) (domain.Experience, error) {
	d := domain.Experience{
		ID:       e.ID,
		Period:   domain.Period{Start: e.PeriodStart, End: e.PeriodEnd},
		Company:  e.Company,
		Role:     e.Role,
		Location: e.Location,
		Summary:  e.Summary,
		Type:     e.Type,
		Order:    int(e.SortOrder),
	}
	for _, f := range []struct {
		name string
		raw  string
		dst  any
	}{
		{"responsibilities", e.Responsibilities, &d.Responsibilities},
		{"achievements", e.Achievements, &d.Achievements},
		{"technologies", e.Technologies, &d.Technologies},
	} {
		if err := decodeCol(f.raw, f.dst); err != nil {
			return d, fmt.Errorf("experience %s %s: %w", e.ID, f.name, err)
		}
	}
	return d, nil
}

func experienceFromDomain(d domain.Experience) (store.UpsertExperienceParams, error) {
	var err error
	js := jsonEncoder(&err)
	return store.UpsertExperienceParams{
		ID:               d.ID,
		PeriodStart:      d.Period.Start,
		PeriodEnd:        d.Period.End,
		Company:          d.Company,
		Role:             d.Role,
		Location:         d.Location,
		Summary:          d.Summary,
		Responsibilities: js(d.Responsibilities, "[]"),
		Achievements:     js(d.Achievements, "[]"),
		Technologies:     js(d.Technologies, "[]"),
		Type:             d.Type,
		SortOrder:        int64(d.Order),
	}, err
}

// --- TrainingSim ---

func trainingToDomain(t store.TrainingSim) (domain.TrainingSim, error) {
	d := domain.TrainingSim{
		Repo:    t.Repo,
		Code:    t.Code,
		Title:   t.Title,
		Summary: t.Summary,
		Year:    int(t.Year),
		RepoURL: t.RepoUrl,
		Order:   int(t.SortOrder),
	}
	if err := decodeCol(t.Stack, &d.Stack); err != nil {
		return d, fmt.Errorf("training %s stack: %w", t.Repo, err)
	}
	return d, nil
}

func trainingFromDomain(d domain.TrainingSim) (store.UpsertTrainingSimParams, error) {
	var err error
	js := jsonEncoder(&err)
	return store.UpsertTrainingSimParams{
		Repo:      d.Repo,
		Code:      d.Code,
		Title:     d.Title,
		Summary:   d.Summary,
		Stack:     js(d.Stack, "[]"),
		Year:      int64(d.Year),
		RepoUrl:   d.RepoURL,
		SortOrder: int64(d.Order),
	}, err
}

// --- ArchiveSection ---

func archiveSectionToDomain(s store.ArchiveSection) (domain.ArchiveSection, error) {
	return domain.ArchiveSection{
		ID:    s.ID,
		Code:  s.Code,
		Label: s.Label,
		Order: int(s.SortOrder),
	}, nil
}

func archiveSectionFromDomain(d domain.ArchiveSection) (store.UpsertArchiveSectionParams, error) {
	return store.UpsertArchiveSectionParams{
		ID:        d.ID,
		Code:      d.Code,
		Label:     d.Label,
		SortOrder: int64(d.Order),
	}, nil
}

// --- ArchiveRecord ---

func archiveRecordToDomain(a store.ArchiveRecord) (domain.ArchiveRecord, error) {
	d := domain.ArchiveRecord{
		ID:             a.ID,
		Code:           a.Code,
		Title:          a.Title,
		Abstract:       a.Abstract,
		Section:        a.Section,
		ArchivedAt:     a.ArchivedAt,
		ReadingMinutes: int(a.ReadingMinutes),
		Order:          int(a.SortOrder),
	}
	for _, f := range []struct {
		name string
		raw  string
		dst  any
	}{
		{"tags", a.Tags, &d.Tags},
		{"body", a.Body, &d.Body},
		{"refs", a.Refs, &d.Refs},
	} {
		if err := decodeCol(f.raw, f.dst); err != nil {
			return d, fmt.Errorf("archive %s %s: %w", a.ID, f.name, err)
		}
	}
	return d, nil
}

func archiveRecordFromDomain(d domain.ArchiveRecord) (store.UpsertArchiveRecordParams, error) {
	var err error
	js := jsonEncoder(&err)
	return store.UpsertArchiveRecordParams{
		ID:             d.ID,
		Code:           d.Code,
		Title:          d.Title,
		Abstract:       d.Abstract,
		Section:        d.Section,
		Tags:           js(d.Tags, "[]"),
		ArchivedAt:     d.ArchivedAt,
		ReadingMinutes: int64(d.ReadingMinutes),
		Body:           js(d.Body, "[]"),
		Refs:           js(d.Refs, "[]"),
		SortOrder:      int64(d.Order),
	}, err
}

// --- Frequency ---

func frequencyToDomain(f store.Frequency) (domain.Frequency, error) {
	return domain.Frequency{
		ID:      f.ID,
		Label:   f.Label,
		Handle:  f.Handle,
		URL:     f.Url,
		Icon:    f.Icon,
		Primary: i2b(f.IsPrimary),
		Order:   int(f.SortOrder),
	}, nil
}

func frequencyFromDomain(d domain.Frequency) (store.UpsertFrequencyParams, error) {
	return store.UpsertFrequencyParams{
		ID:        d.ID,
		Label:     d.Label,
		Handle:    d.Handle,
		Url:       d.URL,
		Icon:      d.Icon,
		IsPrimary: b2i(d.Primary),
		SortOrder: int64(d.Order),
	}, nil
}

// --- SiteCopy ---

func siteCopyToDomain(c store.SiteCopy) (domain.SiteCopyEntry, error) {
	return domain.SiteCopyEntry{Key: c.Key, Value: c.Value}, nil
}

func siteCopyFromDomain(d domain.SiteCopyEntry) (store.UpsertSiteCopyParams, error) {
	return store.UpsertSiteCopyParams{Key: d.Key, Value: d.Value}, nil
}

// jsonEncoder returns a closure that marshals values to JSON strings, latching
// the first error into *err so a chain of calls inside a struct literal stays
// readable. Once an error occurs, subsequent calls return "".
func jsonEncoder(err *error) func(v any, empty string) string {
	return func(v any, empty string) string {
		if *err != nil {
			return ""
		}
		s, e := jsonOr(v, empty)
		if e != nil {
			*err = e
		}
		return s
	}
}
