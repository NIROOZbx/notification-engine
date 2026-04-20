-- name: CreateChannelConfig :one
INSERT INTO channel_configs (
        workspace_id,
        channel,
        provider,
        display_name,
        credentials,
        is_active,
        is_default
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: GetChannelConfigByIDAndWorkspace :one
SELECT *
FROM channel_configs
WHERE id = $1
    and workspace_id = $2;
-- name: GetDefaultChannelConfig :one
SELECT *
FROM channel_configs
WHERE workspace_id = $1
    AND channel = $2
    AND is_default = true
    AND deleted_at IS NULL
LIMIT 1;
-- name: ListChannelConfigs :many
SELECT *
FROM channel_configs
WHERE workspace_id = $1
    AND deleted_at IS NULL
ORDER BY channel,
    created_at DESC;
-- name: UpdateChannelConfig :one
UPDATE channel_configs
SET display_name = COALESCE(sqlc.narg('display_name'), display_name),
    credentials = COALESCE(sqlc.narg('credentials'), credentials),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = $1
    AND workspace_id = $2
    AND deleted_at IS NULL
RETURNING *;
-- name: SetChannelConfigDefault :exec
UPDATE channel_configs
SET is_default = true,
    updated_at = NOW()
WHERE id = $1
    AND workspace_id = $2
    AND deleted_at IS NULL;

-- name: UnsetChannelConfigDefault :exec
UPDATE channel_configs
SET is_default = false,
    updated_at = NOW()
WHERE workspace_id = $1
    AND channel = $2
    AND deleted_at IS NULL;

    
-- name: GetChannelConfigByID :one
SELECT *
FROM channel_configs
WHERE id = $1
    AND workspace_id = $2
    AND deleted_at IS NULL;
-- name: CountActiveProvidersForChannel :one
SELECT COUNT(*)
FROM channel_configs
WHERE workspace_id = $1
    AND channel = $2
    AND deleted_at IS NULL;
-- name: RemoveProviderOverride :exec
UPDATE template_channels tc
SET channel_config_id = NULL
from templates t
WHERE tc.template_id = t.id
    and tc.channel_config_id = $1
    and t.workspace_id = $2;
-- name: SoftDeleteChannelConfig :exec
UPDATE channel_configs
SET deleted_at = NOW(),
    is_active = false,
    is_default = false
WHERE id = $1
    AND workspace_id = $2;
-- name: GetProviderForWorker :one
SELECT *
FROM channel_configs
WHERE id = $1
    AND workspace_id = $2;