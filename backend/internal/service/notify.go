package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/pubsub"
	"github.com/n0m-d/DVAPI/internal/repository"
)

const (
	notifyFanOutTimeout = 45 * time.Second
	notifyPageSize      = 500
)

func publishNotification(
	notifications repository.NotificationRepository,
	notifier pubsub.Notifier,
	userID uuid.UUID,
	notification domain.Notification,
) {
	if notifications == nil || notifier == nil || userID == uuid.Nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), notifyFanOutTimeout)
		defer cancel()
		saved, err := notifications.Create(ctx, domain.CreateNotificationInput{
			UserID:   userID,
			Type:     notification.Type,
			Title:    notification.Title,
			Body:     notification.Body,
			CourseID: notification.CourseID,
		})
		if err != nil {
			return
		}
		_ = notifier.Publish(ctx, userID, saved)
	}()
}

// notifyEnrolledStudentsAsync persists then publishes off the request path
// so assignment/announcement responses are not blocked by fan-out.
func notifyEnrolledStudentsAsync(
	courses repository.CourseRepository,
	notifications repository.NotificationRepository,
	notifier pubsub.Notifier,
	courseID uuid.UUID,
	notification domain.Notification,
) {
	if courses == nil || notifications == nil || notifier == nil || courseID == uuid.Nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), notifyFanOutTimeout)
		defer cancel()
		input := domain.CreateNotificationInput{
			Type:     notification.Type,
			Title:    notification.Title,
			Body:     notification.Body,
			CourseID: notification.CourseID,
		}
		for offset := 0; ; offset += notifyPageSize {
			students, err := courses.ListEnrolledStudents(ctx, repository.EnrolledStudentsListParams{
				CourseID: courseID,
				Limit:    notifyPageSize,
				Offset:   offset,
			})
			if err != nil || len(students) == 0 {
				return
			}
			userIDs := make([]uuid.UUID, len(students))
			for i, student := range students {
				userIDs[i] = student.ID
			}
			saved, recipients, err := notifications.CreateMany(ctx, userIDs, input)
			if err != nil {
				return
			}
			_ = notifier.PublishMany(ctx, recipients, saved)
			if len(students) < notifyPageSize {
				return
			}
		}
	}()
}

func courseIDPtr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
