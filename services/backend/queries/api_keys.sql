-- name: CreateApiKey :one
INSERT INTO api_keys (
        workspace_id,
        environment_id,
        label,
        key_hash,
        key_hint,
        created_by,
        expires_at
    )
VALUES ($1, $2, $3, $4, $5, $6,$7)
RETURNING *;

-- name: ValidateAndTouchAPIKey :one 
UPDATE api_keys
SET last_used_at = NOW()
WHERE key_hash = $1
    AND is_revoked = FALSE
    AND (expires_at IS NULL OR expires_at > NOW())
RETURNING id, workspace_id, environment_id, label;

-- name: GetAPIKeyByID :one
SELECT *
FROM api_keys
WHERE id = $1
    AND workspace_id = $2;

-- name: ListAPIKeysByWorkspaceAndEnv :many
SELECT
    id,
    workspace_id,
    environment_id,
    label,
    key_hint,
    is_revoked,
    revoked_at,
    expires_at,
    created_by,
    last_used_at,
    created_at,
    updated_at
FROM api_keys
WHERE workspace_id = $1
AND environment_id = $2
ORDER BY created_at DESC;

-- name: RevokeAPIKey :one
UPDATE api_keys
SET is_revoked = TRUE,
    revoked_at = NOW()
WHERE id = $1
    AND workspace_id = $2
    AND is_revoked=false
RETURNING id,is_revoked, revoked_at;

-- name: DeleteAPIKey :execrows
DELETE FROM api_keys
WHERE id = $1
    AND workspace_id = $2;

-- name: CountActiveAPIKeys :one
SELECT COUNT(*) FROM api_keys 
WHERE workspace_id = $1 
AND is_revoked = FALSE
AND (expires_at IS NULL OR expires_at > NOW());