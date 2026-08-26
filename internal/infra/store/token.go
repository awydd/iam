package store

import (
	"context"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/infra/ent/db"
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
