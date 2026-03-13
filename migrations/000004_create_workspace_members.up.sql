CREATE TABLE workspace_members (
    id           UUID         PRIMARY KEY,
    workspace_id UUID         NOT NULL,
    user_id      UUID         NOT NULL,
    role         VARCHAR(255) NOT NULL DEFAULT 'member', 
    invited_by   UUID         NULL,
    joined_at    TIMESTAMP    NULL,
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    
    UNIQUE (workspace_id, user_id),
    FOREIGN KEY (workspace_id) REFERENCES workspaces (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id)      REFERENCES users (id)      ON DELETE CASCADE,
    FOREIGN KEY (invited_by)   REFERENCES users (id)      ON DELETE SET NULL
);

CREATE INDEX workspace_members_workspace_id_index ON workspace_members (workspace_id);
CREATE INDEX workspace_members_user_id_index      ON workspace_members (user_id);