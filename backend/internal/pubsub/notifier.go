package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/redis/go-redis/v9"
)

const userChannelPrefix = "notifications:user:"

// Notifier publishes and fans out user notification events.
type Notifier interface {
	Publish(ctx context.Context, userID uuid.UUID, n domain.Notification) error
	// PublishMany delivers already-persisted notifications to many users in one Redis pipeline round-trip.
	PublishMany(ctx context.Context, userIDs []uuid.UUID, notifications []domain.Notification) error
	Subscribe(ctx context.Context, userID uuid.UUID) (<-chan domain.Notification, func(), error)
}

type redisNotifier struct {
	client *redis.Client
	log    *slog.Logger
}

func NewNotifier(client *redis.Client, log *slog.Logger) Notifier {
	if log == nil {
		log = slog.Default()
	}
	return &redisNotifier{client: client, log: log}
}

func userChannel(userID uuid.UUID) string {
	return userChannelPrefix + userID.String()
}

func (n *redisNotifier) Publish(ctx context.Context, userID uuid.UUID, notification domain.Notification) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}
	if err := n.client.Publish(ctx, userChannel(userID), payload).Err(); err != nil {
		return fmt.Errorf("publish notification: %w", err)
	}
	return nil
}

func (n *redisNotifier) PublishMany(ctx context.Context, userIDs []uuid.UUID, notifications []domain.Notification) error {
	if len(userIDs) == 0 {
		return nil
	}
	if len(userIDs) != len(notifications) {
		return fmt.Errorf("publish notifications: userIDs and notifications length mismatch")
	}
	pipe := n.client.Pipeline()
	queued := 0
	for i, userID := range userIDs {
		if userID == uuid.Nil {
			continue
		}
		payload, err := json.Marshal(notifications[i])
		if err != nil {
			return fmt.Errorf("marshal notification: %w", err)
		}
		pipe.Publish(ctx, userChannel(userID), payload)
		queued++
	}
	if queued == 0 {
		return nil
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("publish notifications: %w", err)
	}
	return nil
}

func (n *redisNotifier) Subscribe(ctx context.Context, userID uuid.UUID) (<-chan domain.Notification, func(), error) {
	pubsub := n.client.Subscribe(ctx, userChannel(userID))
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, nil, fmt.Errorf("subscribe: %w", err)
	}

	out := make(chan domain.Notification, 16)
	done := make(chan struct{})

	go func() {
		defer close(out)
		ch := pubsub.Channel()
		for {
			select {
			case <-done: // Channel closed by unsubscribe function
				return
			case <-ctx.Done(): // Context cancelled
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var notification domain.Notification
				if err := json.Unmarshal([]byte(msg.Payload), &notification); err != nil {
					n.log.Warn("invalid notification payload", "error", err)
					continue
				}
				select {
				case out <- notification:
				case <-done:
					return
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	unsubscribe := func() {
		select {
		case <-done:
		default:
			close(done)
		}
		_ = pubsub.Close()
	}

	return out, unsubscribe, nil
}
