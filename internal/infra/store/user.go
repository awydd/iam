package store

import (
	"context"
	"fmt"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/ent/db/user"
)

var _ biz.UserStore = (*UserStore)(nil)

type UserStore struct {
	*baseStore
}

func NewUserStore(client *db.Client) *UserStore {
	return &UserStore{baseStore: newBaseStore(client)}
}

func (s *UserStore) GetSystem(ctx context.Context) (*db.User, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.IsSystemEQ(true),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query system user: %w", err)
	}
	return info, nil
}

func (s *UserStore) InitCreate(ctx context.Context, username, email, passwordHash string, status enum.UserStatus) (*db.User, error) {
	client := s.Client(ctx)

	existing, err := client.User.Query().
		Where(user.IsSystemEQ(true)).
		ForUpdate().
		Only(ctx)
	if err == nil {
		return existing, nil
	}
	if !db.IsNotFound(err) {
		return nil, fmt.Errorf("lock system user: %w", err)
	}

	res, err := client.User.Create().
		SetEmail(email).
		SetUsername(username).
		SetPassword([]byte(passwordHash)).
		SetStatus(status).
		SetIsSystem(true).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create system user: %w", err)
	}
	return res, nil
}
