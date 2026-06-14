-- name: ListExperiences :many
SELECT * FROM experiences ORDER BY sort_order, period_start DESC;

-- name: GetExperience :one
SELECT * FROM experiences WHERE id = @id;

-- name: UpsertExperience :one
INSERT INTO experiences (
    id, period_start, period_end, company, role, location, summary,
    responsibilities, achievements, technologies, type, sort_order
) VALUES (
    @id, @period_start, @period_end, @company, @role, @location, @summary,
    @responsibilities, @achievements, @technologies, @type, @sort_order
)
ON CONFLICT(id) DO UPDATE SET
    period_start     = excluded.period_start,
    period_end       = excluded.period_end,
    company          = excluded.company,
    role             = excluded.role,
    location         = excluded.location,
    summary          = excluded.summary,
    responsibilities = excluded.responsibilities,
    achievements     = excluded.achievements,
    technologies     = excluded.technologies,
    type             = excluded.type,
    sort_order       = excluded.sort_order,
    updated_at       = datetime('now')
RETURNING *;

-- name: DeleteExperience :exec
DELETE FROM experiences WHERE id = @id;
