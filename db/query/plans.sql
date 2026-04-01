-- name: GetPlanByID :one
SELECT *
FROM plans
WHERE id = $1
LIMIT 1;

-- name: GetAllPlans :many
SELECT *
FROM plans
WHERE is_active = true
ORDER BY price_cents ASC;

-- name: GetPlanByWorkspace :one
SELECT p.*
from plans as p
    join workspaces as w on w.plan_id = p.id
where w.id = $1
    and is_active = true;

-- name: CountWorkspaceLayouts :one
select count(*)
from layouts
where workspace_id = $1;

-- name: CountWorkspaceTemplates :one
SELECT COUNT(*)
FROM templates
WHERE workspace_id = $1;