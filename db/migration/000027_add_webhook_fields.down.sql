ALTER TABLE notification_logs
DROP COLUMN delivered_at,
DROP COLUMN failed_at,
DROP COLUMN provider_message_id,
DROP COLUMN delivery_status,
DROP COLUMN provider_response;
