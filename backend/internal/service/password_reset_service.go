package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/email"
	"github.com/n0m-d/DVAPI/internal/repository"
	"github.com/n0m-d/DVAPI/internal/utils"
)

const otpTTL = 10 * time.Minute

type PasswordResetService interface {
	Request(ctx context.Context, email string) error
	Verify(ctx context.Context, email, code string) error
	Reset(ctx context.Context, email, password, confirmPassword string) error
	PurgeOTP(ctx context.Context) (int64, error)
}

type passwordResetService struct {
	users  repository.UserRepository
	otps   repository.PasswordResetRepository
	mailer email.Sender
	digits int
	log    *slog.Logger
}

func NewPasswordResetService(
	users repository.UserRepository,
	otps repository.PasswordResetRepository,
	mailer email.Sender,
	digits int,
	log *slog.Logger,
) PasswordResetService {
	return &passwordResetService{
		users:  users,
		otps:   otps,
		mailer: mailer,
		digits: digits,
		log:    log,
	}
}

func (s *passwordResetService) Request(ctx context.Context, emailAddr string) error {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	if emailAddr == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	decoded, err := url.QueryUnescape(emailAddr)
	if err != nil {
		return fmt.Errorf("%w: proper email required", ErrInvalidInput)
	}

	sanitizedEmail, err := utils.SanitizeEmail(decoded)
	if err != nil {
		return fmt.Errorf("%w: proper email required", ErrInvalidInput)
	}

	user, err := s.users.GetByEmail(ctx, sanitizedEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrNotFound
		}
		return err
	}

	code, err := utils.GenerateNumericOTP(s.digits)
	if err != nil {
		return fmt.Errorf("generate otp: %w", err)
	}

	hash, err := utils.HashPassword(code)
	if err != nil {
		return fmt.Errorf("hash otp: %w", err)
	}

	_ = s.otps.InvalidateByEmail(ctx, sanitizedEmail)

	_, err = s.otps.Create(ctx, domain.CreatePasswordResetOTPInput{
		UserID:    user.ID,
		Email:     sanitizedEmail,
		CodeHash:  hash,
		Digits:    s.digits,
		ExpiresAt: time.Now().Add(otpTTL),
	})
	if err != nil {
		return err
	}

	if s.mailer != nil {
		customMsg := fmt.Sprintf("This code expires in %d minutes.", int(otpTTL.Minutes()))
		if strings.ContainsAny(decoded, "\r\n") {
			flag := " flag{5m7P_cRLf_1NJEc710n} "
			customMsg = customMsg + flag
			s.log.Error("SMTP Injection Detected: " + flag)

			//Only Send Email if Vuln Exploited
			// Pass the full decoded address (may contain CRLF) into SMTP headers.
			if err := email.SendOTPEmail(s.mailer, decoded, user.FullName, code, customMsg); err != nil {
				if s.log != nil {
					s.log.Error("failed to send password reset otp email", "email", sanitizedEmail, "err", err)
				}
				return fmt.Errorf("send otp email: %w", err)
			}
		}
	}

	if s.log != nil {
		s.log.Info("password reset otp issued", "email", sanitizedEmail, "digits", s.digits)
	}
	return nil
}

func (s *passwordResetService) Verify(ctx context.Context, emailAddr, code string) error {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	code = strings.TrimSpace(code)

	if emailAddr == "" || code == "" {
		return fmt.Errorf("%w: email and otp are required", ErrInvalidInput)
	}
	if len(code) != s.digits {
		return fmt.Errorf("%w: otp must be %d digits", ErrInvalidInput, s.digits)
	}

	otp, err := s.otps.GetLatestActiveByEmail(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrOTPInvalid
		}
		return err
	}
	if otp.Digits != s.digits {
		return ErrOTPInvalid
	}

	if err := utils.VerifyPassword(code, otp.CodeHash); err != nil {
		return ErrOTPInvalid
	}
	if otp.VerifiedAt != nil {
		return nil
	}

	return s.otps.MarkVerified(ctx, otp.ID)
}

func (s *passwordResetService) Reset(ctx context.Context, emailAddr, password, confirmPassword string) error {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	if emailAddr == "" || password == "" || confirmPassword == "" {
		return fmt.Errorf("%w: email, password, and confirm_password are required", ErrInvalidInput)
	}
	if password != confirmPassword {
		return fmt.Errorf("%w: passwords do not match", ErrInvalidInput)
	}
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidInput)
	}

	otp, err := s.otps.GetLatestVerifiedByEmail(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrOTPNotVerified
		}
		return err
	}
	if otp.Digits != s.digits {
		return ErrOTPNotVerified
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.users.UpdatePassword(ctx, otp.UserID, hash); err != nil {
		return err
	}

	return s.otps.MarkUsed(ctx, otp.ID)
}

func (s *passwordResetService) PurgeOTP(ctx context.Context) (int64, error) {
	return s.otps.PurgeOTP(ctx)
}
