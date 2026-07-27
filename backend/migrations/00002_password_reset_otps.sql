-- +goose Up
CREATE TABLE password_reset_otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    digits INT NOT NULL CHECK (digits IN (4, 6)),
    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_password_reset_otps_email ON password_reset_otps (email);
CREATE INDEX idx_password_reset_otps_user ON password_reset_otps (user_id);

-- +goose Down
DROP TABLE IF EXISTS password_reset_otps;
