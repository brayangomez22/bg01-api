-- name: ListFrequencies :many
SELECT * FROM frequencies ORDER BY sort_order, id;

-- name: GetFrequency :one
SELECT * FROM frequencies WHERE id = @id;

-- name: UpsertFrequency :one
INSERT INTO frequencies (id, label, handle, url, icon, is_primary, sort_order)
VALUES (@id, @label, @handle, @url, @icon, @is_primary, @sort_order)
ON CONFLICT(id) DO UPDATE SET
    label      = excluded.label,
    handle     = excluded.handle,
    url        = excluded.url,
    icon       = excluded.icon,
    is_primary = excluded.is_primary,
    sort_order = excluded.sort_order
RETURNING *;

-- name: DeleteFrequency :exec
DELETE FROM frequencies WHERE id = @id;
