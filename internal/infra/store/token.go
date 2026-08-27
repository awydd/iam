package store

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/enum"
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

func (s *TokenStore) List(ctx context.Context, userID, applicationID int, page, perPage int) ([]*db.Token, int, error) {
	q := s.Client(ctx).Token.Query().
		Where(
			token.ExpiresAtGT(time.Now()),
			token.RevokedAtIsNil(),
		).
		WithUser().
		WithApplication()

	if userID > 0 {
		q = q.Where(token.UserIDEQ(userID))
	}
	if applicationID > 0 {
		q = q.Where(token.ApplicationIDEQ(applicationID))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count tokens: %w", err)
	}

	if total == 0 {
		return []*db.Token{}, 0, nil
	}

	offset, limit, ok := paginate(total, page, perPage)
	if !ok {
		return []*db.Token{}, 0, nil
	}

	list, err := q.
		Order(db.Desc(token.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list tokens: %w", err)
	}

	return list, total, nil
}

func (s *TokenStore) ListActiveSessionsByUserID(ctx context.Context, userID int) ([]uuid.UUID, error) {
	var sessions []uuid.UUID
	err := s.Client(ctx).Token.Query().
		Where(
			token.UserIDEQ(userID),
			token.RevokedAtIsNil(),
			token.ExpiresAtGT(time.Now()),
		).
		Select(token.FieldSessionID).
		Scan(ctx, &sessions)

	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *TokenStore) ListActiveSessionsByApplicationID(ctx context.Context, applicationID int) ([]uuid.UUID, error) {
	var sessions []uuid.UUID
	err := s.Client(ctx).Token.Query().
		Where(
			token.ApplicationIDEQ(applicationID),
			token.RevokedAtIsNil(),
			token.ExpiresAtGT(time.Now()),
		).
		Select(token.FieldSessionID).
		Scan(ctx, &sessions)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *TokenStore) Get(ctx context.Context, id int) (*db.Token, error) {
	info, err := s.Client(ctx).Token.Query().
		Where(
			token.IDEQ(id),
		).
		WithUser().
		WithApplication().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *TokenStore) GetIfValid(ctx context.Context, identity []byte, tokenType enum.TokenType) (*db.Token, error) {
	info, err := s.Client(ctx).Token.Query().
		Where(
			token.IdentityEQ(identity),
			token.TypeEQ(tokenType),
			token.ExpiresAtGT(time.Now()),
			token.RevokedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *TokenStore) GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*db.Token, error) {
	info, err := s.Client(ctx).Token.Query().
		Where(
			token.SessionIDEQ(sessionID),
			token.ExpiresAtGT(time.Now()),
			token.RevokedAtIsNil(),
		).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return info, nil
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

func (s *TokenStore) Revoke(ctx context.Context, id int) error {
	_, err := s.Client(ctx).Token.UpdateOneID(id).
		Where(
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

func (s *TokenStore) RevokeByJti(ctx context.Context, jti uuid.UUID) error {
	_, err := s.Client(ctx).Token.Update().
		Where(
			token.JtiEQ(jti),
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

func (s *TokenStore) RevokeByUserID(ctx context.Context, userID int) error {
	_, err := s.Client(ctx).Token.Update().
		Where(
			token.UserIDEQ(userID),
			token.RevokedAtIsNil(),
			token.ExpiresAtGT(time.Now()),
		).
		SetRevokedAt(time.Now()).
		Save(ctx)
	return err
}

func (s *TokenStore) RevokeByApplicationID(ctx context.Context, applicationID int) error {
	_, err := s.Client(ctx).Token.Update().
		Where(
			token.ApplicationIDEQ(applicationID),
			token.RevokedAtIsNil(),
			token.ExpiresAtGT(time.Now()),
		).
		SetRevokedAt(time.Now()).
		Save(ctx)
	return err
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
