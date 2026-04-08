-- name: CreateChannelConfig :one
INSERT INTO channel_configs (
    workspace_id,
    channel,
    provider,
    display_name,
    credentials,
    is_active,
    is_default
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetChannelConfigByID :one
SELECT * FROM channel_configs
WHERE id = $1 AND workspace_id = $2
LIMIT 1;

-- name: GetDefaultChannelConfig :one
SELECT * FROM channel_configs
WHERE workspace_id = $1 AND channel = $2 AND is_default = true
LIMIT 1;

-- name: ListChannelConfigs :many
SELECT * FROM channel_configs
WHERE workspace_id = $1
ORDER BY channel, created_at DESC;

-- name: UpdateChannelConfig :one
UPDATE channel_configs
SET
    display_name = $3,
    credentials = $4,
    is_active = $5,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteChannelConfig :exec
DELETE FROM channel_configs
WHERE id = $1 AND workspace_id = $2;

-- name: SetChannelConfigDefault :exec
UPDATE channel_configs
SET is_default = (id = $1)
WHERE workspace_id = $2 AND channel = $3;