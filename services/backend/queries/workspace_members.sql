-- name: CreateWorkspaceMember :one
INSERT INTO workspace_members (id, workspace_id, user_id, role, joined_at)
VALUES (gen_random_uuid(), $1, $2, $3, NOW())
RETURNING *;

-- name: GetWorkspaceMemberByUserID :one

select * from workspace_members where user_id = $1 limit 1;

-- name: GetWorkspaceMemberByID :one

select * from workspace_members where id = $1 limit 1;

-- name: GetWorkspaceMembers :many
SELECT * FROM workspace_members
WHERE workspace_id = $1;

-- name: UpdateMemberRole :one
UPDATE workspace_members
SET role = $2
WHERE workspace_id = $1
AND user_id = $3
RETURNING *;

-- name: DeleteWorkspaceMember :exec
DELETE FROM workspace_members
WHERE workspace_id = $1
AND user_id = $2;

-- name: GetMemberRole :one
SELECT role FROM workspace_members
WHERE workspace_id = $1
AND user_id = $2
LIMIT 1;