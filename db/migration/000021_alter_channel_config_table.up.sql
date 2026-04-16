ALTER TABLE channel_configs
    ALTER COLUMN credentials TYPE TEXT USING credentials::TEXT;

ALTER TABLE channel_configs ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE channel_configs DROP CONSTRAINT channel_configs_workspace_id_channel_provider_key;

CREATE UNIQUE INDEX idx_unique_active_provider 
ON channel_configs(workspace_id, channel, provider) 
WHERE deleted_at IS NULL;