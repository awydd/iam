package redis

import (
	"context"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
	"github.com/google/uuid"
)

type LastActiveCache struct {
	rdb *Redis
}

func NewLastActiveCache() *LastActiveCache {
	return &LastActiveCache{rdb: DB()}
}

var _ biz.LastActiveCache = (*LastActiveCache)(nil)

func (c *LastActiveCache) key(sid string) string {
	return consts.RedisLastActiveThrottle + ":" + sid
}

func (c *LastActiveCache) ShouldUpdate(ctx context.Context, sessionID uuid.UUID, window time.Duration) (bool, error) {
	ok, err := c.rdb.SetNX(ctx, c.key(sessionID.String()), 1, window)
	if err != nil {
		return false, err
	}
	return ok, nil
}
