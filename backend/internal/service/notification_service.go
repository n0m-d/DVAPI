package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/n0m-d/DVAPI/internal/repository"
)

type NotificationService interface {
	List(ctx context.Context, userID uuid.UUID, unreadOnly bool, page, pageSize int) (domain.NotificationListResponse, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) (domain.Notification, error)
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type notificationService struct {
	notifications repository.NotificationRepository
}

func NewNotificationService(notifications repository.NotificationRepository) NotificationService {
	return &notificationService{notifications: notifications}
}

func (s *notificationService) List(
	ctx context.Context,
	userID uuid.UUID,
	unreadOnly bool,
	page, pageSize int,
) (domain.NotificationListResponse, error) {
	page, pageSize = normalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	items, err := s.notifications.ListByUser(ctx, repository.NotificationListParams{
		UserID:     userID,
		UnreadOnly: unreadOnly,
		Limit:      pageSize,
		Offset:     offset,
	})
	if err != nil {
		return domain.NotificationListResponse{}, fmt.Errorf("list notifications: %w", err)
	}
	total, err := s.notifications.CountByUser(ctx, userID, unreadOnly)
	if err != nil {
		return domain.NotificationListResponse{}, fmt.Errorf("count notifications: %w", err)
	}
	unreadCount, err := s.notifications.CountUnread(ctx, userID)
	if err != nil {
		return domain.NotificationListResponse{}, fmt.Errorf("count unread notifications: %w", err)
	}
	pagination := domain.NewPagination(total, page, pageSize)
	return domain.NotificationListResponse{
		Status: "success",
		Data: &domain.NotificationListData{
			Notifications: items,
			UnreadCount:   unreadCount,
			Pagination:    &pagination,
		},
	}, nil
}

func (s *notificationService) MarkRead(ctx context.Context, id, userID uuid.UUID) (domain.Notification, error) {
	if id == uuid.Nil || userID == uuid.Nil {
		return domain.Notification{}, fmt.Errorf("%w: notification id is required", ErrInvalidInput)
	}
	n, err := s.notifications.MarkRead(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return domain.Notification{}, ErrNotFound
		}
		return domain.Notification{}, err
	}
	return n, nil
}

func (s *notificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}
	_, err := s.notifications.MarkAllRead(ctx, userID)
	return err
}

func (s *notificationService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: notification id is required", ErrInvalidInput)
	}
	if err := s.notifications.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
