-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, body, course_id)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(type),
    sqlc.arg(title),
    sqlc.arg(body),
    sqlc.narg(course_id)
)
RETURNING id, user_id, type, title, body, read, course_id, created_at;

-- name: CreateNotifications :many
INSERT INTO notifications (user_id, type, title, body, course_id)
SELECT
    u.user_id,
    sqlc.arg(type),
    sqlc.arg(title),
    sqlc.arg(body),
    sqlc.narg(course_id)
FROM unnest(sqlc.arg(user_ids)::uuid[]) AS u(user_id)
RETURNING id, user_id, type, title, body, read, course_id, created_at;

-- name: GetNotificationByID :one
SELECT id, user_id, type, title, body, read, course_id, created_at
FROM notifications
WHERE id = sqlc.arg(id);

-- name: ListNotificationsByUser :many
SELECT id, user_id, type, title, body, read, course_id, created_at
FROM notifications
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.arg(unread_only)::boolean = false OR read = false)
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountNotificationsByUser :one
SELECT count(*)::bigint
FROM notifications
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.arg(unread_only)::boolean = false OR read = false);

-- name: CountUnreadNotificationsByUser :one
SELECT count(*)::bigint
FROM notifications
WHERE user_id = sqlc.arg(user_id)
  AND read = false;

-- name: MarkNotificationRead :one
UPDATE notifications
SET read = true
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, type, title, body, read, course_id, created_at;

-- name: MarkAllNotificationsRead :execrows
UPDATE notifications
SET read = true
WHERE user_id = sqlc.arg(user_id)
  AND read = false;

-- name: DeleteNotification :execrows
DELETE FROM notifications
WHERE id = sqlc.arg(id);
