-- name: AddUsers :batchone
INSERT INTO users (id, username, team_id) VALUES
  ($1, $2, $3)
RETURNING *;