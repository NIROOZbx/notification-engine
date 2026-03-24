CREATE TABLE environments (
    id           UUID         NOT NULL DEFAULT gen_random_uuid(),
    workspace_id UUID         NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name         VARCHAR(100) NOT NULL CHECK (name IN ('development', 'production')),
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    UNIQUE (workspace_id, name)
);