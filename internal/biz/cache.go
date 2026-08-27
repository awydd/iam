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

type LoginAttemptCache interface {
	IncrFailure(ctx context.Context, username string, window time.Duration) (int64, error)
	ResetFailure(ctx context.Context, username string) error
	Lock(ctx context.Context, username string, duration time.Duration) error
	IsLocked(ctx context.Context, username string) (bool, time.Duration, error)
	IncrFailureByIP(ctx context.Context, ip string, window time.Duration) (int64, error)
	ResetFailureByIP(ctx context.Context, ip string) error
}

const lastActiveThrottleWindow = 5 * time.Minute

type LastActiveCache interface {
	ShouldUpdate(ctx context.Context, sessionID uuid.UUID, window time.Duration) (bool, error)
}

type OAuthCodePayload struct {
	UserID              int
	UserUUID            uuid.UUID
	SessionID           uuid.UUID
	ClientID            string
	RedirectURI         string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
}

type OAuthCodeCache interface {
	SetCode(ctx context.Context, code string, payload OAuthCodePayload) error
	GetAndDelCode(ctx context.Context, code string) (*OAuthCodePayload, error)
}
