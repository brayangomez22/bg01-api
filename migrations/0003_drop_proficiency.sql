-- +goose Up
-- The 0-100 mastery score was never shown on the public site; "since" carries
-- the experience instead.
ALTER TABLE technologies DROP COLUMN proficiency;

-- +goose Down
ALTER TABLE technologies ADD COLUMN proficiency INTEGER NOT NULL DEFAULT 0;
