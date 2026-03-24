-- sql/queries/plans.sql

-- name: GetPlanByID :one
SELECT * FROM plans
WHERE id = $1
LIMIT 1;

-- name: GetAllPlans :many
SELECT * FROM plans
WHERE is_active = true
ORDER BY price_cents ASC;