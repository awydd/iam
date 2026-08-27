package biz

import (
	"context"
	"errors"
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
	tokenCache      TokenCache
}

func NewTokenBiz(store TokenStore, lastActiveCache LastActiveCache, tokenCache TokenCache) *TokenBiz {
	return &TokenBiz{store: store, lastActiveCache: lastActiveCache, tokenCache: tokenCache}
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

type TokenListItemResp struct {
	ID              int            `json:"id"`
	UserID          int            `json:"user_id"`
	Username        string         `json:"username"`
	ApplicationID   int            `json:"application_id"`
	ApplicationName string         `json:"application_name"`
	SessionID       uuid.UUID      `json:"session_id"`
	Type            enum.TokenType `json:"type"`
	IP              string         `json:"ip"`
	UserAgent       string         `json:"user_agent"`
	ExpiresAt       time.Time      `json:"expires_at"`
}

func (b *TokenBiz) List(ctx context.Context, userID, applicationID, page, perPage int) ([]TokenListItemResp, int, error) {
	list, total, err := b.store.List(ctx, userID, applicationID, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("list tokens: %w", err)
	}

	views := make([]TokenListItemResp, 0, len(list))
	for _, t := range list {
		view := TokenListItemResp{
			ID:            t.ID,
			UserID:        t.UserID,
			ApplicationID: t.ApplicationID,
			SessionID:     t.SessionID,
			Type:          t.Type,
			IP:            t.IP,
			UserAgent:     t.UserAgent,
			ExpiresAt:     t.ExpiresAt,
		}
		if t.Edges.User != nil {
			view.Username = t.Edges.User.Username
		}
		if t.Edges.Application != nil {
			view.ApplicationName = t.Edges.Application.Name
		}
		views = append(views, view)
	}
	return views, total, nil
}

type TokenInfoResp struct {
	ID              int            `json:"id"`
	UserID          int            `json:"user_id"`
	Username        string         `json:"username"`
	ApplicationID   int            `json:"application_id"`
	ApplicationName string         `json:"application_name"`
	SessionID       uuid.UUID      `json:"session_id"`
	Type            enum.TokenType `json:"type"`
	IP              string         `json:"ip"`
	UserAgent       string         `json:"user_agent"`
	ExpiresAt       time.Time      `json:"expires_at"`
}

func (b *TokenBiz) Info(ctx context.Context, id int) (*TokenInfoResp, error) {
	t, err := b.store.Get(ctx, id)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("token not found")
		}
		return nil, fmt.Errorf("get token: %w", err)
	}

	view := &TokenInfoResp{
		ID:            t.ID,
		UserID:        t.UserID,
		ApplicationID: t.ApplicationID,
		SessionID:     t.SessionID,
		Type:          t.Type,
		IP:            t.IP,
		UserAgent:     t.UserAgent,
		ExpiresAt:     t.ExpiresAt,
	}
	if t.Edges.User != nil {
		view.Username = t.Edges.User.Username
	}
	if t.Edges.Application != nil {
		view.ApplicationName = t.Edges.Application.Name
	}

	return view, nil
}

func (b *TokenBiz) Revoke(ctx context.Context, id int) error {
	tok, err := b.store.Get(ctx, id)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("token not found")
		}
		return fmt.Errorf("get token: %w", err)
	}

	if tok.RevokedAt != nil {
		return errors.New("token already revoked")
	}

	if err := b.store.Revoke(ctx, id); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	if b.tokenCache != nil {
		if err := b.tokenCache.DelAccess(ctx, tok.SessionID); err != nil {
			logger.Error("del access cache failed: session_id=%s err=%v", tok.SessionID, err)
		}
		if err := b.tokenCache.DelRefresh(ctx, tok.SessionID); err != nil {
			logger.Error("del refresh cache failed: session_id=%s err=%v", tok.SessionID, err)
		}
	}

	return nil
}
