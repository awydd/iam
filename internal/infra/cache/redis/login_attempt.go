package redis

import (
	"context"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
)

type LoginAttemptCache struct {
	rdb *Redis
}

func NewLoginAttemptCache() *LoginAttemptCache {
	return &LoginAttemptCache{rdb: DB()}
}

var _ biz.LoginAttemptCache = (*LoginAttemptCache)(nil)

func (c *LoginAttemptCache) IncrFailure(ctx context.Context, username string, window time.Duration) (int64, error) {
	key := consts.RedisLoginAttemptFail + username
	count, err := c.rdb.Incr(ctx, key)
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := c.rdb.Expire(ctx, key, window); err != nil {
			return count, err
		}
	}
	return count, nil
}

func (c *LoginAttemptCache) ResetFailure(ctx context.Context, username string) error {
	return c.rdb.Del(ctx, consts.RedisLoginAttemptFail+username)
}

func (c *LoginAttemptCache) Lock(ctx context.Context, username string, duration time.Duration) error {
	key := consts.RedisLoginAttemptLock + username
	return c.rdb.Set(ctx, key, "1", duration)
}

func (c *LoginAttemptCache) IsLocked(ctx context.Context, username string) (bool, time.Duration, error) {
	key := consts.RedisLoginAttemptLock + username
	ttl, err := c.rdb.TTL(ctx, key)
	if err != nil {
		return false, 0, err
	}
	if ttl <= 0 {
		return false, 0, nil
	}
	return true, ttl, nil
}

func (c *LoginAttemptCache) IncrFailureByIP(ctx context.Context, ip string, window time.Duration) (int64, error) {
	key := consts.RedisLoginAttemptFailIP + ip
	count, err := c.rdb.Incr(ctx, key)
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := c.rdb.Expire(ctx, key, window); err != nil {
			return count, err
		}
	}
	return count, nil
}

func (c *LoginAttemptCache) ResetFailureByIP(ctx context.Context, ip string) error {
	return c.rdb.Del(ctx, consts.RedisLoginAttemptFailIP+ip)
}
