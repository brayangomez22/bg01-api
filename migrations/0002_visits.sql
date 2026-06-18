-- +goose Up
-- +goose StatementBegin

-- ─────────────────────────────────────────────────────────────────────────
-- VISITS — privacy-preserving traffic counter for the control center.
--
-- One row per (station-local day, anonymous visitor). The visitor_hash is a
-- daily-rotating SHA-256 of the client IP + User-Agent salted with the day and
-- the server secret (see internal/server/visits.go): irreversible, holds no PII,
-- and changes every day. INSERT OR IGNORE on the primary key collapses repeat
-- visits so each person counts once per day. "Total" is the sum of daily
-- uniques (COUNT(*)), the standard privacy-analytics visit metric.
-- ─────────────────────────────────────────────────────────────────────────
CREATE TABLE visits (
    day          TEXT NOT NULL,                              -- 'YYYY-MM-DD' (America/Bogota)
    visitor_hash TEXT NOT NULL,
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (day, visitor_hash)
) WITHOUT ROWID;

-- Range/aggregate reads are always keyed by day.
CREATE INDEX idx_visits_day ON visits (day);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS visits;
-- +goose StatementEnd
