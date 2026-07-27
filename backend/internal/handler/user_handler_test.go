package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/handler"
	"github.com/n0m-d/DVAPI/internal/service"
)

type mockUserService struct {
	// getByIDFn is a function field that allows test cases to inject custom behavior
	// for the GetByID method. This enables testing different scenarios without
	// hitting the real database. If not set, it defaults to nil.
	getByIDFn func(ctx context.Context, id uuid.UUID) (domain.User, error)
}

// GetByID implements UserService.GetByID by delegating to the injected function.
// This allows each test to provide its own behavior for GetByID.
func (m *mockUserService) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return m.getByIDFn(ctx, id)
}

// UpdateProfile is a stub implementation that satisfies the UserService interface.
// It returns default values since these tests don't exercise this method.
func (m *mockUserService) UpdateProfile(context.Context, uuid.UUID, uuid.UUID, domain.UpdateUserRequest) (domain.User, error) {
	return domain.User{}, nil
}

// UpdateProfileV1 is a stub implementation that satisfies the UserService interface.
// It returns default values since these tests don't exercise this method.
func (m *mockUserService) UpdateProfileV1(context.Context, uuid.UUID, uuid.UUID, domain.User) (domain.User, error) {
	return domain.User{}, nil
}

// AdminCreate is a stub implementation that satisfies the UserService interface.
func (m *mockUserService) AdminCreate(context.Context, domain.AdminCreateUserRequest) (domain.User, error) {
	return domain.User{}, nil
}

// AdminList is a stub implementation that satisfies the UserService interface.
func (m *mockUserService) AdminList(context.Context, string, string, int, int) (domain.AdminUserListResponse, error) {
	return domain.AdminUserListResponse{}, nil
}

// AdminUpdate is a stub implementation that satisfies the UserService interface.
func (m *mockUserService) AdminUpdate(context.Context, uuid.UUID, uuid.UUID, domain.AdminUpdateUserRequest) (domain.User, error) {
	return domain.User{}, nil
}

// AdminStats is a stub implementation that satisfies the UserService interface.
func (m *mockUserService) AdminStats(context.Context) (domain.AdminStats, error) {
	return domain.AdminStats{}, nil
}

// setupUserRouter creates a minimal Gin router configured for testing the UserHandler.
// It injects the provided service (typically a mock) and sets up the GET /users/:id route.
// This allows tests to send HTTP requests to the handler without starting a real server.
func setupUserRouter(svc service.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)        // Disables debug output during tests
	r := gin.New()                   // Create a fresh router
	h := handler.NewUserHandler(svc) // Inject the mock or real service
	r.GET("/users/:id", h.GetByID)   // Register the route being tested
	return r
}

// TestGetByID_Returns404 verifies that when the service returns ErrNotFound (user doesn't exist),
// the handler correctly translates it to an HTTP 404 status code.
// This tests the handler's error-to-status-code mapping logic.
func TestGetByID_Returns404(t *testing.T) {
	t.Parallel()

	// Arrange: Set up mock to return "user not found" error
	svc := &mockUserService{
		getByIDFn: func(context.Context, uuid.UUID) (domain.User, error) {
			return domain.User{}, service.ErrNotFound
		},
	}

	// Act: Make HTTP GET request to retrieve a user
	id := uuid.New()
	router := setupUserRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/users/"+id.String(), nil)
	rec := httptest.NewRecorder() // Records the response for inspection
	router.ServeHTTP(rec, req)    // Process the request through the router

	// Assert: Verify the response status is 404 Not Found
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestGetByID_Returns400OnInvalidUUID verifies that when a malformed UUID is provided in the URL,
// the handler rejects the request with HTTP 400 before calling the service layer.
// The 'called' flag ensures the service is never invoked for invalid input, preventing unnecessary database hits.
func TestGetByID_Returns400OnInvalidUUID(t *testing.T) {
	t.Parallel()

	// Arrange: Track whether the service method is called (should NOT be called)
	called := false
	svc := &mockUserService{
		getByIDFn: func(context.Context, uuid.UUID) (domain.User, error) {
			called = true
			return domain.User{}, nil
		},
	}

	// Act: Make HTTP GET request with invalid UUID format
	router := setupUserRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert: Verify 400 Bad Request and that service was never called
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if called {
		t.Error("service GetByID was called, expected UUID parse to fail first")
	}
}

// TestGetByID_Returns500OnUnexpectedError verifies that when the service returns an unmapped error
// (e.g., database connection failure), the handler returns HTTP 500 Internal Server Error.
// This tests the catch-all error handling for errors not explicitly mapped in respondServiceError.
func TestGetByID_Returns500OnUnexpectedError(t *testing.T) {
	t.Parallel()

	// Arrange: Mock service to return an unexpected error
	svc := &mockUserService{
		getByIDFn: func(context.Context, uuid.UUID) (domain.User, error) {
			return domain.User{}, errors.New("database down")
		},
	}

	// Act: Make HTTP GET request with valid UUID (service will fail)
	id := uuid.New()
	router := setupUserRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/users/"+id.String(), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert: Verify 500 Internal Server Error is returned
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
