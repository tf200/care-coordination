-- ============================================================
-- Users
-- ============================================================

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash) 
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :exec
UPDATE users SET 
    email = COALESCE(sqlc.narg('email'), email),
    password_hash = COALESCE(sqlc.narg('password_hash'), password_hash),
    updated_at = now() 
WHERE id = $1;
