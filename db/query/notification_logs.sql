-- name: InsertNotificationLog :one
INSERT INTO notification_logs (
        workspace_id,
        environment_id,
        template_id,
        external_user_id,
        event_type,
        channel,
        recipient,
        idempotency_key,
        is_test,
        scheduled_at,
        trigger_data,
        status
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    )
RETURNING *;
-- name: GetNotificationLogByID :one
SELECT *
FROM notification_logs
WHERE id = $1;
-- name: GetNotificationLogByIdempotencyKey :one
SELECT *
FROM notification_logs
WHERE idempotency_key = $1;
-- name: UpdateNotificationLog :one
UPDATE notification_logs
SET status = $2,
    rendered_content = $3,
    attempt_count = $4,
    sent_at = $5,
    next_retry_at = $6,
    error_message = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
-- name: UpdateNotificationStatus :one
UPDATE notification_logs
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
-- name: GetDueScheduledNotifications :many
SELECT *
FROM notification_logs
WHERE status = 'scheduled'
    AND scheduled_at IS NOT NULL
    AND scheduled_at <= NOW() FOR
UPDATE SKIP LOCKED
LIMIT $1;
-- name: GetDueRetryNotifications :many
SELECT *
FROM notification_logs
WHERE status = 'retrying'
    AND next_retry_at IS NOT NULL
    AND next_retry_at <= NOW() FOR
UPDATE SKIP LOCKED
LIMIT $1;

-- name: UpdateDeliveryStatusByProviderID :exec
UPDATE notification_logs
SET delivery_status = $2,
    delivered_at = CASE
        WHEN $2 = 'delivered' THEN $3
        ELSE NULL
    END,
    failed_at = CASE
        WHEN $2 IN ('failed', 'bounced') THEN $3
        ELSE NULL
    END,
    provider_response = $4,
    updated_at = NOW()
WHERE provider_message_id = $1;
-- name: GetNotificationLogByProviderMessageID :one
SELECT *
FROM notification_logs
WHERE provider_message_id = $1;

-- name: UpdateProviderMessageID :exec
UPDATE notification_logs
SET provider_message_id = $2,
    status = 'sent',
    delivery_status = 'sent',
    updated_at = NOW()
WHERE id = $1;