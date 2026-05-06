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

-- name: ListUserPreferencesBySubscriber :many
SELECT up.* FROM user_preferences up
JOIN user_info ui ON up.subscriber_id = ui.id
WHERE ui.workspace_id = $1 
  AND ui.environment_id = $2 
  AND ui.external_user_id = $3;


-- name: GetContactWithPreference :one
SELECT sqlc.embed(u),sqlc.embed(p) from user_info as u LEFT JOIN
user_preferences as p on 
u.id=p.subscriber_id AND
p.event_type=$4 and p.channel=$5
WHERE u.external_user_id=$1
AND u.workspace_id=$2 
AND u.environment_id=$3
AND u.channel=$5;
