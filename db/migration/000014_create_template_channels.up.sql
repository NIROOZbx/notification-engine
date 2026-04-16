CREATE TABLE IF NOT EXISTS template_channels(
    id uuid primary key default gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    channel_config_id uuid null REFERENCES channel_configs(id) on delete SET NULL,
    channel VARCHAR(100) NOT NULL CHECK (channel IN (
    'email',
    'sms',
    'push',
    'slack',
    'whatsapp',
    'webhook',
    'in_app'
)),
    content JSONB NOT NULL,
    is_active BOOLEAN default true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(template_id, channel)
);

