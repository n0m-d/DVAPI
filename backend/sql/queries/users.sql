-- name: GetUserByID :one
SELECT id, email, password_hash, full_name, role, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name, role, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name, role)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, full_name, role, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2,
    updated_at = now()
WHERE id = $1;

-- name: ListUsers :many
SELECT id, email, password_hash, full_name, role, created_at, updated_at
FROM users
WHERE (sqlc.arg(search)::text = ''
       OR lower(email) LIKE '%' || lower(sqlc.arg(search)) || '%'
       OR lower(full_name) LIKE '%' || lower(sqlc.arg(search)) || '%')
  AND (sqlc.arg(role)::text = '' OR role::text = sqlc.arg(role))
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountUsers :one
SELECT count(*)::bigint
FROM users
WHERE (sqlc.arg(search)::text = ''
       OR lower(email) LIKE '%' || lower(sqlc.arg(search)) || '%'
       OR lower(full_name) LIKE '%' || lower(sqlc.arg(search)) || '%')
  AND (sqlc.arg(role)::text = '' OR role::text = sqlc.arg(role));

-- name: UpdateUser :one
UPDATE users
SET email = sqlc.arg(email),
    full_name = sqlc.arg(full_name),
    role = sqlc.arg(role),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, email, password_hash, full_name, role, created_at, updated_at;

-- name: GetAdminStats :one
SELECT
    (SELECT count(*) FROM users)::bigint AS users,
    (SELECT count(*) FROM users WHERE role = 'student')::bigint AS students,
    (SELECT count(*) FROM users WHERE role = 'instructor')::bigint AS instructors,
    (SELECT count(*) FROM courses)::bigint AS courses,
    (SELECT count(*) FROM enrollments)::bigint AS enrollments,
    (SELECT count(*) FROM assignments)::bigint AS assignments,
    (SELECT count(*) FROM assignment_submissions)::bigint AS submissions;

-- name: ListPublishedCourses :many
SELECT id, instructor_id, title, slug, description, published, created_at, updated_at
FROM courses
WHERE published = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetCourseBySlug :one
SELECT id, instructor_id, title, slug, description, published, created_at, updated_at
FROM courses
WHERE slug = $1;

-- name: ListLessonsByCourse :many
SELECT id, course_id, title, sort_order, content, created_at, updated_at
FROM lessons
WHERE course_id = $1
ORDER BY sort_order ASC;

-- name: CreateEnrollment :one
INSERT INTO enrollments (user_id, course_id)
VALUES ($1, $2)
RETURNING id, user_id, course_id, enrolled_at;

-- name: GetEnrollment :one
SELECT id, user_id, course_id, enrolled_at
FROM enrollments
WHERE user_id = $1 AND course_id = $2;
