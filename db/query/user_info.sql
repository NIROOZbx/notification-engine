-- name: UpsertUserContactInfo :one
INSERT INTO user_info (
    workspace_id,
    environment_id,
    external_user_id,
    channel,
    contact_value,
    verified,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (workspace_id, environment_id, external_user_id, channel)
DO UPDATE SET
    contact_value = EXCLUDED.contact_value,
    metadata      = EXCLUDED.metadata
RETURNING *;

-- name: GetContactByExternalUserAndChannel :one
SELECT * FROM user_info
WHERE workspace_id    = $1
  AND environment_id  = $2
  AND external_user_id = $3
  AND channel         = $4;