-- +goose Up
-- +goose StatementBegin
ALTER TABLE assignment_submissions ADD COLUMN file_name TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE assignment_submissions DROP COLUMN file_name;
-- +goose StatementEnd
