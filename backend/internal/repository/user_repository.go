package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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
	List(ctx context.Context, search, role string, limit, offset int) ([]domain.User, error)
	Count(ctx context.Context, search, role string) (int, error)
	Update(ctx context.Context, user domain.User) (domain.User, error)
	GetAdminStats(ctx context.Context) (domain.AdminStats, error)
}

type userRepository struct {
	queries *db.Queries // sqlc generated queries
}

func NewUserRepository(queries *db.Queries) UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := r.queries.GetUserByID(ctx, pgUUID(id))
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(user)
}

func (r *userRepository) GetByIDWithPassword(ctx context.Context, id uuid.UUID) (domain.UserWithPassword, error) {
	user, err := r.queries.GetUserByID(ctx, pgUUID(id))
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
		ID:           pgUUID(id),
		PasswordHash: passwordHash,
	})
	return mapError(err)
}

func (r *userRepository) List(ctx context.Context, search, role string, limit, offset int) ([]domain.User, error) {
	rows, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		Search:      search,
		Role:        role,
		LimitCount:  int32(limit),
		OffsetCount: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		user, err := toDomainUser(row)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *userRepository) Count(ctx context.Context, search, role string) (int, error) {
	total, err := r.queries.CountUsers(ctx, db.CountUsersParams{Search: search, Role: role})
	return int(total), err
}

func (r *userRepository) Update(ctx context.Context, user domain.User) (domain.User, error) {
	row, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:       pgUUID(user.ID),
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	})
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(row)
}

func (r *userRepository) GetAdminStats(ctx context.Context) (domain.AdminStats, error) {
	row, err := r.queries.GetAdminStats(ctx)
	if err != nil {
		return domain.AdminStats{}, err
	}
	return domain.AdminStats{
		Users:       int(row.Users),
		Students:    int(row.Students),
		Instructors: int(row.Instructors),
		Courses:     int(row.Courses),
		Enrollments: int(row.Enrollments),
		Assignments: int(row.Assignments),
		Submissions: int(row.Submissions),
	}, nil
}

func toDomainUser(user db.User) (domain.User, error) {
	id, err := fromPgUUID(user.ID)
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

func pgUUID(id uuid.UUID) pgtype.UUID {

	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromPgUUID(id pgtype.UUID) (uuid.UUID, error) {
	if !id.Valid {
		return uuid.Nil, fmt.Errorf("invalid uuid")
	}
	return uuid.FromBytes(id.Bytes[:])
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
