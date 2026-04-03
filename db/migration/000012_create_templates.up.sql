create table IF NOT EXISTS templates (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid REFERENCES workspaces(id) on delete CASCADE,
    environment_id uuid REFERENCES environments(id) on delete CASCADE,
    layout_id      UUID         NULL REFERENCES layouts(id) ON DELETE SET NULL,
    created_by     UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
    name varchar(255) not null,
    description text,
    event_type varchar(255) not null,
    status VARCHAR(100) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'live', 'dropped')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, environment_id, event_type)
);

CREATE INDEX templates_env_event_type_index 
    ON templates(environment_id, event_type);

CREATE INDEX templates_env_status_index 
    ON templates(environment_id, status) 
    WHERE status = 'live';
