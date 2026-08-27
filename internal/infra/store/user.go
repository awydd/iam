package store

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/ent/db/user"
	"github.com/google/uuid"
)

var _ biz.UserStore = (*UserStore)(nil)

type UserStore struct {
	*baseStore
}

func NewUserStore(client *db.Client) *UserStore {
	return &UserStore{baseStore: newBaseStore(client)}
}

func (s *UserStore) Get(ctx context.Context, id int) (*db.User, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.IDEQ(id),
			user.DeletedAtIsNil(),
		).
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("get user by id %d: %w", id, err)
	}

	return info, nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.UsernameEQ(username),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user by username %s: %w", username, err)
	}
	return info, nil
}

func (s *UserStore) GetByUUID(ctx context.Context, id uuid.UUID) (*db.User, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.UUIDEQ(id),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user by uuid %s: %w", id, err)
	}
	return info, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.EmailEQ(email),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user by email %s: %w", email, err)
	}
	return info, nil
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

func (s *UserStore) Duplicate(ctx context.Context, username, email string, id ...int) (bool, error) {
	q := s.Client(ctx).User.Query().
		Where(
			user.Or(
				user.UsernameEQ(username),
				user.EmailEQ(email),
			),
		)

	if len(id) > 0 && id[0] > 0 {
		q = q.Where(user.IDNEQ(id[0]))
	}

	exist, err := q.Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("check user duplicate: %w", err)
	}

	return exist, nil
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

func (s *UserStore) Update(ctx context.Context, id int, username, email string, status enum.UserStatus, hashed string) (*db.User, error) {
	builder := s.Client(ctx).User.UpdateOneID(id).
		SetEmail(email).
		SetUsername(username).
		SetStatus(status)

	if hashed != "" {
		builder.SetPassword([]byte(hashed))
	}

	res, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update user %d: %w", id, err)
	}
	return res, nil
}

func (s *UserStore) UpdateLastLogin(ctx context.Context, id int, at time.Time) error {
	if err := s.Client(ctx).User.UpdateOneID(id).
		SetLastLoginAt(at).
		Exec(ctx); err != nil {
		return fmt.Errorf("update last login for user %d: %w", id, err)
	}
	return nil
}

func (s *UserStore) UpdatePassword(ctx context.Context, id int, hashed string) error {
	if err := s.Client(ctx).User.UpdateOneID(id).
		SetPassword([]byte(hashed)).
		Exec(ctx); err != nil {
		return fmt.Errorf("update password for user %d: %w", id, err)
	}
	return nil
}
