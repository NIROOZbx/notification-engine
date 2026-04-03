-- name: GetTemplateChannelByTemplateAndChannel :one
SELECT * FROM template_channels
WHERE template_id = $1
AND channel = $2
AND is_active = true;

-- name: GetActiveChannelsByTemplateID :many
SELECT * FROM template_channels
WHERE template_id = $1
AND is_active = true;