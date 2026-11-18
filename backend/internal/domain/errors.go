package domain

import "errors"

// Sentinel errors shared by the service, query, and handler layers.
var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrForbidden          = errors.New("forbidden")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid email or password")

	ErrOTPInvalid     = errors.New("invalid or expired otp")
	ErrOTPNotVerified = errors.New("otp not verified")

	ErrNotEnrolled      = errors.New("not enrolled in course")
	ErrAssignmentClosed = errors.New("assignment is not open for submissions")
	ErrPastDue          = errors.New("assignment due date has passed")
)
