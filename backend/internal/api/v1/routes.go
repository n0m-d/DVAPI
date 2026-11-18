package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/handler"
	"github.com/n0m-d/DVAPI/internal/middleware"
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

	r.GET("/admin/founder-notes", founderNotes)
}
