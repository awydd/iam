package biz

import (
	"context"
	"time"

	"github.com/awydd/iam/internal/enum"
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
	Create(ctx context.Context, body TokenCreateCommand) error
}

type TokenBiz struct {
	store TokenStore
}

func NewTokenBiz(store TokenStore) *TokenBiz {
	return &TokenBiz{store: store}
}
