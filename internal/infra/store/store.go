package store

import (
	"context"
	"fmt"

	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/logger"
)

type baseStore struct {
	client *db.Client
}

func newBaseStore(client *db.Client) *baseStore {
	return &baseStore{client: client}
}

func (s *baseStore) Client(ctx context.Context) *db.Client {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Client()
	}
	return s.client
}

func paginate(total, page, perPage int) (offset, limit int, ok bool) {
	if perPage <= 0 {
		perPage = 1
	}
	if page < 1 {
		page = 1
	}
	if total == 0 {
		return 0, perPage, true
	}

	totalPages := (total + perPage - 1) / perPage
	if page > totalPages {
		return 0, 0, false
	}
	return (page - 1) * perPage, perPage, true
}

type txKey struct{}

func txFromContext(ctx context.Context) (*db.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*db.Tx)
	return tx, ok
}

type Transactor struct {
	client *db.Client
}

func NewTransactor(client *db.Client) *Transactor {
	return &Transactor{client: client}
}

func (t *Transactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := txFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := t.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			if rerr := tx.Rollback(); rerr != nil {
				logger.Error("rollback after panic failed: %v (panic: %v)", rerr, p)
			}
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rolling back transaction: %v (original error: %w)", rerr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
