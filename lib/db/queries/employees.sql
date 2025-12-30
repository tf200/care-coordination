-- ============================================================
-- Employees
-- ============================================================

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
    contract_hours,
    contract_type,
    location_id
) VALUES (
 $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
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
    e.contract_hours,
    e.contract_type,
    e.location_id,
    l.name as location_name,
    u.email,
    COALESCE(COUNT(DISTINCT c.id), 0) as client_count,
    COUNT(*) OVER() as total_count
FROM employees e
JOIN locations l ON e.location_id = l.id
JOIN users u ON e.user_id = u.id
LEFT JOIN clients c ON c.coordinator_id = e.id
WHERE
(
  sqlc.narg('search')::text IS NULL OR
  e.first_name ILIKE '%' || sqlc.narg('search')::text || '%' OR
  e.last_name ILIKE '%' || sqlc.narg('search')::text || '%' OR
  CONCAT(e.first_name, ' ', e.last_name) ILIKE '%' || sqlc.narg('search')::text || '%'
)
GROUP BY e.id, e.user_id, e.first_name, e.last_name, e.bsn, e.date_of_birth, 
         e.phone_number, e.gender, e.contract_hours, e.contract_type, e.location_id,
         l.name, u.email
ORDER BY e.first_name, e.last_name
LIMIT $1 OFFSET $2;

-- name: GetEmployeeByID :one
SELECT
    e.id,
    e.user_id,
    e.first_name,
    e.last_name,
    e.bsn,
    e.date_of_birth,
    e.phone_number,
    e.gender,
    e.contract_hours,
    e.contract_type,
    e.location_id,
    l.name as location_name,
    u.email,
    r.id as role_id,
    r.name as role_name,
    COALESCE(COUNT(DISTINCT c.id), 0) as client_count
FROM employees e
JOIN locations l ON e.location_id = l.id
JOIN users u ON e.user_id = u.id
LEFT JOIN user_roles ur ON e.user_id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
LEFT JOIN clients c ON c.coordinator_id = e.id
WHERE e.id = $1
GROUP BY e.id, e.user_id, e.first_name, e.last_name, e.bsn, e.date_of_birth, 
         e.phone_number, e.gender, e.contract_hours, e.contract_type, e.location_id,
         l.name, u.email, r.id, r.name
LIMIT 1;

-- name: GetEmployeeByUserID :one
SELECT e.*, u.email,
       r.id as role_id,
       r.name as role_name,
       l.name as location_name
FROM employees e
JOIN users u ON e.user_id = u.id
LEFT JOIN user_roles ur ON e.user_id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
LEFT JOIN locations l ON e.location_id = l.id
WHERE e.user_id = $1 LIMIT 1;

-- name: UpdateEmployee :exec
UPDATE employees SET
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    bsn = COALESCE(sqlc.narg('bsn'), bsn),
    date_of_birth = COALESCE(sqlc.narg('date_of_birth'), date_of_birth),
    phone_number = COALESCE(sqlc.narg('phone_number'), phone_number),
    gender = COALESCE(sqlc.narg('gender'), gender),
    contract_hours = COALESCE(sqlc.narg('contract_hours'), contract_hours),
    contract_type = COALESCE(sqlc.narg('contract_type'), contract_type),
    location_id = COALESCE(sqlc.narg('location_id'), location_id),
    updated_at = now()
WHERE id = $1;

-- name: SoftDeleteEmployee :exec
UPDATE employees SET is_deleted = true, updated_at = now() WHERE id = $1;