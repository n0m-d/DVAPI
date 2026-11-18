package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/service"
)

type mappedError struct {
	target  error
	status  int
	message string // empty means use err.Error()
}

var serviceErrors = []mappedError{
	{service.ErrNotFound, http.StatusNotFound, "Resource not found"},
	{service.ErrAlreadyExists, http.StatusConflict, "Already exists"},
	{service.ErrForbidden, http.StatusForbidden, "Forbidden"},
	{service.ErrNotEnrolled, http.StatusForbidden, "Not enrolled in this course"},
	{service.ErrAssignmentClosed, http.StatusBadRequest, "Assignment is not open for submissions"},
	{service.ErrPastDue, http.StatusBadRequest, "Assignment due date has passed"},
	{service.ErrOTPInvalid, http.StatusBadRequest, "Invalid or expired OTP"},
	{service.ErrOTPNotVerified, http.StatusBadRequest, "OTP verification required"},
	{service.ErrInvalidCredentials, http.StatusUnauthorized, "Invalid credentials"},
	{service.ErrUnauthorized, http.StatusUnauthorized, "Unauthorized"},
	{service.ErrInvalidInput, http.StatusBadRequest, ""},
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
