-- name: CreatePasswordResetOTP :one
INSERT INTO password_reset_otps (user_id, email, code_hash, digits, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, email, code_hash, digits, expires_at, verified_at, used_at, created_at;

-- name: GetLatestActiveOTPByEmail :one
SELECT id, user_id, email, code_hash, digits, expires_at, verified_at, used_at, created_at
FROM password_reset_otps
WHERE email = $1
  AND used_at IS NULL
  AND expires_at > now()
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLatestVerifiedOTPByEmail :one
SELECT id, user_id, email, code_hash, digits, expires_at, verified_at, used_at, created_at
FROM password_reset_otps
WHERE email = $1
  AND verified_at IS NOT NULL
  AND used_at IS NULL
  AND expires_at > now()
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkOTPVerified :exec
UPDATE password_reset_otps
SET verified_at = now()
WHERE id = $1
  AND used_at IS NULL
  AND verified_at IS NULL;

-- name: MarkOTPUsed :exec
UPDATE password_reset_otps
SET used_at = now()
WHERE id = $1;

-- name: InvalidateOTPsByEmail :exec
UPDATE password_reset_otps
SET used_at = now()
WHERE email = $1
  AND used_at IS NULL;


-- name: PurgeUsedOTPs :execrows
DELETE FROM password_reset_otps WHERE used_at IS NOT NULL OR expires_at <= now();