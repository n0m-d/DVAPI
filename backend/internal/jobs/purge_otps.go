package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/n0m-d/DVAPI/internal/service"
)

func registerPurgeOTP(s gocron.Scheduler, deps Deps) error {
	_, err := s.NewJob(
		gocron.DurationJob(10*time.Minute),
		gocron.NewTask(func() {
			purgeOTP(deps.Reset, deps.Log)
		}),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	return err
}

func purgeOTP(reset service.PasswordResetService, log *slog.Logger) {
	started := time.Now()
	log.Info("job started", "job", "purge_otps")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count, err := reset.PurgeOTP(ctx)
	if err != nil {
		log.Error("job failed", "job", "purge_otps", "error", err, "duration", time.Since(started))
		return
	}

	log.Info("job finished",
		"job", "purge_otps",
		"count", count,
		"duration", time.Since(started),
	)
}
