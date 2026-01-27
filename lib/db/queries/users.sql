-- ============================================================
-- Users
-- ============================================================

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash) 
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserMFAState :one
SELECT id, is_mfa_enabled, mfa_secret, mfa_backup_codes
FROM users
WHERE id = $1;

-- name: UpdateUserMFASecret :exec
UPDATE users SET
    mfa_secret = $2,
    is_mfa_enabled = false,
    mfa_backup_codes = NULL,
    updated_at = now()
WHERE id = $1;

-- name: EnableUserMFA :exec
UPDATE users SET
    is_mfa_enabled = true,
    mfa_backup_codes = $2,
    updated_at = now()
WHERE id = $1;

-- name: DisableUserMFA :exec
UPDATE users SET
    is_mfa_enabled = false,
    mfa_secret = NULL,
    mfa_backup_codes = NULL,
    updated_at = now()
WHERE id = $1;

-- name: UpdateUser :exec
UPDATE users SET 
    email = COALESCE(sqlc.narg('email'), email),
    password_hash = COALESCE(sqlc.narg('password_hash'), password_hash),
    updated_at = now() 
WHERE id = $1;
