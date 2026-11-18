package email

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const (
	taskEnrollment = "email:enrollment"
	taskOTP        = "email:otp"
	taskWelcome    = "email:welcome"
	taskGrade      = "email:grade"

	queueCritical = "critical"
	queueDefault  = "default"
)

type job struct {
	To              string `json:"to"`
	Name            string `json:"name"`
	CourseName      string `json:"course_name,omitempty"`
	OTP             string `json:"otp,omitempty"`
	CustomMsg       string `json:"custom_msg,omitempty"`
	AssignmentTitle string `json:"assignment_title,omitempty"`
	Grade           int    `json:"grade,omitempty"`
	Feedback        string `json:"feedback,omitempty"`
}

// Queue enqueues mail jobs onto asynq.
type Queue interface {
	EnqueueEnrollment(ctx context.Context, to, name, courseName string) error
	EnqueueOTP(ctx context.Context, to, name, otp, customMsg string) error
	EnqueueWelcome(ctx context.Context, to, name string) error
	EnqueueGrade(ctx context.Context, to, name, courseName, assignmentTitle string, grade int, feedback string) error
	Close() error
}

type asynqQueue struct {
	client *asynq.Client
}

func NewQueue(redisURL string) (Queue, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis uri: %w", err)
	}
	return &asynqQueue{client: asynq.NewClient(opt)}, nil
}

func (q *asynqQueue) EnqueueEnrollment(ctx context.Context, to, name, courseName string) error {
	return q.enqueue(ctx, taskEnrollment, job{To: to, Name: name, CourseName: courseName},
		asynq.Queue(queueDefault),
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
	)
}

func (q *asynqQueue) EnqueueOTP(ctx context.Context, to, name, otp, customMsg string) error {
	return q.enqueue(ctx, taskOTP, job{To: to, Name: name, OTP: otp, CustomMsg: customMsg},
		asynq.Queue(queueCritical),
		asynq.MaxRetry(8),
		asynq.Timeout(15*time.Second),
	)
}

func (q *asynqQueue) EnqueueWelcome(ctx context.Context, to, name string) error {
	return q.enqueue(ctx, taskWelcome, job{To: to, Name: name},
		asynq.Queue(queueDefault),
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
	)
}

func (q *asynqQueue) EnqueueGrade(ctx context.Context, to, name, courseName, assignmentTitle string, grade int, feedback string) error {
	return q.enqueue(ctx, taskGrade, job{
		To:              to,
		Name:            name,
		CourseName:      courseName,
		AssignmentTitle: assignmentTitle,
		Grade:           grade,
		Feedback:        feedback,
	},
		asynq.Queue(queueDefault),
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
	)
}

func (q *asynqQueue) enqueue(ctx context.Context, taskType string, payload job, opts ...asynq.Option) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal mail job: %w", err)
	}
	if _, err := q.client.EnqueueContext(ctx, asynq.NewTask(taskType, body), opts...); err != nil {
		return fmt.Errorf("enqueue %s: %w", taskType, err)
	}
	return nil
}

func (q *asynqQueue) Close() error {
	if q == nil || q.client == nil {
		return nil
	}
	return q.client.Close()
}
