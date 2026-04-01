CREATE TABLE IF NOT EXISTS user_preferences(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    subscriber_id UUID NOT NULL REFERENCES user_info(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL CHECK (
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
    event_type varchar(255),
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(subscriber_id, channel, event_type)
);
CREATE INDEX idx_user_prefs_lookup ON user_preferences(subscriber_id,event_type)
WHERE is_enabled = false;