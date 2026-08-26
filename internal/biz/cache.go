package biz

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionCachePayload struct {
	UserID    int
	UserUUID  uuid.UUID
	ExpiredAt int64
}

type SessionCache interface {
	Set(ctx context.Context, sid string, payload SessionCachePayload, ttl time.Duration) error
	Get(ctx context.Context, sid string) (*SessionCachePayload, error)
	Del(ctx context.Context, sid string) error
	DelAllByUser(ctx context.Context, userID int) error
	Refresh(ctx context.Context, sid string, payload *SessionCachePayload, ttl time.Duration) error
}

type TokenCache interface {
	SetAccess(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error
	ExistsAccess(ctx context.Context, sessionID uuid.UUID) (bool, error)
	DelAccess(ctx context.Context, sessionID uuid.UUID) error
	SetRefresh(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error
	ExistsRefresh(ctx context.Context, sessionID uuid.UUID) (bool, error)
	DelRefresh(ctx context.Context, sessionID uuid.UUID) error
}
