package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
)

type OAuthCodeCache struct {
	rdb *Redis
}

func NewOAuthCodeCache() *OAuthCodeCache {
	return &OAuthCodeCache{rdb: DB()}
}

var _ biz.OAuthCodeCache = (*OAuthCodeCache)(nil)

func (t *OAuthCodeCache) codeKey(code string) string {
	return consts.RedisOAuthCode + ":" + code
}

func (t *OAuthCodeCache) SetCode(ctx context.Context, code string, payload biz.OAuthCodePayload) error {
	return t.rdb.SetJSON(ctx, t.codeKey(code), payload, time.Minute)
}

func (t *OAuthCodeCache) GetAndDelCode(ctx context.Context, code string) (*biz.OAuthCodePayload, error) {
	key := t.codeKey(code)

	luaScript := `
		local val = redis.call("GET", KEYS[1])
		if val then
			redis.call("DEL", KEYS[1])
		end
		return val
	`

	res, err := t.rdb.Client().Eval(ctx, luaScript, []string{t.rdb.key(key)}).Result()
	if err != nil {
		return nil, fmt.Errorf("eval get-and-del code: %w", err)
	}

	valStr, ok := res.(string)
	if !ok || valStr == "" {
		return nil, fmt.Errorf("code not found or expired")
	}

	var payload biz.OAuthCodePayload
	if err := json.Unmarshal([]byte(valStr), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal oauth payload failed: %w", err)
	}
	return &payload, nil
}
