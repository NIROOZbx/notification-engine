create table if NOT EXISTS notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    template_id UUID NULL REFERENCES templates(id) ON DELETE SET NULL,
    external_user_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
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
    status VARCHAR(50) NOT NULL DEFAULT 'queued' CHECK (
            status IN (
                'queued',
                'processing',
                'sent',
                'failed',
                'cancelled',
                'rate_limited'
            )
        ),
        rendered_content JSONB NULL,
        idempotency_key varchar(255) not null,
        attempt_count  INTEGER not null default 0,
        is_test boolean not null  default false,  
        queued_at TIMESTAMPTZ NULL DEFAULT NOW(),
        recipient VARCHAR(255) NOT NULL,
        sent_at TIMESTAMPTZ NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE(workspace_id, environment_id, idempotency_key)
);


CREATE INDEX idx_logs_worker_queue on notification_logs(workspace_id,environment_id,created_at ASC) where status='queued';

CREATE INDEX idx_logs_user_history on notification_logs(workspace_id,environment_id,external_user_id,created_at DESC);

CREATE INDEX idx_logs_dashboard_filters on notification_logs(workspace_id,environment_id,status,channel,created_at DESC);