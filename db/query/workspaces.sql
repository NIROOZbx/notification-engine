-- name: FindWorkspaceByID :one
SELECT * from workspaces where id = $1 LIMIT 1;

-- name: GetWorkspaceWithPlanName :one
SELECT 
    w.id, 
    w.name, 
    w.slug, 
    p.name as plan_name,
    w.created_at
FROM workspaces w
JOIN plans p ON w.plan_id = p.id
WHERE w.id = $1;

-- name: GetWorkspaceBySlug :one
SELECT * FROM workspaces
WHERE slug = $1
LIMIT 1;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces
WHERE id = $1;

-- name: CreateWorkspace :one

INSERT INTO workspaces (id, name, slug,plan_id) VALUES
(
    gen_random_uuid(),
    $1,
    $2,
    (SELECT id FROM plans WHERE name = 'Free' LIMIT 1)
) RETURNING * ;

-- name: UpdateWorkspaceName :one
UPDATE workspaces
SET name       = $2,
    slug       = $3
WHERE id = $1
RETURNING *;

-- name: UpdateWorkspacePlan :one

UPDATE workspaces
SET plan_id    = $2
WHERE id = $1
RETURNING *;

-- name: GetWorkspaceWithPlan :one

SELECT w.id, w.plan_id, p.api_keys_limit
FROM workspaces w
JOIN plans p ON p.id = w.plan_id
WHERE w.id = $1;