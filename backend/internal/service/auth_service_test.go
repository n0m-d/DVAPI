package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/config"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/repository"
	"github.com/n0m-d/DVAPI/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type mockBlacklist struct {
	blacklistFn     func(ctx context.Context, jti string, ttl time.Duration) error
	isBlacklistedFn func(ctx context.Context, jti string) (bool, error)
}

func (m *mockBlacklist) Blacklist(ctx context.Context, jti string, ttl time.Duration) error {
	return m.blacklistFn(ctx, jti, ttl)
}

func (m *mockBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return m.isBlacklistedFn(ctx, jti)
}

func TestRegister_NormalizesEmailAndDefaultsToStudentRole(t *testing.T) {
	t.Parallel()

	var captured domain.CreateUserInput
	repo := &mockUserRepository{
		createFn: func(_ context.Context, input domain.CreateUserInput) (domain.User, error) {
			captured = input
			return domain.User{
				ID:        uuid.New(),
				Email:     input.Email,
				FullName:  input.FullName,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	svc := service.NewAuthService(repo, &config.Config{JWT_SECRET: "test-secret"}, &mockBlacklist{})
	user, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "  Alice@Example.COM  ",
		Password: "password1",
		FullName: " Alice ",
	})

	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if captured.Email != "alice@example.com" {
		t.Errorf("email = %q, want alice@example.com", captured.Email)
	}
	if captured.FullName != "Alice" {
		t.Errorf("full_name = %q, want Alice", captured.FullName)
	}
	if captured.PasswordHash == "" {
		t.Error("expected password hash to be set")
	}
	if user.Email != "alice@example.com" {
		t.Errorf("returned email = %q", user.Email)
	}
}

func TestRegister_MapsDuplicateEmail(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{
		createFn: func(context.Context, domain.CreateUserInput) (domain.User, error) {
			return domain.User{}, repository.ErrUserAlreadyExists
		},
	}

	svc := service.NewAuthService(repo, &config.Config{JWT_SECRET: "test-secret"}, &mockBlacklist{})
	_, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "taken@example.com",
		Password: "password1",
		FullName: "Taken",
	})

	if !errors.Is(err, service.ErrAlreadyExists) {
		t.Fatalf("error = %v, want ErrAlreadyExists", err)
	}
}

func TestLogin_ReturnsToken(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &mockUserRepository{
		getByEmailWithPasswordFn: func(_ context.Context, email string) (domain.UserWithPassword, error) {
			return domain.UserWithPassword{
				User: domain.User{
					ID:       userID,
					Email:    email,
					FullName: "Alice",
					Role:     "student",
				},
				PasswordHash: string(hash),
			}, nil
		},
	}

	bl := &mockBlacklist{}
	svc := service.NewAuthService(repo, &config.Config{JWT_SECRET: "test-secret"}, bl)
	resp, err := svc.Login(context.Background(), "alice@example.com", "password1")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if resp.User.ID != userID {
		t.Errorf("user id = %v, want %v", resp.User.ID, userID)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &mockUserRepository{
		getByEmailWithPasswordFn: func(_ context.Context, email string) (domain.UserWithPassword, error) {
			return domain.UserWithPassword{
				User:         domain.User{ID: uuid.New(), Email: email},
				PasswordHash: string(hash),
			}, nil
		},
	}

	svc := service.NewAuthService(repo, &config.Config{JWT_SECRET: "test-secret"}, &mockBlacklist{})
	_, err = svc.Login(context.Background(), "alice@example.com", "wrong-password")
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Fatalf("error = %v, want ErrInvalidCredentials", err)
	}
}
