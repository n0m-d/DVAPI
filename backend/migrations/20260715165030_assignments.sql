-- +goose Up
-- +goose StatementBegin


CREATE TYPE status AS ENUM ('draft', 'published', 'closed');

CREATE TABLE assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    due_date TIMESTAMPTZ NOT NULL,
    status status NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE assignment_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES assignments (id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    submission_text TEXT,
    file_path TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    grade INT DEFAULT 0,
    feedback TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_assignments_course ON assignments (course_id);
CREATE INDEX idx_assignments_created_by ON assignments (created_by);
CREATE INDEX idx_assignment_submissions_assignment ON assignment_submissions (assignment_id);
CREATE INDEX idx_assignment_submissions_student ON assignment_submissions (student_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS assignments;
DROP TABLE IF EXISTS assignment_submissions;

-- +goose StatementEnd
