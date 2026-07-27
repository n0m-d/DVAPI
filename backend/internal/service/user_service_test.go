package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/repository"
	"github.com/n0m-d/DVAPI/internal/service"
)

// mockUserRepository implements repository.UserRepository for unit tests.
// No database, no sqlc — only the behavior we stub in each test.
type mockUserRepository struct {
	createFn                 func(ctx context.Context, input domain.CreateUserInput) (domain.User, error)
	getByIDFn                func(ctx context.Context, id uuid.UUID) (domain.User, error)
	getByEmailFn             func(ctx context.Context, email string) (domain.User, error)
	getByEmailWithPasswordFn func(ctx context.Context, email string) (domain.UserWithPassword, error)
}

func (m *mockUserRepository) Create(ctx context.Context, input domain.CreateUserInput) (domain.User, error) {
	return m.createFn(ctx, input)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return domain.User{}, repository.ErrUserNotFound
}

func (m *mockUserRepository) GetByEmailWithPassword(ctx context.Context, email string) (domain.UserWithPassword, error) {
	if m.getByEmailWithPasswordFn != nil {
		return m.getByEmailWithPasswordFn(ctx, email)
	}
	return domain.UserWithPassword{}, repository.ErrUserNotFound
}

func (m *mockUserRepository) GetByIDWithPassword(ctx context.Context, id uuid.UUID) (domain.UserWithPassword, error) {
	return domain.UserWithPassword{}, repository.ErrUserNotFound
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return nil
}

func (m *mockUserRepository) List(context.Context, string, string, int, int) ([]domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) Count(context.Context, string, string) (int, error) {
	return 0, nil
}

func (m *mockUserRepository) Update(_ context.Context, user domain.User) (domain.User, error) {
	return user, nil
}

func (m *mockUserRepository) GetAdminStats(context.Context) (domain.AdminStats, error) {
	return domain.AdminStats{}, nil
}

func TestUpdateProfile_RejectsUpdatingAnotherUser(t *testing.T) {
	t.Parallel()

	svc := service.NewUserService(&mockUserRepository{}, nil)
	_, err := svc.UpdateProfile(context.Background(), uuid.New(), uuid.New(), domain.UpdateUserRequest{})

	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestGetByID_RejectsNilUUIDWithoutCallingRepository(t *testing.T) {
	t.Parallel()

	called := false
	repo := &mockUserRepository{
		getByIDFn: func(context.Context, uuid.UUID) (domain.User, error) {
			called = true
			return domain.User{}, nil
		},
	}

	svc := service.NewUserService(repo, nil)
	_, err := svc.GetByID(context.Background(), uuid.Nil)

	if !errors.Is(err, service.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
	if called {
		t.Error("repository GetByID was called, expected validation to fail first")
	}
}

func TestGetByID_MapsNotFound(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{
		getByIDFn: func(context.Context, uuid.UUID) (domain.User, error) {
			return domain.User{}, repository.ErrUserNotFound
		},
	}

	svc := service.NewUserService(repo, nil)
	_, err := svc.GetByID(context.Background(), uuid.New())

	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
