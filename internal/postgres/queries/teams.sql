-- name: GetTeamByName :one
SELECT *
FROM teams
WHERE name = $1;

-- name: AddTeam :one
INSERT INTO teams (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name       = EXCLUDED.name,
                                 updated_at = CURRENT_TIMESTAMP
RETURNING *;

