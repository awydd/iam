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

func (s *KeypairStore) List(ctx context.Context, page, perPage int) ([]*db.Keypair, int, error) {
	q := s.Client(ctx).Keypair.Query().
		Where(keypair.StatusNEQ(enum.KeypairStatusRetired))

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count keypairs: %w", err)
	}
	if total == 0 {
		return []*db.Keypair{}, 0, nil
	}

	offset, limit, ok := paginate(total, page, perPage)
	if !ok {
		return []*db.Keypair{}, total, nil
	}

	list, err := q.
		Order(db.Desc(keypair.FieldActivatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list keypairs: %w", err)
	}
	return list, total, nil
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

func (s *KeypairStore) Retire(ctx context.Context, kid string) error {
	kp, err := s.Client(ctx).Keypair.Query().
		Where(keypair.KidEQ(kid)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("get key pair %s: %w", kid, err)
	}

	if err := s.Client(ctx).Keypair.UpdateOneID(kp.ID).
		SetStatus(enum.KeypairStatusRetired).
		SetRetireAt(time.Now()).
		Exec(ctx); err != nil {
		return fmt.Errorf("retire key pair %s: %w", kid, err)
	}
	return nil
}

func (s *KeypairStore) Downgrade(ctx context.Context, kid string) error {
	kp, err := s.Client(ctx).Keypair.Query().
		Where(keypair.KidEQ(kid)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("get key pair %s: %w", kid, err)
	}

	if err := s.Client(ctx).Keypair.UpdateOneID(kp.ID).
		SetStatus(enum.KeypairStatusGrace).
		Exec(ctx); err != nil {
		return fmt.Errorf("downgrade key pair %s: %w", kid, err)
	}
	return nil
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
