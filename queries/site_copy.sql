-- name: ListSiteCopy :many
SELECT * FROM site_copy ORDER BY key;

-- name: GetSiteCopy :one
SELECT * FROM site_copy WHERE key = @key;

-- name: UpsertSiteCopy :one
INSERT INTO site_copy (key, value)
VALUES (@key, @value)
ON CONFLICT(key) DO UPDATE SET
    value      = excluded.value,
    updated_at = datetime('now')
RETURNING *;

-- name: DeleteSiteCopy :exec
DELETE FROM site_copy WHERE key = @key;
