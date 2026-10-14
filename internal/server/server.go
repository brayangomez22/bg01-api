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
	// loginLimiter throttles POST /admin/login per client IP.
	loginLimiter *rateLimiter
	// pulseLimiter throttles the public POST /pulse beacon per client IP.
	pulseLimiter *rateLimiter
	// githubAPI and httpClient back POST /admin/publish; tests point them at
	// a fake GitHub.
	githubAPI  string
	httpClient *http.Client
	// now is the clock for visit days; tests freeze it.
	now func() time.Time
}

// New builds the configured *http.Server. Routes are registered on the stdlib
// ServeMux (Go 1.22+ method-aware routing); a third-party router can be
// introduced later without touching the entrypoint.
func New(cfg config.Config, logger *slog.Logger, q *store.Queries) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           newServer(cfg, logger, q).handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		// Must exceed the 10s GitHub call in POST /admin/publish.
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func newServer(cfg config.Config, logger *slog.Logger, q *store.Queries) *server {
	return &server{
		cfg:    cfg,
		logger: logger,
		q:      q,
		secret: []byte(cfg.SessionSecret),

		loginLimiter: newRateLimiter(loginAttempts, loginWindow),
		pulseLimiter: newRateLimiter(pulseHits, pulseWindow),

		githubAPI:  "https://api.github.com",
		httpClient: http.DefaultClient,
		now:        time.Now,
	}
}

// handler returns the routed mux wrapped in the shared middleware.
func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	s.routes(mux)
	return cors(s.cfg, logging(s.logger, mux))
}

// routes mounts every handler. Public reads (health, export) plus the admin
// auth endpoints and the protected CRUD surface.
func (s *server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.handleHealth())
	mux.HandleFunc("GET /export", s.handleExport())

	// Public, anonymous visit beacon fired by the static portfolio on load.
	mux.HandleFunc("POST /pulse", s.pulseLimiter.wrap(s, s.handlePulse()))

	mux.HandleFunc("POST /admin/login", s.loginLimiter.wrap(s, s.handleLogin()))
	mux.HandleFunc("POST /admin/logout", s.handleLogout())
	mux.HandleFunc("GET /admin/session", s.requireAuth(s.handleSession()))
	mux.HandleFunc("GET /admin/stats", s.requireAuth(s.handleStats()))
	mux.HandleFunc("POST /admin/publish", s.requireAuth(s.handlePublish()))

	mux.HandleFunc("GET /admin/pilot", s.requireAuth(s.handlePilotGet()))
	mux.HandleFunc("PUT /admin/pilot", s.requireAuth(s.handlePilotPut()))

	registerResource(mux, s, "technologies", resource[store.Technology, domain.Technology, store.UpsertTechnologyParams]{
		key:  func(d domain.Technology) string { return d.ID },
		list: s.q.ListTechnologies, toDom: technologyToDomain, fromDom: technologyFromDomain, upsert: s.q.UpsertTechnology, del: s.q.DeleteTechnology,
	})
	registerResource(mux, s, "missions", resource[store.Mission, domain.Mission, store.UpsertMissionParams]{
		key:  func(d domain.Mission) string { return d.ID },
		list: s.q.ListMissions, toDom: missionToDomain, fromDom: missionFromDomain, upsert: s.q.UpsertMission, del: s.q.DeleteMission,
	})
	registerResource(mux, s, "experiences", resource[store.Experience, domain.Experience, store.UpsertExperienceParams]{
		key:  func(d domain.Experience) string { return d.ID },
		list: s.q.ListExperiences, toDom: experienceToDomain, fromDom: experienceFromDomain, upsert: s.q.UpsertExperience, del: s.q.DeleteExperience,
	})
	registerResource(mux, s, "training", resource[store.TrainingSim, domain.TrainingSim, store.UpsertTrainingSimParams]{
		key:  func(d domain.TrainingSim) string { return d.Repo },
		list: s.q.ListTrainingSims, toDom: trainingToDomain, fromDom: trainingFromDomain, upsert: s.q.UpsertTrainingSim, del: s.q.DeleteTrainingSim,
	})
	registerResource(mux, s, "archive/sections", resource[store.ArchiveSection, domain.ArchiveSection, store.UpsertArchiveSectionParams]{
		key:  func(d domain.ArchiveSection) string { return d.ID },
		list: s.q.ListArchiveSections, toDom: archiveSectionToDomain, fromDom: archiveSectionFromDomain, upsert: s.q.UpsertArchiveSection, del: s.q.DeleteArchiveSection,
	})
	registerResource(mux, s, "archive/records", resource[store.ArchiveRecord, domain.ArchiveRecord, store.UpsertArchiveRecordParams]{
		key:  func(d domain.ArchiveRecord) string { return d.ID },
		list: s.q.ListArchiveRecords, toDom: archiveRecordToDomain, fromDom: archiveRecordFromDomain, upsert: s.q.UpsertArchiveRecord, del: s.q.DeleteArchiveRecord,
	})
	registerResource(mux, s, "frequencies", resource[store.Frequency, domain.Frequency, store.UpsertFrequencyParams]{
		key:  func(d domain.Frequency) string { return d.ID },
		list: s.q.ListFrequencies, toDom: frequencyToDomain, fromDom: frequencyFromDomain, upsert: s.q.UpsertFrequency, del: s.q.DeleteFrequency,
	})
	registerResource(mux, s, "site-copy", resource[store.SiteCopy, domain.SiteCopyEntry, store.UpsertSiteCopyParams]{
		key:  func(d domain.SiteCopyEntry) string { return d.Key },
		list: s.q.ListSiteCopy, toDom: siteCopyToDomain, fromDom: siteCopyFromDomain, upsert: s.q.UpsertSiteCopy, del: s.q.DeleteSiteCopy,
	})
}
