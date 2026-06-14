// Package server wires HTTP routing and middleware for the BG-01 API.
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/brayangomez22/bg01-api/internal/config"
	"github.com/brayangomez22/bg01-api/internal/domain"
	"github.com/brayangomez22/bg01-api/internal/store"
)

// server holds shared dependencies for HTTP handlers.
type server struct {
	cfg    config.Config
	logger *slog.Logger
	q      *store.Queries
	secret []byte
}

// New builds the configured *http.Server. Routes are registered on the stdlib
// ServeMux (Go 1.22+ method-aware routing); a third-party router can be
// introduced later without touching the entrypoint.
func New(cfg config.Config, logger *slog.Logger, q *store.Queries) *http.Server {
	s := &server{
		cfg:    cfg,
		logger: logger,
		q:      q,
		secret: []byte(cfg.SessionSecret),
	}

	mux := http.NewServeMux()
	s.routes(mux)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           cors(cfg, logging(logger, mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// routes mounts every handler. Public reads (health, export) plus the admin
// auth endpoints and the protected CRUD surface.
func (s *server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.handleHealth())
	mux.HandleFunc("GET /export", s.handleExport())

	mux.HandleFunc("POST /admin/login", s.handleLogin())
	mux.HandleFunc("POST /admin/logout", s.handleLogout())
	mux.HandleFunc("GET /admin/session", s.requireAuth(s.handleSession()))

	mux.HandleFunc("GET /admin/pilot", s.requireAuth(s.handlePilotGet()))
	mux.HandleFunc("PUT /admin/pilot", s.requireAuth(s.handlePilotPut()))

	registerResource(mux, s, "technologies", resource[store.Technology, domain.Technology, store.UpsertTechnologyParams]{
		list: s.q.ListTechnologies, toDom: technologyToDomain, fromDom: technologyFromDomain, upsert: s.q.UpsertTechnology, del: s.q.DeleteTechnology,
	})
	registerResource(mux, s, "missions", resource[store.Mission, domain.Mission, store.UpsertMissionParams]{
		list: s.q.ListMissions, toDom: missionToDomain, fromDom: missionFromDomain, upsert: s.q.UpsertMission, del: s.q.DeleteMission,
	})
	registerResource(mux, s, "experiences", resource[store.Experience, domain.Experience, store.UpsertExperienceParams]{
		list: s.q.ListExperiences, toDom: experienceToDomain, fromDom: experienceFromDomain, upsert: s.q.UpsertExperience, del: s.q.DeleteExperience,
	})
	registerResource(mux, s, "training", resource[store.TrainingSim, domain.TrainingSim, store.UpsertTrainingSimParams]{
		list: s.q.ListTrainingSims, toDom: trainingToDomain, fromDom: trainingFromDomain, upsert: s.q.UpsertTrainingSim, del: s.q.DeleteTrainingSim,
	})
	registerResource(mux, s, "archive/sections", resource[store.ArchiveSection, domain.ArchiveSection, store.UpsertArchiveSectionParams]{
		list: s.q.ListArchiveSections, toDom: archiveSectionToDomain, fromDom: archiveSectionFromDomain, upsert: s.q.UpsertArchiveSection, del: s.q.DeleteArchiveSection,
	})
	registerResource(mux, s, "archive/records", resource[store.ArchiveRecord, domain.ArchiveRecord, store.UpsertArchiveRecordParams]{
		list: s.q.ListArchiveRecords, toDom: archiveRecordToDomain, fromDom: archiveRecordFromDomain, upsert: s.q.UpsertArchiveRecord, del: s.q.DeleteArchiveRecord,
	})
	registerResource(mux, s, "frequencies", resource[store.Frequency, domain.Frequency, store.UpsertFrequencyParams]{
		list: s.q.ListFrequencies, toDom: frequencyToDomain, fromDom: frequencyFromDomain, upsert: s.q.UpsertFrequency, del: s.q.DeleteFrequency,
	})
	registerResource(mux, s, "site-copy", resource[store.SiteCopy, domain.SiteCopyEntry, store.UpsertSiteCopyParams]{
		list: s.q.ListSiteCopy, toDom: siteCopyToDomain, fromDom: siteCopyFromDomain, upsert: s.q.UpsertSiteCopy, del: s.q.DeleteSiteCopy,
	})
}
