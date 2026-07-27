package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/utils"
)

const (
	ClaimsKey = "claims"
)

func JWTVerify(secret string, tokenBlacklist cache.TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.BearerToken(c.GetHeader("Authorization"))
		// EventSource cannot set Authorization headers; allow ?token= for SSE streams.
		if token == "" {
			token = strings.TrimSpace(c.Query("token"))
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token", "status": "error"})
			return
		}

		claims, err := utils.ParseToken(token, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "status": "error"})
			return
		}

		blacklisted, err := tokenBlacklist.IsBlacklisted(c.Request.Context(), claims.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify token", "status": "error"})
			return
		}
		if blacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked/expired. Please login again.", "status": "error"})
			return
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}
