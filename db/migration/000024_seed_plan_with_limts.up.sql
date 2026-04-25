UPDATE plans
SET email_limit_month = 1000,
    sms_limit_month = 100,
    push_limit_month = 10000,
    slack_limit_month = 500,
    whatsapp_limit_month = 50,
    webhook_limit_month = 1000,
    in_app_limit_month = 10000
WHERE name = 'Free';

UPDATE plans
SET email_limit_month = 50000,
    sms_limit_month = 2000,
    push_limit_month = 500000,
    slack_limit_month = 10000,
    whatsapp_limit_month = 1000,
    webhook_limit_month = 50000,
    in_app_limit_month = 500000
WHERE name = 'Pro';

UPDATE plans
SET email_limit_month = -1,
    sms_limit_month = -1,
    push_limit_month = -1,
    slack_limit_month = -1,
    whatsapp_limit_month = -1,
    webhook_limit_month = -1,
    in_app_limit_month = -1
WHERE name= 'Enterprise';