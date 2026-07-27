package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

type NotificationListParams struct {
	UserID     uuid.UUID
	UnreadOnly bool
	Limit      int
	Offset     int
}

type NotificationRepository interface {
	Create(ctx context.Context, input domain.CreateNotificationInput) (domain.Notification, error)
	CreateMany(ctx context.Context, userIDs []uuid.UUID, input domain.CreateNotificationInput) ([]domain.Notification, []uuid.UUID, error)
	ListByUser(ctx context.Context, params NotificationListParams) ([]domain.Notification, error)
	CountByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool) (int, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) (domain.Notification, error)
	MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type notificationRepository struct {
	queries *db.Queries
}

func NewNotificationRepository(queries *db.Queries) NotificationRepository {
	return &notificationRepository{queries: queries}
}

func (r *notificationRepository) Create(ctx context.Context, input domain.CreateNotificationInput) (domain.Notification, error) {
	row, err := r.queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:   pgUUID(input.UserID),
		Type:     input.Type,
		Title:    input.Title,
		Body:     input.Body,
		CourseID: optionalPgUUID(input.CourseID),
	})
	if err != nil {
		return domain.Notification{}, err
	}
	return toDomainNotification(row)
}

func (r *notificationRepository) CreateMany(
	ctx context.Context,
	userIDs []uuid.UUID,
	input domain.CreateNotificationInput,
) ([]domain.Notification, []uuid.UUID, error) {
	if len(userIDs) == 0 {
		return nil, nil, nil
	}
	pgIDs := make([]pgtype.UUID, 0, len(userIDs))
	for _, id := range userIDs {
		if id == uuid.Nil {
			continue
		}
		pgIDs = append(pgIDs, pgUUID(id))
	}
	if len(pgIDs) == 0 {
		return nil, nil, nil
	}
	rows, err := r.queries.CreateNotifications(ctx, db.CreateNotificationsParams{
		Type:     input.Type,
		Title:    input.Title,
		Body:     input.Body,
		CourseID: optionalPgUUID(input.CourseID),
		UserIds:  pgIDs,
	})
	if err != nil {
		return nil, nil, err
	}
	notifications := make([]domain.Notification, 0, len(rows))
	recipients := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		n, err := toDomainNotification(row)
		if err != nil {
			return nil, nil, err
		}
		userID, err := fromPgUUID(row.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("notification user id: %w", err)
		}
		notifications = append(notifications, n)
		recipients = append(recipients, userID)
	}
	return notifications, recipients, nil
}

func (r *notificationRepository) ListByUser(ctx context.Context, params NotificationListParams) ([]domain.Notification, error) {
	rows, err := r.queries.ListNotificationsByUser(ctx, db.ListNotificationsByUserParams{
		UserID:      pgUUID(params.UserID),
		UnreadOnly:  params.UnreadOnly,
		LimitCount:  int32(params.Limit),
		OffsetCount: int32(params.Offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		n, err := toDomainNotification(row)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func (r *notificationRepository) CountByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool) (int, error) {
	total, err := r.queries.CountNotificationsByUser(ctx, db.CountNotificationsByUserParams{
		UserID:     pgUUID(userID),
		UnreadOnly: unreadOnly,
	})
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	total, err := r.queries.CountUnreadNotificationsByUser(ctx, pgUUID(userID))
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *notificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) (domain.Notification, error) {
	row, err := r.queries.MarkNotificationRead(ctx, db.MarkNotificationReadParams{
		ID:     pgUUID(id),
		UserID: pgUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Notification{}, ErrNotificationNotFound
		}
		return domain.Notification{}, err
	}
	return toDomainNotification(row)
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.MarkAllNotificationsRead(ctx, pgUUID(userID))
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := r.queries.DeleteNotification(ctx, pgUUID(id))
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func optionalPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgUUID(*id)
}

func toDomainNotification(row db.Notification) (domain.Notification, error) {
	id, err := fromPgUUID(row.ID)
	if err != nil {
		return domain.Notification{}, fmt.Errorf("notification id: %w", err)
	}
	var courseID *uuid.UUID
	if row.CourseID.Valid {
		cid, err := fromPgUUID(row.CourseID)
		if err != nil {
			return domain.Notification{}, fmt.Errorf("notification course id: %w", err)
		}
		courseID = &cid
	}
	return domain.Notification{
		ID:        id.String(),
		Type:      row.Type,
		Title:     row.Title,
		Body:      row.Body,
		Read:      row.Read,
		CreatedAt: row.CreatedAt.Time.UTC(),
		CourseID:  courseID,
	}, nil
}
