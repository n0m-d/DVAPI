package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/email"
	"github.com/n0m-d/DVAPI/internal/repository"
	"github.com/n0m-d/DVAPI/internal/service"
	"github.com/n0m-d/DVAPI/internal/utils"
)

type PasswordResetHandler struct {
	svc    service.PasswordResetService
	digits int
}

type PasswordResetConfig struct {
	Users  repository.UserRepository
	OTPs   repository.PasswordResetRepository
	Mailer email.Sender
	Digits int
	Log    *slog.Logger
}

func (c PasswordResetConfig) WithDigits(digits int) PasswordResetConfig {
	c.Digits = digits
	return c
}

func NewPasswordResetHandler(cfg PasswordResetConfig) *PasswordResetHandler {
	return &PasswordResetHandler{
		svc:    service.NewPasswordResetService(cfg.Users, cfg.OTPs, cfg.Mailer, cfg.Digits, cfg.Log),
		digits: cfg.Digits,
	}
}

type emailRequest struct {
	Email string `json:"email" binding:"required" example:"user@example.com"` //,email binding removed for SMTP CRLF Injection
}

type verifyOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	OTP   string `json:"otp" binding:"required,numeric" example:"123456"`
}

type resetPasswordRequest struct {
	Email           string `json:"email" binding:"required,email" example:"user@example.com"`
	Password        string `json:"password" binding:"required,min=8" example:"newpassword123"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password" example:"newpassword123"`
}

// Request godoc
// @Summary      Request password reset
// @Description  Send a one-time password (OTP) to the user's email for password reset.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  emailRequest  true  "Account email"
// @Success      200   "OTP issued"
// @Failure      400   "Validation error"
// @Failure      404   "User not found"
// @Failure      500   "Internal server error"
// @Router       /auth/password-reset/request [post]
func (h *PasswordResetHandler) Request(c *gin.Context) {
	var req emailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.MapError(err), "status": "error"})
		return
	}

	if err := h.svc.Request(c.Request.Context(), req.Email); err != nil {
		respondPasswordResetError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User found. OTP has been issued",
		"status":  "success",
	})
}

// Verify godoc
// @Summary      Verify password reset OTP
// @Description  Verify the OTP sent to the user's email before resetting the password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  verifyOTPRequest  true  "Email and OTP"
// @Success      200   "OTP verified successfully"
// @Failure      400   "Validation error or invalid/expired OTP"
// @Failure      404   "User not found"
// @Failure      500   "Internal server error"
// @Router       /auth/password-reset/verify [post]
func (h *PasswordResetHandler) Verify(c *gin.Context) {
	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.MapError(err), "status": "error"})
		return
	}
	if len(req.OTP) != h.digits {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  fmt.Sprintf("OTP must be %d digits", h.digits),
			"status": "error",
		})
		return
	}

	if err := h.svc.Verify(c.Request.Context(), req.Email, req.OTP); err != nil {
		respondPasswordResetError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully",
		"status":  "success",
	})
}

// Confirm godoc
// @Summary      Confirm password reset
// @Description  Set a new password after the OTP has been verified.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  resetPasswordRequest  true  "Email and new password"
// @Success      200   "Password reset successfully"
// @Failure      400   "Validation error or OTP not verified"
// @Failure      404   "User not found"
// @Failure      500   "Internal server error"
// @Router       /auth/password-reset/confirm [post]
func (h *PasswordResetHandler) Confirm(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.MapError(err), "status": "error"})
		return
	}

	if err := h.svc.Reset(c.Request.Context(), req.Email, req.Password, req.ConfirmPassword); err != nil {
		respondPasswordResetError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
		"status":  "success",
	})
}

func respondPasswordResetError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "status": "error"})
	case errors.Is(err, service.ErrOTPInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP", "status": "error"})
	case errors.Is(err, service.ErrOTPNotVerified):
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP verification required", "status": "error"})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "status": "error", "err": err.Error()})
	}
}
