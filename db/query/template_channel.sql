-- name: GetTemplateChannelByTemplateAndChannel :one
SELECT *
FROM template_channels
WHERE template_id = $1
    AND channel = $2
    AND is_active = true;
-- name: GetActiveChannelsByTemplateID :many
SELECT *
FROM template_channels
WHERE template_id = $1
    AND is_active = true;
-- name: CreateTemplateChannel :one
INSERT INTO template_channels (
        template_id,
        channel_config_id,
        channel,
        content,
        is_active
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: GetTemplateChannelByID :one
SELECT *
FROM template_channels
WHERE id = $1
LIMIT 1;
-- name: ListTemplateChannels :many
SELECT *
FROM template_channels
WHERE template_id = $1
ORDER BY created_at DESC;
-- name: UpdateTemplateChannel :one
UPDATE template_channels
SET channel_config_id = $2,
    content = $3,
    is_active = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
-- name: DeleteTemplateChannel :execresult
DELETE FROM template_channels
WHERE id = $1
    AND template_id = $2;
-- name: HasActiveChannels :one
SELECT EXISTS (
        SELECT 1
        FROM template_channels
        WHERE template_id = $1
            AND is_active = true
    );