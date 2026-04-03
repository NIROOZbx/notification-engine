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

