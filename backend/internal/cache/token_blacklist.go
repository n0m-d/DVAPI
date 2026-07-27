package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const blacklistPrefix = "blacklist:"

// TokenBlacklist stores revoked JWT IDs (jti) in Redis until they would expire.
type TokenBlacklist interface {
	Blacklist(ctx context.Context, jti string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

type redisTokenBlacklist struct {
	client *redis.Client
}

func NewTokenBlacklist(client *redis.Client) TokenBlacklist {
	return &redisTokenBlacklist{client: client}
}

func (c *redisTokenBlacklist) Blacklist(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Second
	}
	return c.client.Set(ctx, blacklistPrefix+jti, "1", ttl).Err()
}

func (c *redisTokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := c.client.Exists(ctx, blacklistPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
