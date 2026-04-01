-- name: UpsertUserPreference :one
INSERT INTO user_preferences (
    workspace_id,
    environment_id,
    subscriber_id,
    channel,
    event_type,
    is_enabled
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (subscriber_id, channel, event_type)
DO UPDATE SET
    is_enabled = EXCLUDED.is_enabled
RETURNING *;

-- name: GetPreferencesBySubscriberAndChannel :many
SELECT * FROM user_preferences
WHERE subscriber_id = $1
  AND channel       = $2
  AND (event_type IS NULL OR event_type = $3);