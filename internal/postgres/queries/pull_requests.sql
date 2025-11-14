-- name: CreatePullRequest :one
INSERT INTO pull_requests (id, name, author_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ExistsPullRequestByID :one
SELECT EXISTS (
    SELECT 1
    FROM pull_requests
    WHERE id = $1
) AS "exists";

-- name: GetPullRequestByID :one
SELECT *
FROM pull_requests
WHERE id = $1;

-- name: MergePullRequest :one
UPDATE pull_requests
SET merged_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;