-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    id, user_id, employee_id, client_id, action, resource_type, resource_id,
    old_value, new_value, ip_address, user_agent, request_id, status, failure_reason,
    prev_hash, current_hash
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
);

-- name: GetLatestAuditLog :one
-- Get the most recent audit log entry to retrieve its hash for the chain
SELECT id, current_hash, sequence_number 
FROM audit_logs 
ORDER BY sequence_number DESC 
LIMIT 1;

-- name: ListAuditLogs :many
SELECT 
    al.*,
    u.email as user_email,
    COALESCE(e.first_name || ' ' || e.last_name, '') as employee_name,
    COALESCE(c.first_name || ' ' || c.last_name, '') as client_name,
    COUNT(*) OVER() as total_count
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
LEFT JOIN employees e ON al.employee_id = e.id
LEFT JOIN clients c ON al.client_id = c.id
WHERE 
    (sqlc.narg(user_id)::TEXT IS NULL OR al.user_id = sqlc.narg(user_id))
    AND (sqlc.narg(client_id)::TEXT IS NULL OR al.client_id = sqlc.narg(client_id))
    AND (sqlc.narg(resource_type)::TEXT IS NULL OR al.resource_type = sqlc.narg(resource_type))
    AND (sqlc.narg(resource_id)::TEXT IS NULL OR al.resource_id = sqlc.narg(resource_id))
    AND (sqlc.narg(action)::audit_action_enum IS NULL OR al.action = sqlc.narg(action))
    AND (sqlc.narg(start_date)::TIMESTAMP IS NULL OR al.created_at >= sqlc.narg(start_date))
    AND (sqlc.narg(end_date)::TIMESTAMP IS NULL OR al.created_at <= sqlc.narg(end_date))
ORDER BY al.sequence_number DESC
LIMIT $1 OFFSET $2;

-- name: GetAuditLogsByResource :many
SELECT * FROM audit_logs
WHERE resource_type = $1 AND resource_id = $2
ORDER BY sequence_number DESC
LIMIT $3;

-- name: GetAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE user_id = $1
ORDER BY sequence_number DESC
LIMIT $2 OFFSET $3;

-- name: GetAuditLogStats :one
SELECT 
    COUNT(*) as total_logs,
    COUNT(*) FILTER (WHERE action = 'read') as read_count,
    COUNT(*) FILTER (WHERE action = 'create') as create_count,
    COUNT(*) FILTER (WHERE action = 'update') as update_count,
    COUNT(*) FILTER (WHERE action = 'delete') as delete_count,
    COUNT(*) FILTER (WHERE status = 'failure') as failure_count
FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '24 hours';

-- name: GetAuditLogsForVerification :many
-- Get audit logs in sequence order for hash chain verification
SELECT id, sequence_number, user_id, employee_id, action, resource_type, resource_id,
       old_value, new_value, ip_address, user_agent, request_id, status, failure_reason,
       prev_hash, current_hash, created_at
FROM audit_logs
WHERE sequence_number >= $1 AND sequence_number <= $2
ORDER BY sequence_number ASC;

-- name: GetAuditLogBySequence :one
SELECT * FROM audit_logs WHERE sequence_number = $1;

-- name: CountAuditLogs :one
SELECT COUNT(*) as total FROM audit_logs;

-- name: GetAuditLogByID :one
SELECT 
    al.*,
    u.email as user_email,
    COALESCE(e.first_name || ' ' || e.last_name, '') as employee_name,
    COALESCE(c.first_name || ' ' || c.last_name, '') as client_name
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
LEFT JOIN employees e ON al.employee_id = e.id
LEFT JOIN clients c ON al.client_id = c.id
WHERE al.id = $1;

