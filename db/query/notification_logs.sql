-- name: InsertNotificationLog :one
INSERT INTO notification_logs (
    workspace_id,
    environment_id,
    template_id,
    external_user_id,
    event_type,
    channel,
    status,
    recipient,
    idempotency_key,
    is_test
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,$10
)
RETURNING *;

-- name: GetNotificationLogByID :one
SELECT * FROM notification_logs
WHERE id = $1;

-- name: GetNotificationLogByIdempotencyKey :one
SELECT * FROM notification_logs
WHERE idempotency_key = $1;

-- name: UpdateNotificationLogStatus :one
UPDATE notification_logs
SET
    status           = $2,
    rendered_content = $3,
    attempt_count    = $4,
    sent_at          = $5
WHERE id = $1
RETURNING *;