create TABLE if NOT EXISTS channel_configs(
    id uuid primary key default gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    channel VARCHAR(100) NOT NULL CHECK (channel IN (
    'email',
    'sms', 
    'push',
    'slack',
    'whatsapp',
    'webhook',
    'in_app'
)),
    provider varchar(150) not null,
    display_name varchar(255),
    credentials jsonb not null,
    is_active boolean not null default false,
    is_default boolean not null default false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ not null DEFAULT NOW(),
    UNIQUE(workspace_id, channel, provider)
);
CREATE UNIQUE INDEX idx_unique_default_channel ON channel_configs(workspace_id, channel)
WHERE is_default = true;