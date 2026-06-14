package server

import (
	"context"
	"net/http"

	"github.com/brayangomez22/bg01-api/internal/domain"
)

// handleExport assembles and returns the full content bundle (the build-time
// contract). It is public: the portfolio's GitHub Action fetches it at build.
func (s *server) handleExport() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bundle, err := s.buildExport(r.Context())
		if err != nil {
			s.logger.Error("export failed", "err", err)
			s.errorJSON(w, http.StatusInternalServerError, "export failed")
			return
		}
		s.writeJSON(w, http.StatusOK, bundle)
	}
}

func (s *server) buildExport(ctx context.Context) (domain.Export, error) {
	var out domain.Export

	pilotRow, err := s.q.GetPilot(ctx)
	if err != nil {
		return out, err
	}
	if out.Pilot, err = pilotToDomain(pilotRow); err != nil {
		return out, err
	}

	if out.Missions, err = collect(ctx, s.q.ListMissions, missionToDomain); err != nil {
		return out, err
	}
	if out.Technologies, err = collect(ctx, s.q.ListTechnologies, technologyToDomain); err != nil {
		return out, err
	}
	if out.Experience, err = collect(ctx, s.q.ListExperiences, experienceToDomain); err != nil {
		return out, err
	}
	if out.Training, err = collect(ctx, s.q.ListTrainingSims, trainingToDomain); err != nil {
		return out, err
	}
	if out.Archive.Sections, err = collect(ctx, s.q.ListArchiveSections, archiveSectionToDomain); err != nil {
		return out, err
	}
	if out.Archive.Records, err = collect(ctx, s.q.ListArchiveRecords, archiveRecordToDomain); err != nil {
		return out, err
	}
	if out.Frequencies, err = collect(ctx, s.q.ListFrequencies, frequencyToDomain); err != nil {
		return out, err
	}

	// Derive each technology's usedInMissions index from the mission list,
	// preserving mission order (matches how the frontend computes it today).
	used := make(map[string][]string)
	for _, m := range out.Missions {
		for _, techID := range m.Technologies {
			used[techID] = append(used[techID], m.ID)
		}
	}
	for i := range out.Technologies {
		out.Technologies[i].UsedInMissions = used[out.Technologies[i].ID]
	}

	copyRows, err := s.q.ListSiteCopy(ctx)
	if err != nil {
		return out, err
	}
	out.SiteCopy = make(map[string]string, len(copyRows))
	for _, c := range copyRows {
		out.SiteCopy[c.Key] = c.Value
	}

	return out, nil
}

// collect lists rows via list and maps each through toDom into a domain slice.
func collect[Row any, Dom any](
	ctx context.Context,
	list func(context.Context) ([]Row, error),
	toDom func(Row) (Dom, error),
) ([]Dom, error) {
	rows, err := list(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Dom, 0, len(rows))
	for _, row := range rows {
		d, err := toDom(row)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}
