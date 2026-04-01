ALTER TABLE plans
ADD COLUMN max_layouts INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN max_templates INTEGER NOT NULL DEFAULT 10;
    
UPDATE plans
SET max_layouts = 1,
    max_templates = 10
WHERE name = 'Free';

UPDATE plans
SET max_layouts = 5,
    max_templates = 100
WHERE name = 'Pro';
UPDATE plans
SET max_layouts = 50,
    max_templates = 1000
WHERE name = 'Enterprise';