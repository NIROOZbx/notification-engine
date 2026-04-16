-- name: CreateLayout :one
INSERT INTO layouts (
    workspace_id,
    name,
    html,
    is_default
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetLayoutByID :one
SELECT * FROM layouts
WHERE id = $1 AND workspace_id = $2
LIMIT 1;

-- name: GetDefaultLayout :one
SELECT * FROM layouts
WHERE workspace_id = $1 AND is_default = true
LIMIT 1;

-- name: ListLayouts :many
SELECT * FROM layouts
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: UpdateLayout :one
UPDATE layouts
SET
    name = $3,
    html = $4,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteLayout :exec
DELETE FROM layouts
WHERE id = $1 AND workspace_id = $2;

-- name: SetLayoutDefault :exec
UPDATE layouts
SET is_default = (id = $1)
WHERE workspace_id = $2;

