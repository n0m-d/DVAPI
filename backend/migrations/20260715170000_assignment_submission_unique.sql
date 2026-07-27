-- +goose Up
-- +goose StatementBegin
ALTER TABLE assignment_submissions
    ADD CONSTRAINT assignment_submissions_assignment_student_key UNIQUE (assignment_id, student_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE assignment_submissions
    DROP CONSTRAINT IF EXISTS assignment_submissions_assignment_student_key;
-- +goose StatementEnd
