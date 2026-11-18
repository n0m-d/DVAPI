package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/domain"
)

type mappedError struct {
	target  error
	status  int
	message string // empty means use err.Error()
}

var serviceErrors = []mappedError{
	{domain.ErrNotFound, http.StatusNotFound, "Resource not found"},
	{domain.ErrAlreadyExists, http.StatusConflict, "Already exists"},
	{domain.ErrForbidden, http.StatusForbidden, "Forbidden"},
	{domain.ErrNotEnrolled, http.StatusForbidden, "Not enrolled in this course"},
	{domain.ErrAssignmentClosed, http.StatusBadRequest, "Assignment is not open for submissions"},
	{domain.ErrPastDue, http.StatusBadRequest, "Assignment due date has passed"},
	{domain.ErrOTPInvalid, http.StatusBadRequest, "Invalid or expired OTP"},
	{domain.ErrOTPNotVerified, http.StatusBadRequest, "OTP verification required"},
	{domain.ErrInvalidCredentials, http.StatusUnauthorized, "Invalid credentials"},
	{domain.ErrUnauthorized, http.StatusUnauthorized, "Unauthorized"},
	{domain.ErrInvalidInput, http.StatusBadRequest, ""},
}

func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	msg := "Internal server error"
	for _, m := range serviceErrors {
		if errors.Is(err, m.target) {
			status = m.status
			if m.message != "" {
				msg = m.message
			} else {
				msg = err.Error()
			}
			break
		}
	}
	c.JSON(status, gin.H{"error": msg, "status": "error"})
}
