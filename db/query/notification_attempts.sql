-- name: InsertNotificationAttempt :one
INSERT INTO notification_attempts (
    notification_log_id,
    attempt_count,
    status,
    error_message,
    provider,
    duration_ms,
    attempted_at
) VALUES (
    $1, $2, $3, $4, $5,$6, NOW()
)
RETURNING *;

-- name: GetAttemptsByNotificationLogID :many
SELECT * FROM notification_attempts
WHERE notification_log_id = $1
ORDER BY attempt_count ASC;

-- name: GetAttemptCountByLogID :one

SELECT attempt_count from notification_attempts where notification_log_id =$1;