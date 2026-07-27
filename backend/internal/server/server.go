package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/n0m-d/DVAPI/docs"
	apiv1 "github.com/n0m-d/DVAPI/internal/api/v1"
	apiv2 "github.com/n0m-d/DVAPI/internal/api/v2"
	apiv3 "github.com/n0m-d/DVAPI/internal/api/v3"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/config"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/email"
	"github.com/n0m-d/DVAPI/internal/handler"
	"github.com/n0m-d/DVAPI/internal/jobs"
	"github.com/n0m-d/DVAPI/internal/middleware"
	"github.com/n0m-d/DVAPI/internal/pubsub"
	"github.com/n0m-d/DVAPI/internal/repository"
	"github.com/n0m-d/DVAPI/internal/service"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router *graceful.Graceful
}

func New(cfg *config.Config, pool *pgxpool.Pool, rdb *redis.Client, log *slog.Logger) (*Server, jobs.Deps, error) {
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	router, err := graceful.New(
		gin.New(),
		graceful.WithAddr(cfg.HTTPAddr),
		graceful.WithShutdownTimeout(30*time.Second),
	)
	if err != nil {
		return nil, jobs.Deps{}, err
	}

	router.Use(gin.Recovery())

	router.Use(
		middleware.RequestID(),
		middleware.Logger(log),
		cors.New(cors.Config{
			AllowOrigins:     cfg.CORSOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", middleware.RequestIDHeader},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
	)

	router.MaxMultipartMemory = 8 << 20 // 8 MiB, upload limit for file uploads

	health := handler.NewHealthHandler(pool, rdb)
	router.GET("/healthz", health.Liveness)
	router.GET("/readyz", health.Readiness)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Caches
	userCache := cache.NewUserCache(rdb)
	tokenBlacklist := cache.NewTokenBlacklist(rdb)
	notifier := pubsub.NewNotifier(rdb, log)

	queries := db.New(pool)

	// Repositories
	courseRepo := repository.NewCourseRepository(queries)
	userRepo := repository.NewUserRepository(queries)
	otpRepo := repository.NewPasswordResetRepository(queries)
	assignmentRepo := repository.NewAssignmentRepository(queries)
	learningRepo := repository.NewLearningRepository(queries)
	notificationRepo := repository.NewNotificationRepository(queries)
	noteRepo := repository.NewNoteRepository(queries, pool)

	mailer := email.NewSMTPSender(cfg.SMTPConfig.SMTPHost, cfg.SMTPConfig.SMTPPort, cfg.SMTPConfig.SMTPUser, cfg.SMTPConfig.SMTPPass, cfg.SMTPConfig.SMTPFrom)
	resetCfg := handler.PasswordResetConfig{
		Users:  userRepo,
		OTPs:   otpRepo,
		Mailer: mailer,
		Log:    log,
	}
	// Handlers
	userHandler := handler.NewUserHandler(service.NewUserService(userRepo, userCache))
	authHandler := handler.NewAuthHandler(service.NewAuthService(userRepo, cfg, tokenBlacklist))
	courseHandler := handler.NewCourseHandler(service.NewCourseService(courseRepo))
	assignmentSvc := service.NewAssignmentService(assignmentRepo, courseRepo, notificationRepo, notifier)
	assignmentHandler := handler.NewAssignmentHandler(assignmentSvc)
	learningHandler := handler.NewLearningHandler(service.NewLearningService(learningRepo, courseRepo, assignmentRepo, notificationRepo, notifier))
	notificationHandler := handler.NewNotificationHandler(service.NewNotificationService(notificationRepo), notifier)
	// Digits only matter for issue/verify; purge is shared across v1/v2.
	resetSvc := service.NewPasswordResetService(userRepo, otpRepo, mailer, 6, log)
	resetV1 := handler.NewPasswordResetHandler(resetCfg.WithDigits(4))
	resetV2 := handler.NewPasswordResetHandler(resetCfg.WithDigits(6))
	v3Handler := handler.NewV3Handler(service.NewCourseService(courseRepo))

	/* Basic Rate Limiter
	   TODO: Improve rate limiting further.
	*/
	rateLimiter := redis_rate.NewLimiter(rdb)

	apiv1.Register(router.Group("/api/v1"), apiv1.Config{
		PasswordReset:  resetV1,
		Users:          userHandler,
		JWTSecret:      cfg.JWT_SECRET,
		TokenBlacklist: tokenBlacklist,
		Course:         courseHandler,
	})
	apiv2.Register(router.Group("/api/v2"), apiv2.Config{
		Auth:           authHandler,
		Users:          userHandler,
		PasswordReset:  resetV2,
		Admin:          handler.NewAdminHandler(cfg.LogFile),
		JWTSecret:      cfg.JWT_SECRET,
		TokenBlacklist: tokenBlacklist,
		Course:         courseHandler,
		Assignments:    assignmentHandler,
		Learning:       learningHandler,
		Notifications:  notificationHandler,
		Notes:          service.NewNoteService(noteRepo),
		RateLimiter:    rateLimiter,
	})

	apiv3.Register(router.Group("/api/v3/beta/"), apiv3.Config{
		JWTSecret:      cfg.JWT_SECRET,
		TokenBlacklist: tokenBlacklist,
		V3:             v3Handler,
	})

	docs.SwaggerInfo.BasePath = "/api/v2"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	return &Server{router: router}, jobs.Deps{
		Assignments: assignmentSvc,
		Reset:       resetSvc,
		Log:         log,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	return s.router.RunWithContext(ctx)
}

func (s *Server) Close() {
	s.router.Close()
}
