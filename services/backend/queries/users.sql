
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
    is_verified
) VALUES (
    gen_random_uuid(),
    $1, $2, $3, $4, $5,
    true
)
ON CONFLICT (email) 
DO UPDATE SET 
    full_name     = EXCLUDED.full_name,
    avatar_url    = EXCLUDED.avatar_url,
    auth_provider      = EXCLUDED.auth_provider,
    provider_id   = EXCLUDED.provider_id,
    last_login_at = NOW(),
    updated_at    = NOW()
RETURNING *;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = NOW(),
    updated_at    = NOW()
WHERE id = $1;