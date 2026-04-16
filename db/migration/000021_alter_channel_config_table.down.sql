DROP INDEX IF EXISTS idx_unique_active_provider;

ALTER TABLE channel_configs 
    ADD CONSTRAINT channel_configs_workspace_id_channel_provider_key 
    UNIQUE (workspace_id, channel, provider);

ALTER TABLE channel_configs DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE channel_configs 
    ALTER COLUMN credentials TYPE JSONB USING credentials::JSONB;