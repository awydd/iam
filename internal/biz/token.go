package biz

import (
	"context"
	"time"

	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/logger"
	"github.com/google/uuid"
)

type TokenCreateCommand struct {
	UserID        int
	ApplicationID int
	Jti           uuid.UUID
	SessionID     uuid.UUID
	Type          enum.TokenType
	Identity      []byte
	ExpiresAt     time.Time
	IP            string
	UserAgent     string
}

type TokenStore interface {
	Get(ctx context.Context, id int) (*db.Token, error)
	GetIfValid(ctx context.Context, identity []byte, tokenType enum.TokenType) (*db.Token, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*db.Token, error)

	Create(ctx context.Context, body TokenCreateCommand) error
	RevokeBySessionID(ctx context.Context, sessionID uuid.UUID) error
	UpdateLastActiveBySessionID(ctx context.Context, sessionID uuid.UUID, t time.Time) error
}

type TokenBiz struct {
	store           TokenStore
	lastActiveCache LastActiveCache
}

func NewTokenBiz(store TokenStore, lastActiveCache LastActiveCache) *TokenBiz {
	return &TokenBiz{store: store, lastActiveCache: lastActiveCache}
}

func (b *TokenBiz) Touch(ctx context.Context, sessionID uuid.UUID) {
	// cmd
	if b.lastActiveCache == nil {
		return
	}

	should, err := b.lastActiveCache.ShouldUpdate(ctx, sessionID, lastActiveThrottleWindow)
	if err != nil {
		logger.Error("last active throttle check failed, session=%s err=%v", sessionID, err)
		return
	}
	if !should {
		return
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := b.store.UpdateLastActiveBySessionID(bgCtx, sessionID, time.Now()); err != nil {
			logger.Error("update last active failed, session=%s err=%v", sessionID, err)
		}
	}()
}
