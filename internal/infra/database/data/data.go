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
}

func Init(ctx context.Context, client *db.Client) (*Result, error) {
	result := &Result{}

	kpStore := store.NewKeypairStore(client)
	if err := initKeypair(ctx, kpStore, result); err != nil {
		return result, fmt.Errorf("init keypair: %w", err)
	}

	return result, nil
}
