
-- name: FindUserByProviderID :one
SELECT * FROM users
WHERE auth_provider = $1
AND provider_id = $2
LIMIT 1;

-- name: FindUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: FindUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: UpsertOAuthUser :one
INSERT INTO users (
    id,
    email,
    full_name,
    auth_provider,
    provider_id,
    avatar_url,
    is_verified,
    last_login_at
) VALUES (
    gen_random_uuid(),
    $1, $2, $3, $4, $5,
    true,
    NOW()
)
ON CONFLICT (email) 
DO UPDATE SET 
    full_name     = EXCLUDED.full_name,
    avatar_url    = EXCLUDED.avatar_url,
    auth_provider      = EXCLUDED.auth_provider,
    provider_id   = EXCLUDED.provider_id,
    last_login_at = NOW() 
RETURNING *;


-- name: CreateUser :one

INSERT INTO users (id,email, full_name, password_hash, auth_provider)
VALUES (gen_random_uuid(),$1, $2, $3, 'local')
RETURNING *;

-- name: GetUserAuthContext :one
SELECT 
    u.id, 
    COALESCE(m.role, 'member') as role,
    COALESCE(m.workspace_id, '00000000-0000-0000-0000-000000000000')::uuid as workspace_id
FROM users u
LEFT JOIN workspace_members m ON u.id = m.user_id
WHERE u.id = $1 
LIMIT 1;

-- name: GetUserWithWorkspace :one
SELECT 
    u.id AS user_id, 
    u.full_name, 
    u.email, 
    u.avatar_url,
    COALESCE(m.role, 'member')::varchar AS role,
    COALESCE(w.id, '00000000-0000-0000-0000-000000000000')::uuid AS workspace_id,
    COALESCE(w.name, '')::varchar AS workspace_name,
    COALESCE(w.slug, '')::varchar AS slug
FROM users u
LEFT JOIN workspace_members m ON u.id = m.user_id
LEFT JOIN workspaces w ON m.workspace_id = w.id
WHERE u.id = $1; 

-- name: GetAuthContextByEmail :one
SELECT sqlc.embed(u),
    m.role,
    w.id AS workspace_id,
    w.name AS workspace_name,
    w.slug AS workspace_slug
FROM users as u
LEFT JOIN workspace_members m ON u.id = m.user_id
LEFT JOIN workspaces w ON m.workspace_id = w.id
WHERE u.email = $1 
LIMIT 1;