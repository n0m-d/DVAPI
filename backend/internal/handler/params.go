package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/middleware"
)

func currentClaims(c *gin.Context) (*domain.Claims, bool) {
	raw, ok := c.Get(middleware.ClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return nil, false
	}
	claims, ok := raw.(*domain.Claims)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "status": "error"})
		return nil, false
	}
	return claims, true
}

func pathUUID(c *gin.Context, param, resource string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + resource + " id", "status": "error"})
		return uuid.Nil, false
	}
	return id, true
}
