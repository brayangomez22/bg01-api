-- name: ListTrainingSims :many
SELECT * FROM training_sims ORDER BY sort_order, year DESC;

-- name: GetTrainingSim :one
SELECT * FROM training_sims WHERE repo = @repo;

-- name: UpsertTrainingSim :one
INSERT INTO training_sims (repo, code, title, summary, stack, year, repo_url, sort_order)
VALUES (@repo, @code, @title, @summary, @stack, @year, @repo_url, @sort_order)
ON CONFLICT(repo) DO UPDATE SET
    code       = excluded.code,
    title      = excluded.title,
    summary    = excluded.summary,
    stack      = excluded.stack,
    year       = excluded.year,
    repo_url   = excluded.repo_url,
    sort_order = excluded.sort_order,
    updated_at = datetime('now')
RETURNING *;

-- name: DeleteTrainingSim :exec
DELETE FROM training_sims WHERE repo = @repo;
