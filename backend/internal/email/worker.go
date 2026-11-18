package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

// Worker is an asynq server that sends queued mail over SMTP.
type Worker struct {
	srv *asynq.Server
	mux *asynq.ServeMux
	log *slog.Logger
}

func NewWorker(redisURL string, sender Sender, log *slog.Logger) (*Worker, error) {
	if log == nil {
		log = slog.Default()
	}
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis uri: %w", err)
	}

	srv := asynq.NewServer(opt, asynq.Config{
		Concurrency: 4,
		Queues: map[string]int{
			queueCritical: 6,
			queueDefault:  3,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(_ context.Context, task *asynq.Task, err error) {
			log.Error("mail task failed", "type", task.Type(), "error", err)
		}),
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(taskEnrollment, handleEnrollment(sender))
	mux.HandleFunc(taskOTP, handleOTP(sender))
	mux.HandleFunc(taskWelcome, handleWelcome(sender))
	mux.HandleFunc(taskGrade, handleGrade(sender))

	return &Worker{srv: srv, mux: mux, log: log}, nil
}

// Run starts the asynq server and blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	w.log.Info("mail worker started", "queues", []string{queueCritical, queueDefault})
	if err := w.srv.Start(w.mux); err != nil {
		w.log.Error("mail worker failed to start", "error", err)
		return
	}
	<-ctx.Done()
	w.srv.Shutdown()
	w.log.Info("mail worker stopped")
}

func handleEnrollment(sender Sender) asynq.HandlerFunc {
	return func(_ context.Context, t *asynq.Task) error {
		j, err := decodeJob(t)
		if err != nil {
			return fmt.Errorf("invalid enrollment payload: %v: %w", err, asynq.SkipRetry)
		}
		return sendEnrollmentEmail(sender, j.To, j.Name, j.CourseName)
	}
}

func handleOTP(sender Sender) asynq.HandlerFunc {
	return func(_ context.Context, t *asynq.Task) error {
		j, err := decodeJob(t)
		if err != nil {
			return fmt.Errorf("invalid otp payload: %v: %w", err, asynq.SkipRetry)
		}
		return sendOTPEmail(sender, j.To, j.Name, j.OTP, j.CustomMsg)
	}
}

func handleWelcome(sender Sender) asynq.HandlerFunc {
	return func(_ context.Context, t *asynq.Task) error {
		j, err := decodeJob(t)
		if err != nil {
			return fmt.Errorf("invalid welcome payload: %v: %w", err, asynq.SkipRetry)
		}
		return sendWelcomeEmail(sender, j.To, j.Name)
	}
}

func handleGrade(sender Sender) asynq.HandlerFunc {
	return func(_ context.Context, t *asynq.Task) error {
		j, err := decodeJob(t)
		if err != nil {
			return fmt.Errorf("invalid grade payload: %v: %w", err, asynq.SkipRetry)
		}
		return sendGradeEmail(sender, j.To, j.Name, j.CourseName, j.AssignmentTitle, j.Grade, j.Feedback)
	}
}

func decodeJob(t *asynq.Task) (job, error) {
	var j job
	if err := json.Unmarshal(t.Payload(), &j); err != nil {
		return job{}, err
	}
	return j, nil
}
