CREATE TABLE IF NOT EXISTS workspaces (
    id                   UUID         PRIMARY KEY,
    name                 VARCHAR(255) NOT NULL,
    slug                 VARCHAR(255) UNIQUE NOT NULL,   
    plan_id              UUID         NOT NULL,
    notif_count_month    INTEGER      NOT NULL DEFAULT 0,
    billing_cycle_start  DATE         NOT NULL DEFAULT CURRENT_DATE,
    created_at           TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    FOREIGN KEY (plan_id) REFERENCES plans (id) on delete cascade
);

CREATE INDEX workspaces_slug_index    ON workspaces (slug);
CREATE INDEX workspaces_plan_id_index ON workspaces (plan_id);


