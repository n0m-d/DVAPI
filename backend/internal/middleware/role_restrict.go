package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/domain"
)

// RoleRestrict allows the request only if JWT claims.Role is one of the given roles.
// Must run after JWTVerify so claims are present in the context.
func RoleRestrict(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(ClaimsKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing claims", "status": "error"})
			return
		}

		claims, ok := raw.(*domain.Claims)
		if !ok || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing claims", "status": "error"})
			return
		}

		if !slices.Contains(roles, claims.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden. You are not authorized to access this resource.", "status": "error"})
			return
		}

		c.Next()
	}
}
