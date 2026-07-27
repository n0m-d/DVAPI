-- name: GetCourses :many
SELECT c.id, c.instructor_id, c.title, c.slug, c.description, c.published, c.created_at, c.updated_at,
       u.full_name AS instructor_name, u.email AS email
FROM courses c
JOIN users u ON c.instructor_id = u.id
WHERE c.published = sqlc.arg(published)
  AND lower(c.title) LIKE '%' || lower(sqlc.arg(title)) || '%'
ORDER BY c.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountCourses :one
SELECT count(*)::bigint
FROM courses c
WHERE c.published = sqlc.arg(published)
  AND lower(c.title) LIKE '%' || lower(sqlc.arg(title)) || '%';

-- name: GetCourseByID :one
SELECT c.id, c.instructor_id, c.title, c.slug, c.description, c.published, c.created_at, c.updated_at,
       u.full_name AS instructor_name, u.email AS email
FROM courses c
JOIN users u ON c.instructor_id = u.id
WHERE c.id = sqlc.arg(id);

-- name: GetEnrolledCourses :many
SELECT c.id, c.instructor_id, c.title, c.slug, c.description, c.published, c.created_at, c.updated_at,
       u.full_name AS instructor_name, u.email AS email
FROM courses c
JOIN users u ON c.instructor_id = u.id
JOIN enrollments e ON c.id = e.course_id
WHERE e.user_id = sqlc.arg(student_id)
  AND lower(c.title) LIKE '%' || lower(sqlc.arg(title)) || '%'
ORDER BY e.enrolled_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountEnrolledCourses :one
SELECT count(*)::bigint
FROM courses c
JOIN enrollments e ON c.id = e.course_id
WHERE e.user_id = sqlc.arg(student_id)
  AND lower(c.title) LIKE '%' || lower(sqlc.arg(title)) || '%';

-- name: GetInstructorCourses :many
SELECT c.id, c.instructor_id, c.title, c.slug, c.description, c.published, c.created_at, c.updated_at,
       u.full_name AS instructor_name, u.email AS email
FROM courses c
JOIN users u ON c.instructor_id = u.id
WHERE c.instructor_id = sqlc.arg(instructor_id)
  AND lower(c.title) LIKE '%' || lower(sqlc.arg(title)) || '%'
ORDER BY c.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountInstructorCourses :one
SELECT count(*)::bigint
FROM courses c
WHERE c.instructor_id = sqlc.arg(instructor_id)
  AND lower(c.title) LIKE '%' || lower(sqlc.arg(title)) || '%';

-- name: CreateCourse :one
INSERT INTO courses (instructor_id, title, slug, description, published)
VALUES (sqlc.arg(instructor_id), sqlc.arg(title), sqlc.arg(slug), sqlc.arg(description), sqlc.arg(published))
RETURNING id, instructor_id, title, slug, description, published, created_at, updated_at;

-- name: UpdateCourse :one
UPDATE courses
SET title = sqlc.arg(title),
    slug = sqlc.arg(slug),
    description = sqlc.arg(description),
    published = sqlc.arg(published),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, instructor_id, title, slug, description, published, created_at, updated_at;

-- name: DeleteCourse :execrows
DELETE FROM courses WHERE id = sqlc.arg(id);

-- name: CreateLesson :one
INSERT INTO lessons (course_id, title, sort_order, content)
VALUES (sqlc.arg(course_id), sqlc.arg(title), sqlc.arg(sort_order), sqlc.arg(content))
RETURNING id, course_id, title, sort_order, content, created_at, updated_at;

-- name: GetLessonByID :one
SELECT id, course_id, title, sort_order, content, created_at, updated_at
FROM lessons
WHERE id = sqlc.arg(id);

-- name: ListCourseLessons :many
SELECT id, course_id, title, sort_order, content, created_at, updated_at
FROM lessons
WHERE course_id = sqlc.arg(course_id)
ORDER BY sort_order ASC, created_at ASC;

-- name: UpdateLesson :one
UPDATE lessons
SET title = sqlc.arg(title),
    sort_order = sqlc.arg(sort_order),
    content = sqlc.arg(content),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, course_id, title, sort_order, content, created_at, updated_at;

-- name: DeleteLesson :execrows
DELETE FROM lessons WHERE id = sqlc.arg(id);

-- name: DeleteEnrollment :execrows
DELETE FROM enrollments
WHERE user_id = sqlc.arg(user_id) AND course_id = sqlc.arg(course_id);

-- name: ListInstructors :many
SELECT id, email, password_hash, full_name, role, created_at, updated_at
FROM users
WHERE role = 'instructor'
ORDER BY created_at ASC;

-- name: ListStudents :many
SELECT id, email, password_hash, full_name, role, created_at, updated_at
FROM users
WHERE role = 'student'
ORDER BY created_at ASC;

-- name: ListEnrolledStudentsByCourse :many
SELECT u.id, u.email, u.full_name, u.role, u.created_at, u.updated_at, e.enrolled_at
FROM enrollments e
JOIN users u ON e.user_id = u.id
WHERE e.course_id = sqlc.arg(course_id)
  AND lower(u.full_name) LIKE '%' || lower(sqlc.arg(name)) || '%'
ORDER BY e.enrolled_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountEnrolledStudentsByCourse :one
SELECT count(*)::bigint
FROM enrollments e
JOIN users u ON e.user_id = u.id
WHERE e.course_id = sqlc.arg(course_id)
  AND lower(u.full_name) LIKE '%' || lower(sqlc.arg(name)) || '%';
