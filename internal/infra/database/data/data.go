package data

import (
	"context"
	"fmt"

	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/store"
)

type Result struct {
	KeypairCreated bool
	KeypairKid     string

	UserCreated bool
	Username    string
	Email       string
	Password    string
}

type Options struct {
	Username string
	Email    string
}

func Init(ctx context.Context, client *db.Client, opts Options) (*Result, error) {
	result := &Result{}

	kpStore := store.NewKeypairStore(client)
	if err := initKeypair(ctx, kpStore, result); err != nil {
		return result, fmt.Errorf("init keypair: %w", err)
	}

	userStore := store.NewUserStore(client)
	tx := store.NewTransactor(client)

	_, err := initUser(ctx, userStore, tx, opts, result)
	if err != nil {
		return result, fmt.Errorf("init user: %w", err)
	}

	return result, nil
}
