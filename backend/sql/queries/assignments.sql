-- name: CreateAssignment :one
INSERT INTO assignments (course_id, title, description, due_date, status, created_by)
VALUES (
    sqlc.arg(course_id),
    sqlc.arg(title),
    sqlc.arg(description),
    sqlc.arg(due_date),
    sqlc.arg(status),
    sqlc.arg(created_by)
)
RETURNING id, course_id, title, description, due_date, status, created_at, updated_at, created_by;

-- name: GetAssignmentByID :one
SELECT id, course_id, title, description, due_date, status, created_at, updated_at, created_by
FROM assignments
WHERE id = sqlc.arg(id);

-- name: UpdateAssignment :one
UPDATE assignments
SET title = sqlc.arg(title),
    description = sqlc.arg(description),
    due_date = sqlc.arg(due_date),
    status = sqlc.arg(status),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, course_id, title, description, due_date, status, created_at, updated_at, created_by;

-- name: DeleteAssignment :execrows
DELETE FROM assignments WHERE id = sqlc.arg(id);

-- name: ListPublishedAssignmentsByCourse :many
SELECT id, course_id, title, description, due_date, status, created_at, updated_at, created_by
FROM assignments
WHERE course_id = sqlc.arg(course_id)
  AND status = 'published'
ORDER BY due_date ASC;

-- name: ListAssignmentsByCourse :many
SELECT id, course_id, title, description, due_date, status, created_at, updated_at, created_by
FROM assignments
WHERE course_id = sqlc.arg(course_id)
  AND lower(title) LIKE '%' || lower(sqlc.arg(title)) || '%'
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountAssignmentsByCourse :one
SELECT count(*)::bigint
FROM assignments
WHERE course_id = sqlc.arg(course_id)
  AND lower(title) LIKE '%' || lower(sqlc.arg(title)) || '%';

-- name: GetAssignmentByCourseAndTitle :one
SELECT id, course_id, title, description, due_date, status, created_at, updated_at, created_by
FROM assignments
WHERE course_id = sqlc.arg(course_id)
  AND lower(title) = lower(sqlc.arg(title));

-- name: CreateAssignmentSubmission :one
INSERT INTO assignment_submissions (assignment_id, student_id, submission_text, file_path, file_name)
VALUES (
    sqlc.arg(assignment_id),
    sqlc.arg(student_id),
    sqlc.arg(submission_text),
    sqlc.arg(file_path),
    sqlc.arg(file_name)
)
RETURNING id, assignment_id, student_id, submission_text, file_path, file_name, submitted_at, grade, feedback, created_at, updated_at;

-- name: GetSubmissionByAssignmentAndStudent :one
SELECT id, assignment_id, student_id, submission_text, file_path, file_name, submitted_at, grade, feedback, created_at, updated_at
FROM assignment_submissions
WHERE assignment_id = sqlc.arg(assignment_id)
  AND student_id = sqlc.arg(student_id);

-- name: GetSubmissionByID :one
SELECT id, assignment_id, student_id, submission_text, file_path, file_name, submitted_at, grade, feedback, created_at, updated_at
FROM assignment_submissions
WHERE id = sqlc.arg(id);

-- name: GradeSubmission :one
UPDATE assignment_submissions
SET grade = sqlc.arg(grade),
    feedback = sqlc.arg(feedback),
    graded_at = now(),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, assignment_id, student_id, submission_text, file_path, file_name, submitted_at, grade, feedback, created_at, updated_at;

-- name: ListSubmissionsByAssignment :many
SELECT s.id, s.assignment_id, s.student_id, s.submission_text, s.file_path, s.file_name, s.submitted_at, s.grade, s.feedback, s.created_at, s.updated_at,
       u.full_name AS student_name, u.email AS student_email
FROM assignment_submissions s
JOIN users u ON s.student_id = u.id
WHERE s.assignment_id = sqlc.arg(assignment_id)
  AND lower(u.full_name) LIKE '%' || lower(sqlc.arg(name)) || '%'
ORDER BY s.submitted_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountSubmissionsByAssignment :one
SELECT count(*)::bigint
FROM assignment_submissions s
JOIN users u ON s.student_id = u.id
WHERE s.assignment_id = sqlc.arg(assignment_id)
  AND lower(u.full_name) LIKE '%' || lower(sqlc.arg(name)) || '%';

-- name: CloseOverdueAssignments :many
UPDATE assignments
SET status = 'closed',
    updated_at = now()
WHERE due_date < now()
  AND status = 'published'
RETURNING id, course_id, title, description, due_date, status, created_at, updated_at, created_by;
