-- name: GetTemplateByEventType :one
SELECT * FROM templates 
WHERE event_type = $1 
  AND workspace_id = $2 
  AND environment_id = $3
LIMIT 1;

-- name: GetTemplateByEventAndChannel :one
SELECT * FROM templates 
WHERE event_type = $1 
  AND workspace_id = $2
  AND environment_id = $3
  AND status = 'live'
LIMIT 1;

-- name: GetTemplateByID :one
SELECT * FROM templates
WHERE id = $1 AND workspace_id = $2
LIMIT 1;

-- name: ListTemplates :many
SELECT * FROM templates
WHERE workspace_id = $1 AND environment_id = $2
ORDER BY created_at DESC;

-- name: CreateTemplate :one
INSERT INTO templates (
    workspace_id,
    environment_id,
    layout_id,
    created_by,
    name,
    description,
    event_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateTemplate :one
UPDATE templates
SET
    name = $3,
    description = $4,
    status = $5,
    layout_id = $6,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteTemplate :exec
DELETE FROM templates
WHERE id = $1 AND workspace_id = $2;

