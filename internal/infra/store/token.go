package store

import (
	"context"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/ent/db/token"
	"github.com/google/uuid"
)

var _ biz.TokenStore = (*TokenStore)(nil)

type TokenStore struct {
	*baseStore
}

func NewTokenStore(client *db.Client) *TokenStore {
	return &TokenStore{baseStore: newBaseStore(client)}
}

func (s *TokenStore) Create(ctx context.Context, body biz.TokenCreateCommand) error {
	_, err := s.Client(ctx).Token.Create().
		SetUserID(body.UserID).
		SetJti(body.Jti).
		SetApplicationID(body.ApplicationID).
		SetSessionID(body.SessionID).
		SetType(body.Type).
		SetIdentity(body.Identity).
		SetIP(body.IP).
		SetUserAgent(body.UserAgent).
		SetExpiresAt(body.ExpiresAt).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *TokenStore) RevokeBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	_, err := s.Client(ctx).Token.Update().
		Where(
			token.SessionIDEQ(sessionID),
			token.RevokedAtIsNil(),
			token.ExpiresAtGT(time.Now()),
		).
		SetRevokedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *TokenStore) UpdateLastActiveBySessionID(ctx context.Context, sessionID uuid.UUID, t time.Time) error {
	_, err := s.Client(ctx).Token.Update().
		Where(
			token.SessionIDEQ(sessionID),
			token.RevokedAtIsNil(),
			token.ExpiresAtGT(time.Now()),
		).
		SetLastActiveAt(t).
		Save(ctx)
	return err
}
