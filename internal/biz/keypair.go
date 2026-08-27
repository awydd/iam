package biz

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/jwt"
	"github.com/awydd/iam/internal/logger"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v4/jwk"
)

var _ jwt.KeyProvider = (*KeypairBiz)(nil)

type KeypairStore interface {
	List(ctx context.Context, page, perPage int) ([]*db.Keypair, int, error)
	ListVerifiable(ctx context.Context) ([]*db.Keypair, error)

	GetActiveSigningKey(ctx context.Context) (*db.Keypair, error)
	GetByKid(ctx context.Context, kid string) (*db.Keypair, error)

	Create(ctx context.Context, kid string, algorithm enum.KeypairAlgorithm, publicKey, privateKey string) (*db.Keypair, error)

	Retire(ctx context.Context, kid string) error
	Downgrade(ctx context.Context, kid string) error

	DeleteRetiredBefore(ctx context.Context, before time.Time) (int, error)
}

type keypairCache struct {
	mu sync.RWMutex

	verifyKeys map[string]crypto.PublicKey
	loadedAt   time.Time

	signingKid  string
	signingPriv crypto.Signer
	signingAt   time.Time
}

type KeypairBiz struct {
	store KeypairStore
	tx    Transactor
	cache keypairCache

	verifyCacheTTL  time.Duration
	signingCacheTTL time.Duration
}

func durationOrDefault(v, def time.Duration) time.Duration {
	if v <= 0 {
		return def
	}
	return v
}

func NewKeypairBiz(store KeypairStore, tx Transactor) *KeypairBiz {
	cfg := conf.Get().Keypair
	return &KeypairBiz{
		store:           store,
		tx:              tx,
		verifyCacheTTL:  durationOrDefault(cfg.VerifyCacheTTL, conf.DefaultKeypairVerifyCacheTTL),
		signingCacheTTL: durationOrDefault(cfg.SigningCacheTTL, conf.DefaultKeypairSigningCacheTTL),
	}
}

func (b *KeypairBiz) SigningKey(ctx context.Context) (string, crypto.Signer, error) {
	b.cache.mu.RLock()
	if b.cache.signingPriv != nil && time.Since(b.cache.signingAt) < b.signingCacheTTL {
		kid, priv := b.cache.signingKid, b.cache.signingPriv
		b.cache.mu.RUnlock()
		return kid, priv, nil
	}
	b.cache.mu.RUnlock()

	kp, err := b.store.GetActiveSigningKey(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			logger.Warn("no active signing key found")
			return "", nil, errors.New("no active signing key available")
		}
		logger.Error("get active signing key failed: %v", err)
		return "", nil, fmt.Errorf("get active signing key: %w", err)
	}

	priv, err := parsePrivateKeyPEM(kp.PrivateKey)
	if err != nil {
		logger.Error("parse private key failed: kid=%s err=%v", kp.Kid, err)
		return "", nil, fmt.Errorf("parse private key (kid=%s): %w", kp.Kid, err)
	}

	b.cache.mu.Lock()
	b.cache.signingKid = kp.Kid
	b.cache.signingPriv = priv
	b.cache.signingAt = time.Now()
	b.cache.mu.Unlock()

	return kp.Kid, priv, nil
}

func parsePrivateKeyPEM(pemStr string) (crypto.Signer, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, errors.New("private key does not implement crypto.Signer")
	}
	return signer, nil
}

func (b *KeypairBiz) VerifyKey(ctx context.Context, kid string) (crypto.PublicKey, error) {
	if pub, ok := b.lookupCache(kid); ok {
		return pub, nil
	}
	if err := b.refreshVerifyCache(ctx); err != nil {
		return nil, err
	}
	if pub, ok := b.lookupCache(kid); ok {
		return pub, nil
	}

	kp, err := b.store.GetByKid(ctx, kid)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, jwt.ErrKeyNotFound
		}
		logger.Error("get key by kid failed: kid=%s err=%v", kid, err)
		return nil, fmt.Errorf("get key by kid: %w", err)
	}
	if kp.Status == enum.KeypairStatusRetired {
		return nil, jwt.ErrKeyNotFound
	}

	pub, err := parsePublicKeyPEM(kp.PublicKey)
	if err != nil {
		logger.Error("parse public key failed: kid=%s err=%v", kid, err)
		return nil, fmt.Errorf("parse public key (kid=%s): %w", kid, err)
	}
	return pub, nil
}

func (b *KeypairBiz) lookupCache(kid string) (crypto.PublicKey, bool) {
	b.cache.mu.RLock()
	defer b.cache.mu.RUnlock()

	if b.cache.verifyKeys == nil || time.Since(b.cache.loadedAt) >= b.verifyCacheTTL {
		return nil, false
	}
	pub, ok := b.cache.verifyKeys[kid]
	return pub, ok
}

// 检查验证缓存是否过期，如果过期了则自动触发刷新
func (b *KeypairBiz) refreshVerifyCacheIfExpired(ctx context.Context) error {
	b.cache.mu.RLock()
	expired := b.cache.verifyKeys == nil || time.Since(b.cache.loadedAt) >= b.verifyCacheTTL
	b.cache.mu.RUnlock()

	if expired {
		return b.refreshVerifyCache(ctx)
	}
	return nil
}

func (b *KeypairBiz) refreshVerifyCache(ctx context.Context) error {
	keys, err := b.store.ListVerifiable(ctx)
	if err != nil {
		logger.Error("list verifiable keys failed: %v", err)
		return fmt.Errorf("list verifiable keys: %w", err)
	}

	fresh := make(map[string]crypto.PublicKey, len(keys))
	for _, kp := range keys {
		pub, err := parsePublicKeyPEM(kp.PublicKey)
		if err != nil {
			logger.Error("parse public key failed: kid=%s err=%v", kp.Kid, err)
			return fmt.Errorf("parse public key (kid=%s): %w", kp.Kid, err)
		}
		fresh[kp.Kid] = pub
	}

	b.cache.mu.Lock()
	b.cache.verifyKeys = fresh
	b.cache.loadedAt = time.Now()
	b.cache.mu.Unlock()
	return nil
}

func parsePublicKeyPEM(pemStr string) (crypto.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKIX public key: %w", err)
	}
	switch pub.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pub)
	}
}

// 获取 JSON Web Key Set
func (b *KeypairBiz) JWKS(ctx context.Context) (jwt.JWKSet, error) {
	if err := b.refreshVerifyCacheIfExpired(ctx); err != nil {
		return nil, err
	}

	b.cache.mu.RLock()
	defer b.cache.mu.RUnlock()

	set := jwk.NewSet()
	for kid, pub := range b.cache.verifyKeys {
		jwkObj, err := jwt.PublicKeyToJWK(kid, pub)
		if err != nil {
			logger.Error("convert to jwk failed: kid=%s err=%v", kid, err)
			return nil, fmt.Errorf("convert to jwk (kid=%s): %w", kid, err)
		}

		if err := set.AddKey(jwkObj); err != nil {
			logger.Error("add jwk to set failed: kid=%s err=%v", kid, err)
			return nil, fmt.Errorf("add jwk to set (kid=%s): %w", kid, err)
		}
	}
	return set, nil
}

const keypairRetiredRetention = 30 * 24 * time.Hour

func (b *KeypairBiz) CleanRetired(ctx context.Context) (int, error) {
	cutoff := time.Now().Add(-keypairRetiredRetention)
	affected, err := b.store.DeleteRetiredBefore(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("clean retired keypairs: %w", err)
	}
	return affected, nil
}

type KeypairListItemResp struct {
	Kid         string                `json:"kid"`
	Algorithm   enum.KeypairAlgorithm `json:"algorithm"`
	Status      enum.KeypairStatus    `json:"status"`
	ActivatedAt time.Time             `json:"activated_at"`
	RetireAt    *time.Time            `json:"retire_at,omitempty"`
}

func (b *KeypairBiz) List(ctx context.Context, page, perPage int) ([]KeypairListItemResp, int, error) {
	list, total, err := b.store.List(ctx, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("list keypairs: %w", err)
	}

	views := make([]KeypairListItemResp, 0, len(list))
	for _, kp := range list {
		views = append(views, KeypairListItemResp{
			Kid:         kp.Kid,
			Algorithm:   kp.Algorithm,
			Status:      kp.Status,
			ActivatedAt: kp.ActivatedAt,
			RetireAt:    kp.RetireAt,
		})
	}
	return views, total, nil
}

type KeypairRotateResp struct {
	Kid string `json:"kid"`
}

func (b *KeypairBiz) Rotate(ctx context.Context) (*KeypairRotateResp, error) {
	kid, publicPEM, privatePEM, err := generateRSAKeypairPEM()
	if err != nil {
		return nil, fmt.Errorf("generate rsa keypair: %w", err)
	}

	var newKid string
	err = b.tx.WithTx(ctx, func(ctx context.Context) error {
		oldActive, err := b.store.GetActiveSigningKey(ctx)
		if err != nil && !db.IsNotFound(err) {
			return fmt.Errorf("get active signing key: %w", err)
		}

		kp, err := b.store.Create(ctx, kid, enum.KeypairAlgoRS256, publicPEM, privatePEM)
		if err != nil {
			return fmt.Errorf("create keypair: %w", err)
		}
		newKid = kp.Kid

		if oldActive != nil {
			if err := b.store.Downgrade(ctx, oldActive.Kid); err != nil {
				return fmt.Errorf("downgrade old active key: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	b.invalidateCache()
	return &KeypairRotateResp{Kid: newKid}, nil
}

func (b *KeypairBiz) Downgrade(ctx context.Context, kid string) error {
	kp, err := b.store.GetByKid(ctx, kid)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("keypair not found")
		}
		return fmt.Errorf("get keypair: %w", err)
	}

	if kp.Status == enum.KeypairStatusActive {
		return errors.New("cannot downgrade the active signing key directly; use Rotate to replace it first")
	}
	if kp.Status == enum.KeypairStatusGrace {
		return errors.New("keypair is already in grace status")
	}

	if err := b.store.Downgrade(ctx, kid); err != nil {
		return fmt.Errorf("downgrade keypair: %w", err)
	}

	b.invalidateCache()
	return nil
}

func (b *KeypairBiz) Retire(ctx context.Context, kid string) error {
	kp, err := b.store.GetByKid(ctx, kid)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("keypair not found")
		}
		return fmt.Errorf("get keypair: %w", err)
	}

	switch kp.Status {
	case enum.KeypairStatusActive:
		return errors.New("cannot retire the active signing key; downgrade it to grace first")
	case enum.KeypairStatusRetired:
		return errors.New("keypair is already retired")
	}

	if err := b.store.Retire(ctx, kid); err != nil {
		return fmt.Errorf("retire keypair: %w", err)
	}

	b.invalidateCache()
	return nil
}

func generateRSAKeypairPEM() (kid, publicPEM, privatePEM string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", fmt.Errorf("generate rsa key: %w", err)
	}

	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", "", fmt.Errorf("marshal public key: %w", err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
	privPEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}))

	return uuid.NewString(), pubPEM, privPEM, nil
}

func (b *KeypairBiz) invalidateCache() {
	b.cache.mu.Lock()
	defer b.cache.mu.Unlock()
	b.cache.verifyKeys = nil
	b.cache.loadedAt = time.Time{}
	b.cache.signingKid = ""
	b.cache.signingPriv = nil
	b.cache.signingAt = time.Time{}
}
