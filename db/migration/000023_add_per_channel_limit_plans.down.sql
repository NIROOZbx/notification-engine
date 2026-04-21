ALTER TABLE plans
ADD COLUMN notif_limit_month INTEGER NOT NULL DEFAULT 1000;

ALTER TABLE plans
DROP COLUMN email_limit_month,
DROP COLUMN sms_limit_month,
DROP COLUMN push_limit_month,
DROP COLUMN slack_limit_month,
DROP COLUMN whatsapp_limit_month,
DROP COLUMN webhook_limit_month,
DROP COLUMN in_app_limit_month;