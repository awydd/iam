package jwt

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	jwtlib "github.com/lestrrat-go/jwx/v4/jwt"
)

type Claims struct {
	jwtlib.Token

	UserUUID  uuid.UUID
	SessionID uuid.UUID
	Name      string
}

// 获取 Token 的到期时间
func (c *Claims) ExpiresAt() time.Time {
	exp, _ := c.Token.Expiration()
	return exp
}

func ParseClaims(token jwtlib.Token) (*Claims, error) {
	uidStr, err := jwtlib.Get[string](token, "uid")
	if err != nil {
		return nil, fmt.Errorf("jwt: missing or invalid 'uid' claim: %w", err)
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return nil, fmt.Errorf("jwt: invalid 'uid' format: %w", err)
	}

	sidStr, err := jwtlib.Get[string](token, "sid")
	if err != nil {
		return nil, fmt.Errorf("jwt: missing or invalid 'sid' claim: %w", err)
	}
	sid, err := uuid.Parse(sidStr)
	if err != nil {
		return nil, fmt.Errorf("jwt: invalid 'sid' format: %w", err)
	}

	name, _ := jwtlib.Get[string](token, "name")

	return &Claims{
		Token:     token,
		UserUUID:  uid,
		SessionID: sid,
		Name:      name,
	}, nil
}
