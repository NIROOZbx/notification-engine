ALTER TABLE notification_logs DROP CONSTRAINT IF EXISTS notification_logs_status_check;

ALTER TABLE notification_logs ADD CONSTRAINT notification_logs_status_check 
CHECK (status IN ('queued', 'processing', 'sent', 'failed', 'cancelled', 'rate_limited'));

DROP INDEX IF EXISTS idx_logs_scheduler_ready;

ALTER TABLE notification_logs 
DROP COLUMN IF EXISTS scheduled_at,
DROP COLUMN IF EXISTS next_retry_at,
DROP COLUMN IF EXISTS error_message
DROP COLUMN IF EXISTS trigger_data,
DROP COLUMN IF EXISTS updated_at;
;
