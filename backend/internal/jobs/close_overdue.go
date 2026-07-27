package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/n0m-d/DVAPI/internal/service"
)

func registerCloseOverdue(s gocron.Scheduler, deps Deps) error {
	_, err := s.NewJob(
		gocron.DurationJob(1*time.Minute),
		gocron.NewTask(func() {
			closeOverdueAssignments(deps.Assignments, deps.Log)
		}),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	return err
}

func closeOverdueAssignments(assignments service.AssignmentService, log *slog.Logger) {
	/*
		Each cron tick is independent: uses a fresh Background context so a previous
		tick's cancel/timeout cannot leak into the next minute's run.
		Cap each run at 30s so a slow DB call cannot hang past the next schedule.
	*/
	started := time.Now()
	log.Info("job started", "job", "close_overdue")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	closed, err := assignments.CloseOverdue(ctx)
	if err != nil {
		log.Error("job failed", "job", "close_overdue", "error", err, "duration", time.Since(started))
		return
	}

	ids := make([]string, 0, len(closed))
	for _, a := range closed {
		ids = append(ids, a.ID.String())
	}
	log.Info("job finished",
		"job", "close_overdue",
		"count", len(closed),
		"ids", ids,
		"duration", time.Since(started),
	)
}
