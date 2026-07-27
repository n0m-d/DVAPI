package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/config"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/repository"
	"github.com/n0m-d/DVAPI/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
)

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.LoginResponse, error)
	Logout(ctx context.Context, token string) error
	UpdatePassword(ctx context.Context, userId uuid.UUID, currentPassword, newPassword string) error
}

type RegisterInput struct {
	Email    string
	Password string
	FullName string
}

type authService struct {
	userRepo       repository.UserRepository
	cfg            *config.Config
	tokenBlacklist cache.TokenBlacklist
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config, tokenBlacklist cache.TokenBlacklist) AuthService {
	return &authService{
		userRepo:       userRepo,
		cfg:            cfg,
		tokenBlacklist: tokenBlacklist,
	}
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (domain.User, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	fullName := strings.TrimSpace(input.FullName)
	role := "student"

	if email == "" || input.Password == "" || fullName == "" {
		return domain.User{}, fmt.Errorf("%w: email, password, and full_name are required", ErrInvalidInput)
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, domain.CreateUserInput{
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         role,
	})
	if err != nil {
		return domain.User{}, mapError(err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.LoginResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.userRepo.GetByEmailWithPassword(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := utils.VerifyPassword(password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	pair, err := utils.GenerateToken(&user.User, s.cfg.JWT_SECRET)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &domain.LoginResponse{
		AccessToken: pair.Token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(utils.TokenTTL.Seconds()),
		User:        user.User,
	}, nil
}

func (s *authService) UpdatePassword(ctx context.Context, userId uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByIDWithPassword(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrNotFound
		}
		return err
	}

	if err := utils.VerifyPassword(currentPassword, user.PasswordHash); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userId, hash); err != nil {
		return err
	}
	return nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	claims, err := utils.ParseToken(token, s.cfg.JWT_SECRET)
	if err != nil {
		return ErrUnauthorized
	}

	ttl := utils.TokenTTL
	if claims.ExpiresAt != nil {
		if remaining := time.Until(claims.ExpiresAt.Time); remaining > 0 {
			ttl = remaining
		}
	}

	if err := s.tokenBlacklist.Blacklist(ctx, claims.ID, ttl); err != nil {
		return fmt.Errorf("blacklist token: %w", err)
	}

	return nil
}
