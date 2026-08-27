package data

import (
	"context"
	"errors"
	"fmt"

	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/store"
	"github.com/awydd/iam/pkg/password"
	"github.com/awydd/iam/pkg/random"
)

func initUser(ctx context.Context, userStore *store.UserStore, tx *store.Transactor, opts Options, result *Result) (*db.User, error) {
	if u, err := userStore.GetSystem(ctx); err == nil {
		result.Username = u.Username
		result.Email = u.Email
		return u, nil
	} else if !db.IsNotFound(err) {
		return nil, err
	}

	if opts.Username == "" || opts.Email == "" {
		return nil, errors.New("username and email are required to bootstrap the system user")
	}

	plain, err := random.Unambiguous(12)
	if err != nil {
		return nil, fmt.Errorf("generate password: %w", err)
	}
	hashed, err := password.Hash(plain)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var u *db.User
	err = tx.WithTx(ctx, func(ctx context.Context) error {
		var createErr error
		u, createErr = userStore.InitCreate(ctx, opts.Username, opts.Email, hashed, enum.UserStatusActive)
		return createErr
	})
	if err != nil {
		return nil, err
	}

	result.UserCreated = true
	result.Username = u.Username
	result.Email = u.Email
	result.Password = plain
	return u, nil
}
