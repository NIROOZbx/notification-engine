CREATE TABLE IF NOT EXISTS user_info(
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null REFERENCES workspaces(id) on delete CASCADE,
    environment_id uuid not null REFERENCES environments(id) on delete CASCADE,
    external_user_id varchar(255) not null,
    channel VARCHAR(100) NOT NULL CHECK (
        channel IN (
            'email',
            'sms',
            'push',
            'slack',
            'whatsapp',
            'webhook',
            'in_app'
        )
    ),
    contact_value varchar(200) not null,
    metadata jsonb,
    verified boolean default false,
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    updated_at TIMESTAMPTZ NOT NULL default NOW(),
    UNIQUE(
        workspace_id,
        environment_id,
        external_user_id,
        channel
    )
);
CREATE INDEX idx_user_info_environment_channel ON user_info(environment_id, channel);