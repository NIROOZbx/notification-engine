CREATE TABLE IF NOT EXISTS notification_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    notification_log_id UUID NOT NULL REFERENCES notification_logs(id) ON DELETE CASCADE,
    channel VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    provider_event_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(
        notification_log_id,
        event_type,
        provider_event_id
    )
);
CREATE INDEX idx_events_analytics_filter ON notification_events(
    workspace_id,
    environment_id,
    channel,
    event_type,
    created_at DESC
);
CREATE INDEX idx_events_log_lookup ON notification_events(notification_log_id, created_at ASC);