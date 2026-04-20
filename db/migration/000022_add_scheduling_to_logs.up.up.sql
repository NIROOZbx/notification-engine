ALTER TABLE notification_logs 
ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ NULL,
ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ NULL,
ADD COLUMN IF NOT EXISTS error_message TEXT NULL,
ADD COLUMN IF NOT EXISTS trigger_data  JSONB NULL,
ADD COLUMN IF NOT EXISTS updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE notification_logs DROP CONSTRAINT IF EXISTS notification_logs_status_check;

ALTER TABLE notification_logs ADD CONSTRAINT notification_logs_status_check 
CHECK (status IN ('queued', 'processing', 'sent', 'failed', 'cancelled', 'rate_limited', 'retrying', 'scheduled'));

CREATE INDEX IF NOT EXISTS idx_logs_scheduler_ready ON notification_logs (next_retry_at, scheduled_at) 
WHERE status IN ('retrying', 'scheduled');
