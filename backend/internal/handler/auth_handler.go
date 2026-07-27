package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/middleware"
	"github.com/n0m-d/DVAPI/internal/service"
	"github.com/n0m-d/DVAPI/internal/utils"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"test@example.com"`
	Password string `json:"password" binding:"required" example:"password"`
}

type registerRequest struct {
	Email           string `json:"email" binding:"required,email" example:"newuser@example.com"`
	Password        string `json:"password" binding:"required,min=8" example:"password123"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password" example:"password123"`
	FullName        string `json:"full_name" binding:"required" example:"Jane Doe"`
}

type passwordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword"`
	Password        string `json:"password" binding:"required,min=8" example:"newpassword123"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password" example:"newpassword123"`
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password and receive a JWT access token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  loginRequest  true  "Login credentials"
// @Success      200
// @Failure      400  "Validation error"
// @Failure      401  "Invalid email or password"
// @Failure      500  "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.MapError(err), "status": "error"})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "status": "error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdatePassword godoc
// @Summary      Update password
// @Description  Change the authenticated user's password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  passwordChangeRequest  true  "Password change payload"
// @Success      200   "Password updated successfully"
// @Failure      400   "Validation error"
// @Failure      401   "Unauthorized"
// @Failure      500   "Internal server error"
// @Router       /auth/update-password [post]
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	var req passwordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.MapError(err), "status": "error"})
		return
	}

	if req.Password == req.CurrentPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password and current password cannot be the same", "status": "error"})
		return
	}

	raw, ok := c.Get(middleware.ClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return
	}

	claims, ok := raw.(*domain.Claims)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return
	}

	if err := h.authService.UpdatePassword(c.Request.Context(), claims.UserID, req.CurrentPassword, req.Password); err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully", "status": "success"})
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with email, password, and full name.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  registerRequest  true  "Register request"
// @Success      201   "User registered successfully"
// @Failure      400   "Validation error"
// @Failure      409   "Email already exists"
// @Failure      500   "Internal server error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.MapError(err), "status": "error"})
		return
	}

	_, err := h.authService.Register(c.Request.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "status": "success"})
}

// Logout godoc
// @Summary      Logout
// @Description  Logout
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200   "Logged out successfully"
// @Failure      401   "Missing, invalid, or expired token"
// @Failure      500   "Internal server error"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token", "status": "error"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "status": "error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully", "status": "success"})
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
