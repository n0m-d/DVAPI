-- +goose Up
CREATE TYPE announcement_status AS ENUM ('draft', 'published');

ALTER TABLE assignment_submissions
    ALTER COLUMN grade DROP DEFAULT,
    ADD COLUMN graded_at TIMESTAMPTZ;

UPDATE assignment_submissions
SET grade = NULL
WHERE grade = 0 AND feedback IS NULL;

CREATE TABLE lesson_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    completed BOOLEAN NOT NULL DEFAULT false,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (student_id, lesson_id)
);

CREATE TABLE assignment_submission_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES assignment_submissions(id) ON DELETE CASCADE,
    version INT NOT NULL,
    submission_text TEXT,
    file_path TEXT,
    file_name TEXT,
    submitted_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (submission_id, version)
);

CREATE TABLE announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    status announcement_status NOT NULL DEFAULT 'draft',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lesson_progress_student ON lesson_progress(student_id);
CREATE INDEX idx_submission_versions_submission ON assignment_submission_versions(submission_id);
CREATE INDEX idx_announcements_course_status ON announcements(course_id, status);

-- +goose Down
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS assignment_submission_versions;
DROP TABLE IF EXISTS lesson_progress;
ALTER TABLE assignment_submissions
    DROP COLUMN IF EXISTS graded_at,
    ALTER COLUMN grade SET DEFAULT 0;
DROP TYPE IF EXISTS announcement_status;
