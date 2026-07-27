package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/handler"
	"github.com/n0m-d/DVAPI/internal/middleware"
	"github.com/n0m-d/DVAPI/internal/utils"
)

type Config struct {
	PasswordReset  *handler.PasswordResetHandler
	Users          *handler.UserHandler
	JWTSecret      string
	TokenBlacklist cache.TokenBlacklist
	Course         *handler.CourseHandler
}

func Register(r *gin.RouterGroup, cfg Config) {
	r.POST("/auth/password-reset/request", cfg.PasswordReset.Request)
	r.POST("/auth/password-reset/verify", cfg.PasswordReset.Verify)
	r.POST("/auth/password-reset/confirm", cfg.PasswordReset.Confirm)

	protected := r.Group("/")
	protected.Use(middleware.JWTVerify(cfg.JWTSecret, cfg.TokenBlacklist))

	privOnly := protected.Group("/")
	privOnly.Use(middleware.RoleRestrict("instructor", "admin"))
	privOnly.PATCH("/users/:id", cfg.Users.UpdateProfilev1) // mass assignment route, priv escalation

	r.GET("/admin/founder-notes", func(ctx *gin.Context) {

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

		if claims["role"] != "admin" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Forbidden. Admin Only", "status": "error"})
			return
		}

		iat, ok := claims["iat"].(float64)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid iat",
			})
			return
		}

		/*
			Because jwt.MapClaims stores values as map[string]any,
			 it relies on Go's default encoding/json behavior,
			 which unmarshals all JSON numbers into float64
		*/

		if int64(iat) >= 2147483647 {
			ctx.JSON(http.StatusOK, gin.H{
				"data": "flag{Y2K38_p4s7}",
			})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
			"data": "2³¹ ⏳ Some clocks never make it to tomorrow.",
		})
		return

	})
}
