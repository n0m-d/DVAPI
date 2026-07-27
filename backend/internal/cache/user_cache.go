package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/n0m-d/DVAPI/internal/domain"
	"github.com/redis/go-redis/v9"
)

const userMeTTL = 5 * time.Minute

type UserCache interface {
	GetUser(ctx context.Context, id uuid.UUID) (domain.User, bool, error)
	SetUser(ctx context.Context, user domain.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type redisUserCache struct {
	client *redis.Client
}

func NewUserCache(client *redis.Client) UserCache {
	return &redisUserCache{client: client}
}

func userKey(id uuid.UUID) string {
	return "user:" + id.String()
}

func (c *redisUserCache) GetUser(ctx context.Context, id uuid.UUID) (domain.User, bool, error) {
	raw, err := c.client.Get(ctx, userKey(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, err
	}

	var user domain.User
	if err := json.Unmarshal(raw, &user); err != nil {
		return domain.User{}, false, fmt.Errorf("unmarshal cached user: %w", err)
	}
	return user, true, nil
}

func (c *redisUserCache) SetUser(ctx context.Context, user domain.User) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, userKey(user.ID), payload, userMeTTL).Err()
}

func (c *redisUserCache) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return c.client.Del(ctx, userKey(id)).Err()
}
