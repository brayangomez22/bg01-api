-- +goose Up
-- +goose StatementBegin

-- ─────────────────────────────────────────────────────────────────────────
-- BG-01 content schema. Mirrors src/types/domain.ts of the portfolio.
--
-- Conventions:
--   * List/object fields (highlights, metrics, refs, MediaRef, …) are stored as
--     JSON TEXT — single-user CMS, normalization would be busywork. They are
--     validated/serialized in Go and emitted verbatim by /export.
--   * Booleans are INTEGER 0/1.
--   * `sort_order` drives admin-controlled ordering (domain `order`).
--   * Dates are ISO strings (TEXT), matching ISODate in the frontend.
-- ─────────────────────────────────────────────────────────────────────────

-- PILOT — singleton row (id is pinned to 1).
CREATE TABLE pilot (
    id          INTEGER PRIMARY KEY CHECK (id = 1),
    name        TEXT    NOT NULL,
    callsign    TEXT    NOT NULL,
    role        TEXT    NOT NULL,
    available   INTEGER NOT NULL DEFAULT 1,
    location    TEXT    NOT NULL,
    bio         TEXT    NOT NULL,
    manifesto   TEXT    NOT NULL,
    stats       TEXT    NOT NULL DEFAULT '[]',  -- JSON [{label,value}]
    avatar      TEXT    NOT NULL,               -- JSON MediaRef
    resume_url  TEXT    NOT NULL,
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- TECHNOLOGY — stack "planets". usedInMissions is derived in /export, not stored.
CREATE TABLE technologies (
    id          TEXT    PRIMARY KEY,            -- slug
    name        TEXT    NOT NULL,
    category    TEXT    NOT NULL,               -- language|framework|database|cloud|tooling
    proficiency INTEGER NOT NULL,               -- 0-100
    since       TEXT    NOT NULL,               -- ISODate
    description TEXT    NOT NULL,
    planet      TEXT    NOT NULL,               -- JSON {color,size,orbit}
    featured    INTEGER NOT NULL DEFAULT 0,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- MISSION — projects.
CREATE TABLE missions (
    id             TEXT    PRIMARY KEY,         -- slug, route param
    code           TEXT    NOT NULL,            -- "M-001"
    title          TEXT    NOT NULL,
    summary        TEXT    NOT NULL,
    description    TEXT    NOT NULL,
    status         TEXT    NOT NULL,            -- completed|in-progress|classified
    role           TEXT    NOT NULL,
    duration_label TEXT    NOT NULL,
    period_start   TEXT    NOT NULL,            -- ISODate
    period_end     TEXT,                        -- ISODate, nullable (ongoing)
    technologies   TEXT    NOT NULL DEFAULT '[]', -- JSON [Technology.id]
    highlights     TEXT    NOT NULL DEFAULT '[]', -- JSON [string]
    challenges     TEXT    NOT NULL DEFAULT '[]', -- JSON [string]
    metrics        TEXT    NOT NULL DEFAULT '[]', -- JSON [{label,value}]
    links          TEXT    NOT NULL DEFAULT '{}', -- JSON {live,repo,caseStudy}
    cover          TEXT    NOT NULL,            -- JSON MediaRef
    gallery        TEXT    NOT NULL DEFAULT '[]', -- JSON [MediaRef]
    featured       INTEGER NOT NULL DEFAULT 0,
    sort_order     INTEGER NOT NULL DEFAULT 0,
    updated_at     TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- EXPERIENCE — logbook.
CREATE TABLE experiences (
    id               TEXT    PRIMARY KEY,
    period_start     TEXT    NOT NULL,          -- ISODate
    period_end       TEXT    NOT NULL,          -- ISODate or 'present'
    company          TEXT    NOT NULL,
    role             TEXT    NOT NULL,
    location         TEXT    NOT NULL,
    summary          TEXT    NOT NULL,
    responsibilities TEXT    NOT NULL DEFAULT '[]', -- JSON [string]
    achievements     TEXT    NOT NULL DEFAULT '[]', -- JSON [string]
    technologies     TEXT    NOT NULL DEFAULT '[]', -- JSON [Technology.id]
    type             TEXT    NOT NULL,          -- full-time|contract|freelance
    sort_order       INTEGER NOT NULL DEFAULT 0,
    updated_at       TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- TRAINING SIM — curated practice repos. `stack` holds plain labels (not refs).
CREATE TABLE training_sims (
    repo       TEXT    PRIMARY KEY,             -- GitHub repo name (slug)
    code       TEXT    NOT NULL,               -- "SIM-001"
    title      TEXT    NOT NULL,
    summary    TEXT    NOT NULL,
    stack      TEXT    NOT NULL DEFAULT '[]',  -- JSON [string]
    year       INTEGER NOT NULL,
    repo_url   TEXT    NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- ARCHIVE SECTION — shelf taxonomy (lng|ing|ops|bit).
CREATE TABLE archive_sections (
    id         TEXT    PRIMARY KEY,
    code       TEXT    NOT NULL,               -- "S-LNG"
    label      TEXT    NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- ARCHIVE RECORD — knowledge entries.
CREATE TABLE archive_records (
    id              TEXT    PRIMARY KEY,        -- slug, route param
    code            TEXT    NOT NULL,           -- "REG-001"
    title           TEXT    NOT NULL,
    abstract        TEXT    NOT NULL,
    section         TEXT    NOT NULL REFERENCES archive_sections(id),
    tags            TEXT    NOT NULL DEFAULT '[]', -- JSON [string]
    archived_at     TEXT    NOT NULL,           -- ISODate
    reading_minutes INTEGER NOT NULL,
    body            TEXT    NOT NULL DEFAULT '[]', -- JSON [ArchiveSegment]
    refs            TEXT    NOT NULL DEFAULT '[]', -- JSON [ArchiveMeta.id]
    sort_order      INTEGER NOT NULL DEFAULT 0,
    updated_at      TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- FREQUENCY — socials / comms channels.
CREATE TABLE frequencies (
    id         TEXT    PRIMARY KEY,             -- github|linkedin|email
    label      TEXT    NOT NULL,
    handle     TEXT    NOT NULL,
    url        TEXT    NOT NULL,
    icon       TEXT    NOT NULL,
    is_primary INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- SITE COPY — editable free-text strings (comms blurbs, etc.) as key/value.
CREATE TABLE site_copy (
    key        TEXT    PRIMARY KEY,
    value      TEXT    NOT NULL,
    updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- Seed the pilot singleton so the row always exists (edited, never created).
INSERT INTO pilot (id, name, callsign, role, location, bio, manifesto, avatar, resume_url)
VALUES (1, '', 'BG-01', '', '', '', '', '{}', '');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS site_copy;
DROP TABLE IF EXISTS frequencies;
DROP TABLE IF EXISTS archive_records;
DROP TABLE IF EXISTS archive_sections;
DROP TABLE IF EXISTS training_sims;
DROP TABLE IF EXISTS experiences;
DROP TABLE IF EXISTS missions;
DROP TABLE IF EXISTS technologies;
DROP TABLE IF EXISTS pilot;
-- +goose StatementEnd
