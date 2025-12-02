-- name: CreateEmployee :exec
INSERT INTO employees (
    id,
    user_id,
    first_name,
    last_name,
    bsn,
    date_of_birth,
    phone_number,
    gender,
    role
) VALUES (
 $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: ListEmployees :many
SELECT
    e.id,
    e.user_id,
    e.first_name,
    e.last_name,
    e.bsn,
    e.date_of_birth,
    e.phone_number,
    e.gender,
    e.role,
    COUNT(*) OVER() as total_count
FROM employees e
WHERE
    (sqlc.narg('search')::text IS NULL OR
     LOWER(e.first_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
     LOWER(e.last_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
     LOWER(CONCAT(e.first_name, ' ', e.last_name)) LIKE LOWER('%' || sqlc.narg('search')::text || '%'))
ORDER BY e.first_name, e.last_name
LIMIT $1 OFFSET $2;