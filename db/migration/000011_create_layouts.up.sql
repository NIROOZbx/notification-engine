CREATE TABLE IF NOT EXISTS layouts (
    id uuid primary key not null default gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE not null,
    name VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    html TEXT NOT NULL,
    created_at TIMESTAMPTZ default NOW(),
    updated_at TIMESTAMPTZ default NOW()
);
create unique INDEX default_layout on layouts(workspace_id)
WHERE is_default = true;