package domain

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetOTP struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Email      string
	CodeHash   string
	Digits     int
	ExpiresAt  time.Time
	VerifiedAt *time.Time
	UsedAt     *time.Time
	CreatedAt  time.Time
}

type CreatePasswordResetOTPInput struct {
	UserID    uuid.UUID
	Email     string
	CodeHash  string
	Digits    int
	ExpiresAt time.Time
}
