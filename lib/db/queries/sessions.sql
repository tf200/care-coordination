-- name: CreateUserSession :exec
INSERT INTO sessions (
    id,
    user_id,
    token_hash,
    token_family,
    expires_at,
    user_agent,
    ip_address
) 
VALUES ($1, $2, $3, $4, $5, $6, $7);



-- name: GetUserSession :one
SELECT * FROM sessions WHERE token_hash = $1;


-- name: UpdateUserSession :exec
UPDATE sessions SET
    expires_at = $1,
    user_agent = $2,
    ip_address = $3,
    token_family = $4,
    token_hash = $5
WHERE id = $6;

-- name: DeleteUserSession :exec
DELETE FROM sessions WHERE token_hash = $1;