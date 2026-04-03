-- name: InsertNotificationAttempt :one
INSERT INTO notification_attempts (
    notification_log_id,
    attempt_count,
    status,
    error_message,
    provider,
    attempted_at
) VALUES (
    $1, $2, $3, $4, $5, NOW()
)
RETURNING *;

-- name: GetAttemptsByNotificationLogID :many
SELECT * FROM notification_attempts
WHERE notification_log_id = $1
ORDER BY attempt_count ASC;

--name: GetAttemptCountByLogID :one

SELECT attempt_count from notification_attempts where notif_log_id =$1;