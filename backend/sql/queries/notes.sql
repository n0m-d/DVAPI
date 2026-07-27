-- name: CreateNote :one
INSERT INTO notes (user_id, title, body)
VALUES (sqlc.arg(user_id), sqlc.arg(title), sqlc.arg(body))
RETURNING id, user_id, title, body, created_at, updated_at;

-- name: ListNotes :many
SELECT id, user_id, title, body, created_at, updated_at
FROM notes
WHERE (sqlc.narg(user_id)::uuid IS NULL OR user_id = sqlc.narg(user_id))
ORDER BY created_at DESC;

-- name: UpdateNote :one
UPDATE notes
SET
    title = COALESCE(sqlc.narg(title), title),
    body = COALESCE(sqlc.narg(body), body),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, user_id, title, body, created_at, updated_at;

-- name: DeleteNote :execrows
DELETE FROM notes
WHERE id = sqlc.arg(id);
