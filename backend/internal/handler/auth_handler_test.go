package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/handler"
	"github.com/n0m-d/DVAPI/internal/service"
)

type mockAuthService struct {
	registerFn func(ctx context.Context, input service.RegisterInput) (domain.User, error)
	loginFn    func(ctx context.Context, email, password string) (*domain.LoginResponse, error)
	logoutFn   func(ctx context.Context, token string) error
}

func (m *mockAuthService) Register(ctx context.Context, input service.RegisterInput) (domain.User, error) {
	return m.registerFn(ctx, input)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*domain.LoginResponse, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, email, password)
	}
	return nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, token)
	}
	return nil
}

func (m *mockAuthService) UpdatePassword(ctx context.Context, userId uuid.UUID, currentPassword, newPassword string) error {
	return nil
}

func setupAuthRouter(svc service.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewAuthHandler(svc)
	r.POST("/auth/register", h.Register)
	return r
}

func TestRegister_Returns201(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	now := time.Now()
	svc := &mockAuthService{
		registerFn: func(_ context.Context, input service.RegisterInput) (domain.User, error) {
			if input.Email != "alice@example.com" {
				t.Errorf("service received email %q", input.Email)
			}
			return domain.User{
				ID:        userID,
				Email:     input.Email,
				FullName:  input.FullName,
				Role:      "student",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	router := setupAuthRouter(svc)
	body := `{"email":"alice@example.com","password":"password1","confirm_password":"password1","full_name":"Alice"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("status = %v, want success", resp["status"])
	}
}

func TestRegister_Returns400OnInvalidJSON(t *testing.T) {
	t.Parallel()

	called := false
	svc := &mockAuthService{
		registerFn: func(context.Context, service.RegisterInput) (domain.User, error) {
			called = true
			return domain.User{}, nil
		},
	}

	router := setupAuthRouter(svc)
	body := `{"email":"not-an-email","password":"short","full_name":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if called {
		t.Error("service Register was called, expected Gin validation to fail first")
	}
}

func TestRegister_Returns409OnDuplicateEmail(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		registerFn: func(context.Context, service.RegisterInput) (domain.User, error) {
			return domain.User{}, service.ErrAlreadyExists
		},
	}

	router := setupAuthRouter(svc)
	body := `{"email":"taken@example.com","password":"password1","confirm_password":"password1","full_name":"Taken"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}
