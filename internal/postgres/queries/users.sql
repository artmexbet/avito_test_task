-- name: AddUsers :batchone
INSERT INTO users (id, username, team_name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET username   = EXCLUDED.username,
                               team_name    = EXCLUDED.team_name,
                               updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: BatchExistsUserByID :batchone
SELECT EXISTS (SELECT 1
               FROM users
               WHERE id = $1) AS "exists";

-- name: ExistsUserByID :one
SELECT EXISTS (SELECT 1
               FROM users
               WHERE id = $1) AS "exists";

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUsersByTeamName :many
SELECT *
FROM users
WHERE team_name = $1;

-- name: SetUserIsActiveByID :one
UPDATE users
SET is_active  = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetActiveUsersByTeamName :many
SELECT *
FROM users
WHERE team_name = $1 AND is_active = TRUE;
