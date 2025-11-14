-- name: AssignReviewerToPullRequest :batchone
INSERT INTO pull_requests_reviewers (pull_request_id, reviewer_id)
VALUES ($1, $2)
ON CONFLICT (pull_request_id, reviewer_id) DO NOTHING
RETURNING *;

-- name: GetReviewersByPullRequestID :many
SELECT u.*
FROM users u
JOIN pull_requests_reviewers prr ON u.id = prr.reviewer_id
WHERE prr.pull_request_id = $1;

-- name: IsUserReviewerForPullRequest :one
SELECT EXISTS (
    SELECT 1
    FROM pull_requests_reviewers
    WHERE pull_request_id = $1 AND reviewer_id = $2
) AS "exists";

-- name: ReassignReviewerForPullRequest :exec
UPDATE pull_requests_reviewers
SET reviewer_id = $2
WHERE pull_request_id = $1 AND reviewer_id = $3;

-- name: GetUsersReviewingPullRequest :many
SELECT pr.* FROM pull_requests_reviewers prr
JOIN pull_requests pr ON pr.id = prr.pull_request_id AND pr.merged_at IS NULL
WHERE prr.reviewer_id = $1;