package v3

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/handler"
	"github.com/n0m-d/DVAPI/internal/middleware"
	"github.com/n0m-d/DVAPI/internal/utils"
)

type Config struct {
	JWTSecret      string
	V3             *handler.V3Handler
	TokenBlacklist cache.TokenBlacklist
}

func Register(r *gin.RouterGroup, cfg Config) {
	protected := r.Group("/")
	protected.Use(middleware.JWTVerify(cfg.JWTSecret, cfg.TokenBlacklist))

	beta := protected.Group("/")
	beta.POST("/course/enroll", cfg.V3.EnrollBeta)
	beta.DELETE("/course/unenroll", cfg.V3.UnenrollBeta)

	r.POST("/auth/refresh", func(ctx *gin.Context) {

		tokenString := utils.BearerToken(ctx.GetHeader("Authorization"))
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token", "status": "error"})
			return
		}

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
		signedToken, err := tokenNew.SignedString([]byte(cfg.JWTSecret))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Something went wrong.", "status": "error"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"token": signedToken, "status": "success"})

	})
}
