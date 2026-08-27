package redis

import (
	"context"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
	"github.com/google/uuid"
)

type TokenCache struct {
	rdb *Redis
}

func NewTokenCache() *TokenCache {
	return &TokenCache{rdb: DB()}
}

var _ biz.TokenCache = (*TokenCache)(nil)

func (t *TokenCache) accessKey(sessionID uuid.UUID) string {
	return consts.RedisAccessToken + ":" + sessionID.String()
}

func (t *TokenCache) SetAccess(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error {
	return t.rdb.Set(ctx, t.accessKey(sessionID), "1", ttl)
}

func (t *TokenCache) ExistsAccess(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	return t.rdb.Exists(ctx, t.accessKey(sessionID))
}

func (t *TokenCache) DelAccess(ctx context.Context, sessionID uuid.UUID) error {
	return t.rdb.Del(ctx, t.accessKey(sessionID))
}

func (t *TokenCache) refreshKey(sessionID uuid.UUID) string {
	return consts.RedisRefreshToken + ":" + sessionID.String()
}

func (t *TokenCache) SetRefresh(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error {
	return t.rdb.Set(ctx, t.refreshKey(sessionID), "1", ttl)
}

func (t *TokenCache) ExistsRefresh(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	return t.rdb.Exists(ctx, t.refreshKey(sessionID))
}

func (t *TokenCache) DelRefresh(ctx context.Context, sessionID uuid.UUID) error {
	return t.rdb.Del(ctx, t.refreshKey(sessionID))
}
