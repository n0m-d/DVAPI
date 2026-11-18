package v3

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/n0m-d/DVAPI/internal/utils"
)

func refreshToken(secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := utils.BearerToken(ctx.GetHeader("Authorization"))
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token", "status": "error"})
			return
		}

		// VULN(JWT): unverified claims are re-signed with the real secret.
		parser := jwt.NewParser()

		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Something went wrong.", "status": "error"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		now := time.Now()

		claims["iat"] = now.Unix()
		claims["exp"] = now.Add(24 * time.Hour).Unix()

		tokenNew := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := tokenNew.SignedString([]byte(secret))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Something went wrong.", "status": "error"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"token": signedToken, "status": "success"})
	}
}
