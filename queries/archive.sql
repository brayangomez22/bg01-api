-- name: ListArchiveSections :many
SELECT * FROM archive_sections ORDER BY sort_order, id;

-- name: UpsertArchiveSection :one
INSERT INTO archive_sections (id, code, label, sort_order)
VALUES (@id, @code, @label, @sort_order)
ON CONFLICT(id) DO UPDATE SET
    code       = excluded.code,
    label      = excluded.label,
    sort_order = excluded.sort_order
RETURNING *;

-- name: DeleteArchiveSection :exec
DELETE FROM archive_sections WHERE id = @id;

-- name: ListArchiveRecords :many
SELECT * FROM archive_records ORDER BY sort_order, archived_at DESC;

-- name: GetArchiveRecord :one
SELECT * FROM archive_records WHERE id = @id;

-- name: UpsertArchiveRecord :one
INSERT INTO archive_records (
    id, code, title, abstract, section, tags, archived_at, reading_minutes,
    body, refs, sort_order
) VALUES (
    @id, @code, @title, @abstract, @section, @tags, @archived_at, @reading_minutes,
    @body, @refs, @sort_order
)
ON CONFLICT(id) DO UPDATE SET
    code            = excluded.code,
    title           = excluded.title,
    abstract        = excluded.abstract,
    section         = excluded.section,
    tags            = excluded.tags,
    archived_at     = excluded.archived_at,
    reading_minutes = excluded.reading_minutes,
    body            = excluded.body,
    refs            = excluded.refs,
    sort_order      = excluded.sort_order,
    updated_at      = datetime('now')
RETURNING *;

-- name: DeleteArchiveRecord :exec
DELETE FROM archive_records WHERE id = @id;
