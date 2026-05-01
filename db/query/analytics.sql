-- name: GetAggregateMetrics :many
SELECT nl.channel,
    na.provider,
    COUNT(*) AS total_count,
    COUNT(*) FILTER (
        WHERE nl.delivery_status = 'delivered'
    ) AS delivered_count,
    COUNT(*) FILTER (
        WHERE nl.delivery_status IN ('failed', 'bounced', 'complained')
            OR nl.delivery_status = 'failed'
    ) AS failed_count
FROM notification_logs nl
    LEFT JOIN notification_attempts na ON nl.id = na.notification_log_id
WHERE nl.workspace_id = $1
    AND nl.created_at >= $2
    AND nl.created_at <= $3
GROUP BY nl.channel,
    na.provider;
-- name: GetTimeSeriesData :many
SELECT CASE
        WHEN sqlc.arg('group_by')::text = 'hour' THEN date_trunc('hour', created_at)
        WHEN sqlc.arg('group_by')::text = 'week' THEN date_trunc('week', created_at)
        WHEN sqlc.arg('group_by')::text = 'month' THEN date_trunc('month', created_at)
        ELSE date_trunc('day', created_at)
    END::timestamptz AS time_bucket,
    count(*) FILTER (
        WHERE nl.delivery_status = 'sent'
    ) as total_sent,
    COUNT(*) FILTER (
        WHERE nl.delivery_status IN ('failed', 'bounced')
            OR nl.status = 'failed'
    ) as total_failed,
    count(*) FILTER (
        WHERE nl.delivery_status = 'delivered'
    ) as total_delivered
FROM notification_logs as nl
WHERE nl.workspace_id = $1
    AND nl.created_at >= $2
    AND nl.created_at <= $3
GROUP BY time_bucket
ORDER BY time_bucket;
-- name: GetProviderHealth :many
SELECT na.provider,
    AVG(na.duration_ms)::int AS avg_latency,
    MAX(na.attempted_at)::timestamptz AS last_sync
FROM notification_attempts na
    INNER JOIN notification_logs nl ON na.notification_log_id = nl.id
WHERE nl.workspace_id = $1
    AND na.attempted_at >= NOW() - INTERVAL '24 hours'
GROUP BY na.provider
ORDER BY last_sync DESC;
-- name: GetLatencyTrend :many
SELECT na.duration_ms
FROM notification_attempts na
    INNER JOIN notification_logs nl ON na.notification_log_id = nl.id
WHERE nl.workspace_id = $1
ORDER BY na.attempted_at DESC
LIMIT 20;

-- name: ListActivityLogs :many
SELECT 
    nl.id,
    nl.channel,
    nl.delivery_status,
    nl.recipient,
    nl.error_message,
    nl.template_id,
    nl.external_user_id,
    nl.trigger_data,
    nl.attempt_count,
    nl.provider_message_id,
    nl.provider_response,
    nl.created_at,
    nl.sent_at,
    nl.delivered_at,
    nl.failed_at,
    t.name as template_name,
    na.provider,
    na.duration_ms
FROM notification_logs nl
LEFT JOIN notification_attempts na ON nl.id = na.notification_log_id
LEFT JOIN templates t ON nl.template_id = t.id
WHERE nl.workspace_id = $1
  AND (sqlc.narg('channel')::text IS NULL OR nl.channel = sqlc.narg('channel'))
  AND (sqlc.narg('status')::text IS NULL OR nl.delivery_status = sqlc.narg('status'))
ORDER BY nl.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActivityLogs :one
SELECT COUNT(*)
FROM notification_logs
WHERE workspace_id = $1
  AND (sqlc.narg('channel')::text IS NULL OR channel = sqlc.narg('channel'))
  AND (sqlc.narg('status')::text IS NULL OR delivery_status = sqlc.narg('status'));