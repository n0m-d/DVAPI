package jobs

import (
	"log/slog"

	"github.com/go-co-op/gocron/v2"
	"github.com/n0m-d/DVAPI/internal/service"
)

// Deps holds dependencies shared by scheduled jobs.
type Deps struct {
	Assignments service.AssignmentService
	Reset       service.PasswordResetService
	Log         *slog.Logger
}

// Scheduler registers and runs background jobs.
type Scheduler struct {
	scheduler gocron.Scheduler
	deps      Deps
}

// New creates a scheduler and registers all jobs.
func New(deps Deps) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	if err := registerCloseOverdue(s, deps); err != nil {
		_ = s.Shutdown()
		return nil, err
	}
	if err := registerPurgeOTP(s, deps); err != nil {
		_ = s.Shutdown()
		return nil, err
	}

	return &Scheduler{scheduler: s, deps: deps}, nil // return the scheduler with the dependencies
}

func (s *Scheduler) Start() {
	s.deps.Log.Info("starting job scheduler", "jobs", []string{"close_overdue", "purge_otps"})
	s.scheduler.Start()
	// Run once on startup so work is not delayed until the first tick.
	closeOverdueAssignments(s.deps.Assignments, s.deps.Log)
	purgeOTP(s.deps.Reset, s.deps.Log)
}

func (s *Scheduler) Shutdown() error {
	s.deps.Log.Info("shutting down job scheduler")
	return s.scheduler.Shutdown()
}
