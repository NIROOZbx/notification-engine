-- name: CreateEnvironment :exec

INSERT INTO environments (workspace_id, name)
VALUES ($1, $2);

-- name: GetEnvironmentsByWorkspace :many
SELECT * FROM environments
WHERE workspace_id = $1;

-- name: GetEnvironmentByID :one
SELECT * FROM environments
WHERE id = $1
LIMIT 1;

-- name: DeleteEnvironment :exec
DELETE FROM environments
WHERE id = $1;