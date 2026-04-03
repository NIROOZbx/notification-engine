-- name: GetChannelConfigByID :one
SELECT * FROM channel_configs
WHERE id = $1
AND workspace_id = $2
AND is_active = true;