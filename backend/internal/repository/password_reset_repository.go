package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, input domain.CreatePasswordResetOTPInput) (domain.PasswordResetOTP, error)
	GetLatestActiveByEmail(ctx context.Context, email string) (domain.PasswordResetOTP, error)
	GetLatestVerifiedByEmail(ctx context.Context, email string) (domain.PasswordResetOTP, error)
	MarkVerified(ctx context.Context, id uuid.UUID) error
	MarkUsed(ctx context.Context, id uuid.UUID) error
	InvalidateByEmail(ctx context.Context, email string) error
	PurgeOTP(ctx context.Context) (int64, error)
}

type passwordResetRepository struct {
	queries *db.Queries
}

func NewPasswordResetRepository(queries *db.Queries) PasswordResetRepository {
	return &passwordResetRepository{queries: queries}
}

func (r *passwordResetRepository) Create(ctx context.Context, input domain.CreatePasswordResetOTPInput) (domain.PasswordResetOTP, error) {
	row, err := r.queries.CreatePasswordResetOTP(ctx, db.CreatePasswordResetOTPParams{
		UserID:    pgUUID(input.UserID),
		Email:     input.Email,
		CodeHash:  input.CodeHash,
		Digits:    int32(input.Digits),
		ExpiresAt: pgtype.Timestamptz{Time: input.ExpiresAt, Valid: true},
	})
	if err != nil {
		return domain.PasswordResetOTP{}, mapError(err)
	}
	return toDomainOTP(row)
}

func (r *passwordResetRepository) GetLatestActiveByEmail(ctx context.Context, email string) (domain.PasswordResetOTP, error) {
	row, err := r.queries.GetLatestActiveOTPByEmail(ctx, email)
	if err != nil {
		return domain.PasswordResetOTP{}, mapError(err)
	}
	return toDomainOTP(row)
}

func (r *passwordResetRepository) GetLatestVerifiedByEmail(ctx context.Context, email string) (domain.PasswordResetOTP, error) {
	row, err := r.queries.GetLatestVerifiedOTPByEmail(ctx, email)
	if err != nil {
		return domain.PasswordResetOTP{}, mapError(err)
	}
	return toDomainOTP(row)
}

func (r *passwordResetRepository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	return mapError(r.queries.MarkOTPVerified(ctx, pgUUID(id)))
}

func (r *passwordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return mapError(r.queries.MarkOTPUsed(ctx, pgUUID(id)))
}

func (r *passwordResetRepository) InvalidateByEmail(ctx context.Context, email string) error {
	return mapError(r.queries.InvalidateOTPsByEmail(ctx, email))
}

func toDomainOTP(row db.PasswordResetOtp) (domain.PasswordResetOTP, error) {
	id, err := fromPgUUID(row.ID)
	if err != nil {
		return domain.PasswordResetOTP{}, fmt.Errorf("parse otp id: %w", err)
	}
	userID, err := fromPgUUID(row.UserID)
	if err != nil {
		return domain.PasswordResetOTP{}, fmt.Errorf("parse user id: %w", err)
	}

	otp := domain.PasswordResetOTP{
		ID:        id,
		UserID:    userID,
		Email:     row.Email,
		CodeHash:  row.CodeHash,
		Digits:    int(row.Digits),
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
	}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		otp.VerifiedAt = &t
	}
	if row.UsedAt.Valid {
		t := row.UsedAt.Time
		otp.UsedAt = &t
	}
	return otp, nil
}

func (r *passwordResetRepository) PurgeOTP(ctx context.Context) (int64, error) {
	return r.queries.PurgeUsedOTPs(ctx)
}
