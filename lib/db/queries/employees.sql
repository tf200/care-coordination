-- name: CreateEmployee :exec
INSERT INTO employees (
    id,
    user_id,
    first_name,
    last_name,
    bsn,
    date_of_birth,
    phone_number,
    gender
) VALUES (
 $1, $2, $3, $4, $5, $6, $7, $8
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
    COUNT(*) OVER() as total_count
FROM employees e
WHERE
(
  sqlc.narg('search')::text IS NULL OR
  e.first_name ILIKE '%' || sqlc.narg('search')::text || '%' OR
  e.last_name ILIKE '%' || sqlc.narg('search')::text || '%' OR
  CONCAT(e.first_name, ' ', e.last_name) ILIKE '%' || sqlc.narg('search')::text || '%'
)
ORDER BY e.first_name, e.last_name
LIMIT $1 OFFSET $2;

-- name: GetEmployeeByUserID :one
SELECT e.*, u.email,
       r.id as role_id,
       r.name as role_name
FROM employees e
JOIN users u ON e.user_id = u.id
LEFT JOIN user_roles ur ON e.user_id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
WHERE e.user_id = $1 LIMIT 1;