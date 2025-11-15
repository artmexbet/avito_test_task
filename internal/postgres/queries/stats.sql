-- name: GetAssignmentStats :many
SELECT prr.reviewer_id,
       u.is_active,
       COUNT(prr.pull_request_id) AS assigned_pull_requests
FROM pull_requests_reviewers prr
         JOIN users u ON prr.reviewer_id = u.id
GROUP BY prr.reviewer_id, u.id
ORDER BY assigned_pull_requests;

-- name: GetTeamsCount :many
SELECT t.name, COUNT(*) AS pr_count
FROM pull_requests_reviewers prr
         JOIN users u ON u.id = prr.reviewer_id
         JOIN teams t ON u.team_name = t.name
GROUP BY t.name;

-- name: GetUsersCount :many
SELECT u.is_active, COUNT(*) AS user_count
FROM users u
GROUP BY u.is_active;