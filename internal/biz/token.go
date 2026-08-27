package biz

import (
	"context"
	"fmt"
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
	List(ctx context.Context, userID, applicationID int, page, perPage int) ([]*db.Token, int, error)
	ListActiveSessionsByUserID(ctx context.Context, userID int) ([]uuid.UUID, error)
	ListActiveSessionsByApplicationID(ctx context.Context, applicationID int) ([]uuid.UUID, error)

	Get(ctx context.Context, id int) (*db.Token, error)
	GetIfValid(ctx context.Context, identity []byte, tokenType enum.TokenType) (*db.Token, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*db.Token, error)

	Create(ctx context.Context, body TokenCreateCommand) error

	Revoke(ctx context.Context, id int) error
	RevokeBySessionID(ctx context.Context, sessionID uuid.UUID) error
	RevokeByJti(ctx context.Context, jti uuid.UUID) error
	RevokeByUserID(ctx context.Context, userID int) error
	RevokeByApplicationID(ctx context.Context, applicationID int) error

	UpdateLastActiveBySessionID(ctx context.Context, sessionID uuid.UUID, t time.Time) error

	ClearExpireOrRevoked(ctx context.Context) (int, error)
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

func (b *TokenBiz) CleanExpiredOrRevoked(ctx context.Context) (int, error) {
	const maxBatches = 1000
	total := 0
	for range maxBatches {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}

		affected, err := b.store.ClearExpireOrRevoked(ctx)
		if err != nil {
			return total, fmt.Errorf("clear batch: %w", err)
		}
		total += affected

		if affected == 0 {
			break
		}
		logger.Info("token cleanup: removed %d rows this batch, %d total so far", affected, total)
	}
	return total, nil
}
