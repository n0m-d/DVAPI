package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/n0m-d/DVAPI/internal/config"
	"github.com/n0m-d/DVAPI/internal/database"
	"github.com/n0m-d/DVAPI/internal/jobs"
	"github.com/n0m-d/DVAPI/internal/logger"
	"github.com/n0m-d/DVAPI/internal/server"
	"github.com/redis/go-redis/v9"
)

// @title           DVAPI (Schole)
// @version         1.0
// @description     Damn Vulnerable API
// @host            localhost:8080
// @BasePath        /api/v2
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT. Example: Bearer eyJhbGciOi...

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log, close, err := logger.New(cfg)
	if err != nil {
		slog.Error("failed to create logger", "error", err)
		os.Exit(1)
	}
	defer close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb, err := newRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	srv, jobDeps, err := server.New(cfg, pool, rdb, log)
	if err != nil {
		log.Error("failed to create server", "error", err)
		os.Exit(1)
	}
	defer srv.Close()

	scheduler, err := jobs.New(jobDeps)
	if err != nil {
		log.Error("failed to create job scheduler", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Error("failed to shutdown job scheduler", "error", err)
		}
	}()
	scheduler.Start()

	log.Info("starting server", "addr", cfg.HTTPAddr, "env", cfg.Env)

	if err := srv.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func newRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
