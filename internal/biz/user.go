package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/consts"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/logger"
	"github.com/awydd/iam/pkg/hashutil"
	"github.com/awydd/iam/pkg/password"
	"github.com/awydd/iam/pkg/random"
	"github.com/google/uuid"
)

type UserStore interface {
	Get(ctx context.Context, id int) (*db.User, error)
	GetByUsername(ctx context.Context, username string) (*db.User, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	GetSystem(ctx context.Context) (*db.User, error)

	UpdateLastLogin(ctx context.Context, id int, at time.Time) error
}

type TokenSigner interface {
	SignAccessToken(ctx context.Context, userID, sessionID uuid.UUID, name string, ttl time.Duration) (string, error)
}

type UserBiz struct {
	store        UserStore
	tokenStore   TokenStore
	tx           Transactor
	signer       TokenSigner
	sessionCache SessionCache
	tokenCache   TokenCache
	loginAttempt LoginAttemptCache
}

func NewUserBiz(
	store UserStore,
	tokenStore TokenStore,
	tx Transactor,
	signer TokenSigner,
	sessionCache SessionCache,
	tokenCache TokenCache,
	loginAttempt LoginAttemptCache,
) *UserBiz {
	return &UserBiz{
		store:        store,
		tokenStore:   tokenStore,
		tx:           tx,
		signer:       signer,
		sessionCache: sessionCache,
		tokenCache:   tokenCache,
		loginAttempt: loginAttempt,
	}
}

type LoginResp struct {
	AccessToken  string
	RefreshToken string
	SessionID    string
}

func (b *UserBiz) Login(ctx context.Context, username, passwordStr, ip, userAgent string) (*LoginResp, error) {
	locked, ttl, err := b.loginAttempt.IsLocked(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("check login lock: %w", err)
	}
	if locked {
		return nil, fmt.Errorf("%w: try again in %d seconds", ErrAccountLocked, int(ttl.Seconds()))
	}

	u, err := b.store.GetByUsername(ctx, username)
	if err != nil {
		if db.IsNotFound(err) {
			b.recordFailure(ctx, username, ip)
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	if u.Status != enum.UserStatusActive {
		return nil, ErrAccountNotActive
	}

	ok, err := password.Validate(passwordStr, string(u.Password))
	if err != nil {
		return nil, fmt.Errorf("validate password: %w", err)
	}
	if !ok {
		b.recordFailure(ctx, username, ip)
		return nil, ErrInvalidCredentials
	}

	_ = b.loginAttempt.ResetFailure(ctx, username)
	if ip != "" {
		_ = b.loginAttempt.ResetFailureByIP(ctx, ip)
	}

	if err := b.store.UpdateLastLogin(ctx, u.ID, time.Now()); err != nil {
		logger.Error("update last_login_at failed: %s", err)
	}

	jwtSessionID := uuid.New()
	jwtCfg := conf.Get().JWT

	sessionID, err := b.startSession(ctx, u.ID, u.UUID)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}

	accessToken, err := b.signer.SignAccessToken(ctx, u.UUID, jwtSessionID, u.Username, jwtCfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshToken, err := b.issueRefreshToken(ctx, u.ID, jwtSessionID, ip, userAgent)
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	if err := b.tokenCache.SetAccess(ctx, jwtSessionID, jwtCfg.AccessTokenTTL); err != nil {
		return nil, fmt.Errorf("cache access token: %w", err)
	}
	if err := b.tokenCache.SetRefresh(ctx, jwtSessionID, jwtCfg.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("cache refresh token: %w", err)
	}

	return &LoginResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}, nil
}

func (b *UserBiz) recordFailure(ctx context.Context, username, ip string) {
	loginSecurityCfg := conf.Get().LoginSecurity
	count, err := b.loginAttempt.IncrFailure(ctx, username, loginSecurityCfg.AttemptWindow)
	if err != nil {
		logger.Error("incr login failure failed: %s", err)
		return
	}
	if count >= int64(loginSecurityCfg.MaxAttempts) {
		if err := b.loginAttempt.Lock(ctx, username, loginSecurityCfg.LockoutDuration); err != nil {
			logger.Error("lock account failed: %s", err)
		}
		_ = b.loginAttempt.ResetFailure(ctx, username)
	}
	if ip != "" {
		if _, err := b.loginAttempt.IncrFailureByIP(ctx, ip, loginSecurityCfg.AttemptWindow); err != nil {
			logger.Error("incr login failure by ip failed: %s", err)
		}
	}
}

func (b *UserBiz) startSession(ctx context.Context, userID int, userUUID uuid.UUID) (string, error) {
	sid, err := random.Alphanumeric(32)
	if err != nil {
		return "", err
	}

	payload := SessionCachePayload{UserID: userID, UserUUID: userUUID}
	if err := b.sessionCache.Set(ctx, sid, payload, consts.SessionTTL); err != nil {
		return "", err
	}
	return sid, nil
}

func (b *UserBiz) issueRefreshToken(ctx context.Context, userID int, sessionID uuid.UUID, ip, userAgent string) (string, error) {
	return b.createRefreshToken(ctx, userID, 0, sessionID, ip, userAgent)
}

func (b *UserBiz) createRefreshToken(ctx context.Context, userID, applicationID int, sessionID uuid.UUID, ip, userAgent string) (string, error) {
	plain, err := random.Alphanumeric(64)
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	identity := hashutil.Sum256([]byte(plain))

	err = b.tx.WithTx(ctx, func(ctx context.Context) error {
		return b.tokenStore.Create(ctx, TokenCreateCommand{
			UserID:        userID,
			ApplicationID: applicationID,
			Jti:           uuid.New(),
			SessionID:     sessionID,
			Type:          enum.TokenTypeRefresh,
			Identity:      identity,
			ExpiresAt:     time.Now().Add(conf.Get().JWT.RefreshTokenTTL),
			IP:            ip,
			UserAgent:     userAgent,
		})
	})
	if err != nil {
		return "", err
	}
	return plain, nil
}
