package store

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/ent/db/keypair"
)

var _ biz.KeypairStore = (*KeypairStore)(nil)

type KeypairStore struct {
	*baseStore
}

func NewKeypairStore(client *db.Client) *KeypairStore {
	return &KeypairStore{baseStore: newBaseStore(client)}
}

func (s *KeypairStore) ListVerifiable(ctx context.Context) ([]*db.Keypair, error) {
	list, err := s.Client(ctx).Keypair.Query().
		Where(keypair.StatusIn(enum.KeypairStatusActive, enum.KeypairStatusGrace)).
		Order(db.Desc(keypair.FieldActivatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list verifiable keys: %w", err)
	}
	return list, nil
}

func (s *KeypairStore) GetActiveSigningKey(ctx context.Context) (*db.Keypair, error) {
	kp, err := s.Client(ctx).Keypair.Query().
		Where(keypair.StatusEQ(enum.KeypairStatusActive)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active signing key: %w", err)
	}
	return kp, nil
}

func (s *KeypairStore) GetByKid(ctx context.Context, kid string) (*db.Keypair, error) {
	kp, err := s.Client(ctx).Keypair.Query().
		Where(keypair.KidEQ(kid)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get key pair by kid %s: %w", kid, err)
	}
	return kp, nil
}

func (s *KeypairStore) Create(ctx context.Context, kid string, algorithm enum.KeypairAlgorithm, publicKey, privateKey string) (*db.Keypair, error) {
	kp, err := s.Client(ctx).Keypair.Create().
		SetKid(kid).
		SetAlgorithm(algorithm).
		SetPublicKey(publicKey).
		SetPrivateKey(privateKey).
		SetStatus(enum.KeypairStatusActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create key pair: %w", err)
	}
	return kp, nil
}

func (s *KeypairStore) DeleteRetiredBefore(ctx context.Context, before time.Time) (int, error) {
	affected, err := s.Client(ctx).Keypair.Delete().
		Where(
			keypair.StatusEQ(enum.KeypairStatusRetired),
			keypair.RetireAtLTE(before),
		).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete retired keypairs: %w", err)
	}
	return affected, nil
}
