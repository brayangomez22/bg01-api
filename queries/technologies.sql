-- name: ListTechnologies :many
SELECT * FROM technologies ORDER BY sort_order, name;

-- name: GetTechnology :one
SELECT * FROM technologies WHERE id = @id;

-- name: UpsertTechnology :one
INSERT INTO technologies (id, name, category, proficiency, since, description, planet, featured, sort_order)
VALUES (@id, @name, @category, @proficiency, @since, @description, @planet, @featured, @sort_order)
ON CONFLICT(id) DO UPDATE SET
    name        = excluded.name,
    category    = excluded.category,
    proficiency = excluded.proficiency,
    since       = excluded.since,
    description = excluded.description,
    planet      = excluded.planet,
    featured    = excluded.featured,
    sort_order  = excluded.sort_order,
    updated_at  = datetime('now')
RETURNING *;

-- name: DeleteTechnology :exec
DELETE FROM technologies WHERE id = @id;
