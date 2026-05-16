-- name: CreateNotification :exec
INSERT INTO notifications (id, user_id, title, message, type, is_read, created_at)
VALUES ($1, $2, $3, $4, $5, FALSE, NOW());

-- name: GetMyNotifications :many
SELECT * FROM notifications 
WHERE user_id = $1 OR user_id IS NULL
ORDER BY created_at DESC;

-- name: MarkNotificationAsRead :exec
UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2;

-- name: UpdateUserFCMToken :exec
UPDATE users SET fcm_token = $2, updated_at = NOW() WHERE id = $1;
