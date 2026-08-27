package jwt

import (
	"crypto"
	"fmt"

	"github.com/lestrrat-go/jwx/v4/jwk"
)

type JWK = jwk.Key
type JWKSet = jwk.Set

func PublicKeyToJWK(kid string, pub crypto.PublicKey) (jwk.Key, error) {
	key, err := jwk.Import[jwk.Key](pub)
	if err != nil {
		return nil, fmt.Errorf("jwt: failed to create jwk from public key: %w", err)
	}

	if err := key.Set(jwk.KeyIDKey, kid); err != nil {
		return nil, fmt.Errorf("jwt: failed to set kid: %w", err)
	}
	if err := key.Set(jwk.KeyUsageKey, "sig"); err != nil {
		return nil, fmt.Errorf("jwt: failed to set use: %w", err)
	}

	return key, nil
}
