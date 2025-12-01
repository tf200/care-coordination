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
SELECT * FROM employees;