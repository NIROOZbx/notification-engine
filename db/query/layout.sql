-- name: GetLayoutByID :one
SELECT * FROM layouts
WHERE id = $1
AND workspace_id = $2
AND is_default=true;