package service

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrInvalidInput   = errors.New("invalid input")
	ErrOTPInvalid     = errors.New("invalid or expired otp")
	ErrOTPNotVerified = errors.New("otp not verified")
)
