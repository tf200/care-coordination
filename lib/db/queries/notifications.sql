-- name: CreateNotification :one
INSERT INTO notifications (
    id,
    user_id,
    type,
    priority,
    title,
    message,
    resource_type,
    resource_id,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetNotification :one
SELECT * FROM notifications
WHERE id = $1;

-- name: ListNotifications :many
SELECT 
    *,
    COUNT(*) OVER() as total_count
FROM notifications
WHERE user_id = $1
    AND (sqlc.narg('is_read')::boolean IS NULL OR is_read = sqlc.narg('is_read')::boolean)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUnreadCount :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1 AND is_read = FALSE;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET is_read = TRUE, read_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET is_read = TRUE, read_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND is_read = FALSE;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: DeleteExpiredNotifications :exec
DELETE FROM notifications
WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP;
