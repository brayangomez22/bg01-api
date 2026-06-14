-- name: GetPilot :one
SELECT * FROM pilot WHERE id = 1;

-- name: UpdatePilot :one
UPDATE pilot SET
    name       = @name,
    callsign   = @callsign,
    role       = @role,
    available  = @available,
    location   = @location,
    bio        = @bio,
    manifesto  = @manifesto,
    stats      = @stats,
    avatar     = @avatar,
    resume_url = @resume_url,
    updated_at = datetime('now')
WHERE id = 1
RETURNING *;
