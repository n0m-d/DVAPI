package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/repository"
	"github.com/n0m-d/DVAPI/internal/utils"
)

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	UpdateProfile(ctx context.Context, id, actorID uuid.UUID, input domain.UpdateUserRequest) (domain.User, error)
	UpdateProfileV1(ctx context.Context, id, actorID uuid.UUID, input domain.User) (domain.User, error)
	AdminCreate(ctx context.Context, input domain.AdminCreateUserRequest) (domain.User, error)
	AdminList(ctx context.Context, search, role string, page, pageSize int) (domain.AdminUserListResponse, error)
	AdminUpdate(ctx context.Context, id, actorID uuid.UUID, input domain.AdminUpdateUserRequest) (domain.User, error)
	AdminStats(ctx context.Context) (domain.AdminStats, error)
}

type userService struct {
	userRepo repository.UserRepository
	cache    cache.UserCache
}

func NewUserService(userRepo repository.UserRepository, userCache cache.UserCache) UserService {
	return &userService{userRepo: userRepo, cache: userCache}
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	if id == uuid.Nil {
		return domain.User{}, fmt.Errorf("%w: invalid user id", ErrInvalidInput)
	}

	if s.cache != nil {
		if user, ok, err := s.cache.GetUser(ctx, id); err == nil && ok {
			return user, nil
		}
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, mapError(err)
	}

	if s.cache != nil {
		_ = s.cache.SetUser(ctx, user)
	}

	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, id, actorID uuid.UUID, input domain.UpdateUserRequest) (domain.User, error) {
	if id == uuid.Nil || id != actorID {
		return domain.User{}, ErrForbidden
	}
	if input.Email == nil && input.FullName == nil {
		return domain.User{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, mapError(err)
	}
	if input.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*input.Email))
		if email == "" {
			return domain.User{}, fmt.Errorf("%w: email cannot be empty", ErrInvalidInput)
		}
		user.Email = email
	}
	if input.FullName != nil {
		fullName := strings.TrimSpace(*input.FullName)
		if fullName == "" {
			return domain.User{}, fmt.Errorf("%w: full_name cannot be empty", ErrInvalidInput)
		}
		user.FullName = fullName
	}

	updated, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return domain.User{}, mapError(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteUser(ctx, id)
	}
	return updated, nil
}

func (s *userService) UpdateProfileV1(ctx context.Context, id, actorID uuid.UUID, input domain.User) (domain.User, error) {
	if id == uuid.Nil || id != actorID {
		return domain.User{}, ErrForbidden
	}

	if strings.TrimSpace(input.Email) == "" && strings.TrimSpace(input.FullName) == "" && strings.TrimSpace(input.Role) == "" {
		return domain.User{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, mapError(err)
	}

	if strings.TrimSpace(input.Email) != "" {
		email := strings.TrimSpace(strings.ToLower(input.Email))
		if email == "" {
			return domain.User{}, fmt.Errorf("%w: email cannot be empty", ErrInvalidInput)
		}
		user.Email = email
	}
	if strings.TrimSpace(input.FullName) != "" {
		fullName := strings.TrimSpace(input.FullName)
		if fullName == "" {
			return domain.User{}, fmt.Errorf("%w: full_name cannot be empty", ErrInvalidInput)
		}
		user.FullName = fullName
	}
	if strings.TrimSpace(input.Role) != "" {
		role := strings.TrimSpace(input.Role)
		if !validUserRole(role) {
			return domain.User{}, fmt.Errorf("%w: invalid role", ErrInvalidInput)
		}
		user.Role = role
	}

	updated, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return domain.User{}, mapError(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteUser(ctx, id)
	}
	return updated, nil
}

func (s *userService) AdminCreate(ctx context.Context, input domain.AdminCreateUserRequest) (domain.User, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	fullName := strings.TrimSpace(input.FullName)
	role := strings.TrimSpace(input.Role)
	if email == "" || fullName == "" || input.Password == "" || !validUserRole(role) {
		return domain.User{}, fmt.Errorf("%w: valid email, password, full_name, and role are required", ErrInvalidInput)
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.userRepo.Create(ctx, domain.CreateUserInput{
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         role,
	})
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return user, nil
}

func (s *userService) AdminList(ctx context.Context, search, role string, page, pageSize int) (domain.AdminUserListResponse, error) {
	search = strings.TrimSpace(search)
	role = strings.TrimSpace(role)
	if role != "" && !validUserRole(role) {
		return domain.AdminUserListResponse{}, fmt.Errorf("%w: invalid role", ErrInvalidInput)
	}

	page, pageSize = normalizePagination(page, pageSize)
	users, err := s.userRepo.List(ctx, search, role, pageSize, (page-1)*pageSize)
	if err != nil {
		return domain.AdminUserListResponse{}, fmt.Errorf("list users: %w", err)
	}
	total, err := s.userRepo.Count(ctx, search, role)
	if err != nil {
		return domain.AdminUserListResponse{}, fmt.Errorf("count users: %w", err)
	}
	pagination := domain.NewPagination(total, page, pageSize)
	return domain.AdminUserListResponse{
		Status: "success",
		Data: &domain.AdminUserListData{
			Users:      users,
			Pagination: &pagination,
		},
	}, nil
}

func (s *userService) AdminUpdate(ctx context.Context, id, actorID uuid.UUID, input domain.AdminUpdateUserRequest) (domain.User, error) {
	if id == uuid.Nil {
		return domain.User{}, fmt.Errorf("%w: invalid user id", ErrInvalidInput)
	}
	if input.Email == nil && input.FullName == nil && input.Role == nil {
		return domain.User{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, mapError(err)
	}

	if input.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*input.Email))
		if email == "" {
			return domain.User{}, fmt.Errorf("%w: email cannot be empty", ErrInvalidInput)
		}
		user.Email = email
	}
	if input.FullName != nil {
		fullName := strings.TrimSpace(*input.FullName)
		if fullName == "" {
			return domain.User{}, fmt.Errorf("%w: full_name cannot be empty", ErrInvalidInput)
		}
		user.FullName = fullName
	}
	if input.Role != nil {
		role := strings.TrimSpace(*input.Role)
		if !validUserRole(role) {
			return domain.User{}, fmt.Errorf("%w: invalid role", ErrInvalidInput)
		}
		if id == actorID && role != "admin" {
			return domain.User{}, fmt.Errorf("%w: you cannot remove your own admin role", ErrInvalidInput)
		}
		user.Role = role
	}

	updated, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return domain.User{}, mapError(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteUser(ctx, id)
	}
	return updated, nil
}

func (s *userService) AdminStats(ctx context.Context) (domain.AdminStats, error) {
	return s.userRepo.GetAdminStats(ctx)
}

func validUserRole(role string) bool {
	return role == "student" || role == "instructor" || role == "admin"
}

func mapError(err error) error {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		return ErrNotFound
	case errors.Is(err, repository.ErrUserAlreadyExists):
		return ErrAlreadyExists
	default:
		return err
	}
}
