DROP TRIGGER IF EXISTS set_updated_at ON workspaces;
DROP TRIGGER IF EXISTS set_updated_at ON users;
DROP TRIGGER IF EXISTS set_updated_at ON workspace_members;

DROP FUNCTION IF EXISTS update_updated_at();