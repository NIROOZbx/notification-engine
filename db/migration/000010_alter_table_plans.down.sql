ALTER TABLE plans
    DROP COLUMN IF EXISTS max_layouts,
    DROP COLUMN IF EXISTS max_templates;