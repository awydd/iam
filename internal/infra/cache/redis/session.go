package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
)

type SessionCache struct {
	rdb *Redis
}

func NewSessionCache() *SessionCache {
	return &SessionCache{rdb: DB()}
}

var _ biz.SessionCache = (*SessionCache)(nil)

func (c *SessionCache) sessionKey(sid string) string {
	return consts.RedisSession + ":" + sid
}

func (c *SessionCache) userSessionsKey(userID int) string {
	return consts.RedisSession + ":user:" + strconv.Itoa(userID)
}

func (c *SessionCache) Get(ctx context.Context, sid string) (*biz.SessionCachePayload, error) {
	var payload biz.SessionCachePayload
	err := c.rdb.GetJSON(ctx, c.sessionKey(sid), &payload)
	if err != nil {
		if errors.Is(err, ErrNil) {
			return nil, nil
		}
		return nil, err
	}
	return &payload, nil
}

func (c *SessionCache) Set(ctx context.Context, sid string, payload biz.SessionCachePayload, ttl time.Duration) error {
	payload.ExpiredAt = time.Now().Add(ttl).Unix()

	if err := c.rdb.SetJSON(ctx, c.sessionKey(sid), payload, ttl); err != nil {
		return err
	}

	userKey := c.userSessionsKey(payload.UserID)
	if err := c.rdb.SAdd(ctx, userKey, sid); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, userKey, ttl)
}

func (c *SessionCache) Del(ctx context.Context, sid string) error {
	payload, err := c.Get(ctx, sid)
	if err == nil && payload != nil {
		if err := c.rdb.SRem(ctx, c.userSessionsKey(payload.UserID), sid); err != nil {
			return err
		}
	}
	return c.rdb.Del(ctx, c.sessionKey(sid))
}

func (c *SessionCache) DelAllByUser(ctx context.Context, userID int) error {
	userKey := c.userSessionsKey(userID)
	sids, err := c.rdb.SMembers(ctx, userKey)
	if err != nil {
		return err
	}

	for _, sid := range sids {
		if err := c.rdb.Del(ctx, c.sessionKey(sid)); err != nil {
			return err
		}
	}

	return c.rdb.Del(ctx, userKey)
}

func (c *SessionCache) Refresh(ctx context.Context, sid string, payload *biz.SessionCachePayload, ttl time.Duration) error {
	payload.ExpiredAt = time.Now().Add(ttl).Unix()

	if err := c.rdb.SetJSON(ctx, c.sessionKey(sid), *payload, ttl); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, c.userSessionsKey(payload.UserID), ttl)
}
