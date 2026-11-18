package v3

import (
	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/handler"
	"github.com/n0m-d/DVAPI/internal/middleware"
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

	r.POST("/auth/refresh", refreshToken(cfg.JWTSecret))
}
