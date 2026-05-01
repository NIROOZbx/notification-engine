ALTER TABLE notification_logs
ADD COLUMN delivery_status     VARCHAR(50) DEFAULT 'pending',
ADD COLUMN delivered_at        TIMESTAMPTZ,
ADD COLUMN failed_at           TIMESTAMPTZ,
ADD COLUMN provider_message_id TEXT,
ADD COLUMN provider_response   TEXT;