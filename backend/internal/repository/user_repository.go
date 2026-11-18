package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetByIDWithPassword(ctx context.Context, id uuid.UUID) (domain.UserWithPassword, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByEmailWithPassword(ctx context.Context, email string) (domain.UserWithPassword, error)
	Create(ctx context.Context, input domain.CreateUserInput) (domain.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	Update(ctx context.Context, user domain.User) (domain.User, error)
}

type userRepository struct {
	queries *db.Queries // sqlc generated queries
}

func NewUserRepository(queries *db.Queries) UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := r.queries.GetUserByID(ctx, db.PgUUID(id))
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(user)
}

func (r *userRepository) GetByIDWithPassword(ctx context.Context, id uuid.UUID) (domain.UserWithPassword, error) {
	user, err := r.queries.GetUserByID(ctx, db.PgUUID(id))
	if err != nil {
		return domain.UserWithPassword{}, mapError(err)
	}
	return toDomainUserWithPassword(user)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(user)
}

func (r *userRepository) GetByEmailWithPassword(ctx context.Context, email string) (domain.UserWithPassword, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.UserWithPassword{}, mapError(err)
	}

	return toDomainUserWithPassword(user)
}

func (r *userRepository) Create(ctx context.Context, input domain.CreateUserInput) (domain.User, error) {
	user, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
		FullName:     input.FullName,
		Role:         input.Role,
	})
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(user)
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	err := r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           db.PgUUID(id),
		PasswordHash: passwordHash,
	})
	return mapError(err)
}

func (r *userRepository) Update(ctx context.Context, user domain.User) (domain.User, error) {
	row, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:       db.PgUUID(user.ID),
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	})
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(row)
}

func toDomainUser(user db.User) (domain.User, error) {
	id, err := db.FromPgUUID(user.ID)
	if err != nil {
		return domain.User{}, fmt.Errorf("parse user id: %w", err)
	}

	return domain.User{
		ID:        id,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func toDomainUserWithPassword(dbUser db.User) (domain.UserWithPassword, error) {
	domainUser, err := toDomainUser(dbUser)
	if err != nil {
		return domain.UserWithPassword{}, fmt.Errorf("convert user to domain: %w", err)
	}
	return domain.UserWithPassword{
		User:         domainUser,
		PasswordHash: dbUser.PasswordHash,
	}, nil
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrUserAlreadyExists
	}

	return err
}
