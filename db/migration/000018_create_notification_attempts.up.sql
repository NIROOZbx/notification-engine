CREATE TABLE if NOT EXISTS notification_attempts (
    id uuid primary key default gen_random_uuid(),
    notification_log_id uuid not null REFERENCES notification_logs(id) on delete cascade,
    attempt_count INTEGER not null default 1,
    provider varchar(100) not null ,
    channel_config_id uuid REFERENCES channel_configs(id) on delete cascade,
    status varchar(255) not null,
    error_message text ,
    error_code varchar(100),
    provider_message_id varchar(255),
    duration_ms INTEGER,
    attempted_at timestamptz default now()

);

CREATE INDEX idx_attempts_log_id on notification_attempts(notification_log_id,attempted_at desc);

CREATE INDEX idx_attempts_status_provider on notification_attempts(status,provider,attempted_at DESC);



