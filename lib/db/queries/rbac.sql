-- ============================================================
-- Roles
-- ============================================================

-- name: CreateRole :one
INSERT INTO roles (id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT 
    r.*,
    COUNT(*) OVER() as total_count,
    (SELECT COUNT(*) FROM role_permissions rp WHERE rp.role_id = r.id) as permission_count,
    (SELECT COUNT(*) FROM user_roles ur WHERE ur.role_id = r.id) as user_count
FROM roles r
ORDER BY r.name
LIMIT $1 OFFSET $2;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- ============================================================
-- Permissions
-- ============================================================

-- name: CreatePermission :one
INSERT INTO permissions (id, resource, action, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPermissionByID :one
SELECT * FROM permissions WHERE id = $1;

-- name: ListPermissions :many
SELECT *, COUNT(*) OVER() as total_count
FROM permissions
ORDER BY resource, action
LIMIT $1 OFFSET $2;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;

-- ============================================================
-- Role Permissions
-- ============================================================

-- name: AssignPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: BatchAssignPermissionsToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
SELECT @role_id::text, UNNEST(@permission_ids::text[])
ON CONFLICT DO NOTHING;

-- name: DeleteAllPermissionsFromRole :exec
DELETE FROM role_permissions WHERE role_id = $1;

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: ListPermissionsForRole :many
SELECT p.*
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.resource, p.action;

-- ============================================================
-- User Roles
-- ============================================================

-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET role_id = $2, assigned_at = CURRENT_TIMESTAMP;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1;

-- name: GetRoleForUser :one
SELECT r.*
FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: ListUsersWithRole :many
SELECT u.id, u.email
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role_id = $1
ORDER BY u.email;
