package data

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/store"
	"github.com/google/uuid"
)

func initKeypair(ctx context.Context, kpStore *store.KeypairStore, result *Result) error {
	if _, err := kpStore.GetActiveSigningKey(ctx); err == nil {
		return nil
	} else if !db.IsNotFound(err) {
		return err
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate rsa key: %w", err)
	}

	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
	privPEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}))

	kid := uuid.NewString()
	if _, err := kpStore.Create(ctx, kid, enum.KeypairAlgoRS256, pubPEM, privPEM); err != nil {
		return err
	}

	result.KeypairCreated = true
	result.KeypairKid = kid
	return nil
}
