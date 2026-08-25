package jwt

import (
	"context"
	"crypto"
	"errors"
)

var ErrKeyNotFound = errors.New("jwt: key not found")

type KeyProvider interface {
	SigningKey(ctx context.Context) (kid string, priv crypto.Signer, err error)
	VerifyKey(ctx context.Context, kid string) (pub crypto.PublicKey, err error)
}

type MemoryKeyProvider struct {
	kid  string
	priv crypto.Signer
}

func NewMemoryKeyProvider(kid string, priv crypto.Signer) *MemoryKeyProvider {
	return &MemoryKeyProvider{kid: kid, priv: priv}
}

func (p *MemoryKeyProvider) SigningKey(_ context.Context) (string, crypto.Signer, error) {
	return p.kid, p.priv, nil
}

func (p *MemoryKeyProvider) VerifyKey(_ context.Context, kid string) (crypto.PublicKey, error) {
	if kid != p.kid {
		return nil, ErrKeyNotFound
	}
	return p.priv.Public(), nil
}
