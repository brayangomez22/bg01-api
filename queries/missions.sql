-- name: ListMissions :many
SELECT * FROM missions ORDER BY sort_order, code;

-- name: GetMission :one
SELECT * FROM missions WHERE id = @id;

-- name: UpsertMission :one
INSERT INTO missions (
    id, code, title, summary, description, status, role, duration_label,
    period_start, period_end, technologies, highlights, challenges, metrics,
    links, cover, gallery, featured, sort_order
) VALUES (
    @id, @code, @title, @summary, @description, @status, @role, @duration_label,
    @period_start, @period_end, @technologies, @highlights, @challenges, @metrics,
    @links, @cover, @gallery, @featured, @sort_order
)
ON CONFLICT(id) DO UPDATE SET
    code           = excluded.code,
    title          = excluded.title,
    summary        = excluded.summary,
    description    = excluded.description,
    status         = excluded.status,
    role           = excluded.role,
    duration_label = excluded.duration_label,
    period_start   = excluded.period_start,
    period_end     = excluded.period_end,
    technologies   = excluded.technologies,
    highlights     = excluded.highlights,
    challenges     = excluded.challenges,
    metrics        = excluded.metrics,
    links          = excluded.links,
    cover          = excluded.cover,
    gallery        = excluded.gallery,
    featured       = excluded.featured,
    sort_order     = excluded.sort_order,
    updated_at     = datetime('now')
RETURNING *;

-- name: DeleteMission :exec
DELETE FROM missions WHERE id = @id;
