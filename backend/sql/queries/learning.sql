-- name: UpsertLessonProgress :one
INSERT INTO lesson_progress (student_id, lesson_id, completed, completed_at)
VALUES (
    sqlc.arg(student_id),
    sqlc.arg(lesson_id),
    sqlc.arg(completed),
    CASE WHEN sqlc.arg(completed)::boolean THEN now() ELSE NULL END
)
ON CONFLICT (student_id, lesson_id) DO UPDATE
SET completed = EXCLUDED.completed,
    completed_at = CASE WHEN EXCLUDED.completed THEN now() ELSE NULL END,
    updated_at = now()
RETURNING id, student_id, lesson_id, completed, completed_at, updated_at;

-- name: GetCourseProgress :one
SELECT
    count(l.id)::bigint AS total_lessons,
    count(l.id) FILTER (WHERE lp.completed = true)::bigint AS completed_lessons
FROM lessons l
LEFT JOIN lesson_progress lp
  ON lp.lesson_id = l.id AND lp.student_id = sqlc.arg(student_id)
WHERE l.course_id = sqlc.arg(course_id);

-- name: GetNextIncompleteLesson :one
SELECT l.id, l.course_id, l.title, l.sort_order, l.content, l.created_at, l.updated_at
FROM lessons l
LEFT JOIN lesson_progress lp
  ON lp.lesson_id = l.id AND lp.student_id = sqlc.arg(student_id)
WHERE l.course_id = sqlc.arg(course_id)
  AND COALESCE(lp.completed, false) = false
ORDER BY l.sort_order ASC, l.created_at ASC
LIMIT 1;

-- name: ListStudentGrades :many
SELECT s.id AS submission_id, s.assignment_id, a.title AS assignment_title,
       a.course_id, c.title AS course_title, s.grade, s.feedback, s.submitted_at,
       s.updated_at
FROM assignment_submissions s
JOIN assignments a ON a.id = s.assignment_id
JOIN courses c ON c.id = a.course_id
WHERE s.student_id = sqlc.arg(student_id)
ORDER BY s.updated_at DESC;

-- name: GetStudentGradeSummary :one
SELECT count(*)::bigint AS submitted,
       COALESCE(avg(grade), 0)::float8 AS average_grade
FROM assignment_submissions
WHERE student_id = sqlc.arg(student_id);

-- name: ResubmitAssignment :one
WITH archived AS (
    INSERT INTO assignment_submission_versions
        (submission_id, version, submission_text, file_path, file_name, submitted_at)
    SELECT s.id,
           COALESCE((SELECT max(v.version) + 1 FROM assignment_submission_versions v WHERE v.submission_id = s.id), 1),
           s.submission_text, s.file_path, s.file_name, s.submitted_at
    FROM assignment_submissions s
    WHERE s.id = sqlc.arg(id)
),
updated AS (
    UPDATE assignment_submissions
    SET submission_text = sqlc.arg(submission_text),
        file_path = sqlc.arg(file_path),
        file_name = sqlc.arg(file_name),
        submitted_at = now(),
        grade = NULL,
        feedback = NULL,
        graded_at = NULL,
        updated_at = now()
    WHERE id = sqlc.arg(id)
    RETURNING id, assignment_id, student_id, submission_text, file_path, file_name,
              submitted_at, grade, feedback, created_at, updated_at
)
SELECT * FROM updated;

-- name: ListSubmissionVersions :many
SELECT id, submission_id, version, submission_text, file_path, file_name, submitted_at, created_at
FROM assignment_submission_versions
WHERE submission_id = sqlc.arg(submission_id)
ORDER BY version DESC;

-- name: CreateAnnouncement :one
INSERT INTO announcements (course_id, title, content, status, created_by)
VALUES (sqlc.arg(course_id), sqlc.arg(title), sqlc.arg(content), sqlc.arg(status), sqlc.arg(created_by))
RETURNING id, course_id, title, content, status, created_by, created_at, updated_at;

-- name: GetAnnouncementByID :one
SELECT id, course_id, title, content, status, created_by, created_at, updated_at
FROM announcements
WHERE id = sqlc.arg(id);

-- name: ListInstructorAnnouncements :many
SELECT id, course_id, title, content, status, created_by, created_at, updated_at
FROM announcements
WHERE course_id = sqlc.arg(course_id)
ORDER BY created_at DESC;

-- name: ListPublishedAnnouncements :many
SELECT id, course_id, title, content, status, created_by, created_at, updated_at
FROM announcements
WHERE course_id = sqlc.arg(course_id) AND status = 'published'
ORDER BY created_at DESC;

-- name: UpdateAnnouncement :one
UPDATE announcements
SET title = sqlc.arg(title),
    content = sqlc.arg(content),
    status = sqlc.arg(status),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, course_id, title, content, status, created_by, created_at, updated_at;

-- name: DeleteAnnouncement :execrows
DELETE FROM announcements WHERE id = sqlc.arg(id);

-- name: GetCourseAnalytics :one
SELECT
    (SELECT count(*) FROM enrollments e WHERE e.course_id = sqlc.arg(target_course_id))::bigint AS enrollments,
    (SELECT count(*) FROM assignments a WHERE a.course_id = sqlc.arg(target_course_id))::bigint AS assignments,
    (SELECT count(*)
       FROM assignment_submissions s
       JOIN assignments a ON a.id = s.assignment_id
      WHERE a.course_id = sqlc.arg(target_course_id))::bigint AS submissions,
    COALESCE((SELECT avg(s.grade)
       FROM assignment_submissions s
       JOIN assignments a ON a.id = s.assignment_id
      WHERE a.course_id = sqlc.arg(target_course_id) AND s.grade IS NOT NULL), 0)::float8 AS average_grade,
    (SELECT count(*)
       FROM lessons l
      WHERE l.course_id = sqlc.arg(target_course_id))::bigint AS lessons,
    (SELECT count(*)
       FROM lesson_progress lp
       JOIN lessons l ON l.id = lp.lesson_id
      WHERE l.course_id = sqlc.arg(target_course_id) AND lp.completed = true)::bigint AS lesson_completions;

-- name: GetInstructorStats :one
SELECT
    (SELECT count(*) FROM courses c WHERE c.instructor_id = sqlc.arg(instructor_id))::bigint AS courses,
    (SELECT count(*) FROM courses c WHERE c.instructor_id = sqlc.arg(instructor_id) AND c.published = true)::bigint AS published_courses,
    (SELECT count(*)
       FROM enrollments e
       JOIN courses c ON c.id = e.course_id
      WHERE c.instructor_id = sqlc.arg(instructor_id))::bigint AS enrollments,
    (SELECT count(DISTINCT e.user_id)
       FROM enrollments e
       JOIN courses c ON c.id = e.course_id
      WHERE c.instructor_id = sqlc.arg(instructor_id))::bigint AS students,
    (SELECT count(*)
       FROM lessons l
       JOIN courses c ON c.id = l.course_id
      WHERE c.instructor_id = sqlc.arg(instructor_id))::bigint AS lessons,
    (SELECT count(*)
       FROM assignments a
       JOIN courses c ON c.id = a.course_id
      WHERE c.instructor_id = sqlc.arg(instructor_id))::bigint AS assignments,
    (SELECT count(*)
       FROM assignment_submissions s
       JOIN assignments a ON a.id = s.assignment_id
       JOIN courses c ON c.id = a.course_id
      WHERE c.instructor_id = sqlc.arg(instructor_id))::bigint AS submissions,
    (SELECT count(*)
       FROM assignment_submissions s
       JOIN assignments a ON a.id = s.assignment_id
       JOIN courses c ON c.id = a.course_id
      WHERE c.instructor_id = sqlc.arg(instructor_id) AND s.grade IS NULL)::bigint AS ungraded_submissions,
    (SELECT count(*)
       FROM announcements an
       JOIN courses c ON c.id = an.course_id
      WHERE c.instructor_id = sqlc.arg(instructor_id))::bigint AS announcements;

-- name: GetStudentStats :one
SELECT
    (SELECT count(*) FROM enrollments e WHERE e.user_id = sqlc.arg(student_id))::bigint AS enrolled_courses,
    (SELECT count(*)
       FROM lesson_progress lp
      WHERE lp.student_id = sqlc.arg(student_id) AND lp.completed = true)::bigint AS completed_lessons,
    (SELECT count(*)
       FROM assignment_submissions s
      WHERE s.student_id = sqlc.arg(student_id))::bigint AS submissions,
    (SELECT count(*)
       FROM assignment_submissions s
      WHERE s.student_id = sqlc.arg(student_id) AND s.grade IS NOT NULL)::bigint AS graded_submissions,
    COALESCE((SELECT avg(s.grade)
       FROM assignment_submissions s
      WHERE s.student_id = sqlc.arg(student_id) AND s.grade IS NOT NULL), 0)::float8 AS average_grade,
    (SELECT count(*)
       FROM assignments a
       JOIN enrollments e ON e.course_id = a.course_id AND e.user_id = sqlc.arg(student_id)
      WHERE a.status = 'published'
        AND NOT EXISTS (
            SELECT 1
              FROM assignment_submissions s
             WHERE s.assignment_id = a.id AND s.student_id = sqlc.arg(student_id)
        ))::bigint AS pending_assignments;

