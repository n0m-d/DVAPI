package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/middleware"
	"github.com/n0m-d/DVAPI/internal/service"
)

type UserHandler struct {
	users service.UserService
}

func NewUserHandler(users service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetCurrentUser godoc
// @Summary      Get current user
// @Description  Get the authenticated user's profile.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200
// @Failure      401  "Unauthorized"
// @Failure      404  "User not found"
// @Failure      500  "Internal server error"
// @Router       /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
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

	user, err := h.users.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User retrieved successfully", "status": "success", "data": user})
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update the profile for a user (own profile only).
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                   true  "User ID"
// @Param        body  body  domain.UpdateUserRequest  true  "Profile fields to update"
// @Success      200
// @Failure      400  "Invalid user id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "You can only update your own profile"
// @Failure      404  "User not found"
// @Failure      500  "Internal server error"
// @Router       /users/{id} [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id", "status": "error"})
		return
	}

	var input domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	user, err := h.users.UpdateProfile(c.Request.Context(), id, claims.UserID, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile updated successfully",
		"data":    user,
	})
}

func (h *UserHandler) UpdateProfilev1(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id", "status": "error"})
		return
	}

	var input domain.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	user, err := h.users.UpdateProfileV1(c.Request.Context(), id, claims.UserID, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile updated successfully",
		"data":    user,
	})
}

// AdminCreate godoc
// @Summary      Create user
// @Description  Create a new user account (admin only).
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  domain.AdminCreateUserRequest  true  "New user payload"
// @Success      201
// @Failure      400  "Validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      409  "Email already registered"
// @Failure      500  "Internal server error"
// @Router       /admin/users [post]
func (h *UserHandler) AdminCreate(c *gin.Context) {
	var input domain.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	user, err := h.users.AdminCreate(c.Request.Context(), input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User created successfully",
		"data":    user,
	})
}

// AdminList godoc
// @Summary      List users
// @Description  List users with optional search, role filter, and pagination.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        search     query  string  false  "Search by name or email"
// @Param        role       query  string  false  "Filter by role"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Success      200
// @Failure      400  "Invalid pagination"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      500  "Internal server error"
// @Router       /admin/users [get]
func (h *UserHandler) AdminList(c *gin.Context) {
	page, pageSize, ok := parsePagination(c)
	if !ok {
		return
	}

	response, err := h.users.AdminList(
		c.Request.Context(),
		c.Query("search"),
		c.Query("role"),
		page,
		pageSize,
	)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

// AdminUpdate godoc
// @Summary      Update user
// @Description  Update a user account (admin only).
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                        true  "User ID"
// @Param        body  body  domain.AdminUpdateUserRequest  true  "User fields to update"
// @Success      200
// @Failure      400  "Invalid user id or validation error"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "User not found"
// @Failure      500  "Internal server error"
// @Router       /admin/users/{id} [patch]
func (h *UserHandler) AdminUpdate(c *gin.Context) {
	claims, ok := currentClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id", "status": "error"})
		return
	}

	var input domain.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}
	user, err := h.users.AdminUpdate(c.Request.Context(), id, claims.UserID, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User updated successfully",
		"data":    user,
	})
}

// AdminStats godoc
// @Summary      Admin dashboard stats
// @Description  Get aggregate user and platform statistics (admin only).
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      500  "Internal server error"
// @Router       /admin/stats [get]
func (h *UserHandler) AdminStats(c *gin.Context) {
	stats, err := h.users.AdminStats(c.Request.Context())
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": stats})
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "status": "error"})
	case errors.Is(err, service.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered", "status": "error"})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid current password", "status": "error"})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
	case errors.Is(err, service.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile", "status": "error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error"})
	}
}
