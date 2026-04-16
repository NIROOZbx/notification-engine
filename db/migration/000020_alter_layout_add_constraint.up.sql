ALTER TABLE layouts
ADD CONSTRAINT layouts_workspace_id_name_key UNIQUE(workspace_id, name);