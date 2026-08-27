package biz

import (
	"context"
	"errors"
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

	Duplicate(ctx context.Context, username, email string, id ...int) (bool, error)

	Update(ctx context.Context, id int, username, email string, status enum.UserStatus, hashed string) (*db.User, error)
	UpdateLastLogin(ctx context.Context, id int, at time.Time) error
	UpdatePassword(ctx context.Context, id int, hashed string) error
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

func (b *UserBiz) Logout(ctx context.Context, sessionID uuid.UUID, sid string) error {
	if err := b.tokenStore.RevokeBySessionID(ctx, sessionID); err != nil {
		return fmt.Errorf("revoke tokens: %w", err)
	}
	if err := b.invalidateSession(ctx, sessionID); err != nil {
		logger.Error("invalidate session cache failed: %s", err)
	}
	if sid != "" {
		if err := b.sessionCache.Del(ctx, sid); err != nil {
			logger.Error("clear session failed: %s", err)
		}
	}
	return nil
}

func (b *UserBiz) invalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	var errs []error
	if err := b.tokenCache.DelAccess(ctx, sessionID); err != nil {
		errs = append(errs, fmt.Errorf("del access cache: %w", err))
	}
	if err := b.tokenCache.DelRefresh(ctx, sessionID); err != nil {
		errs = append(errs, fmt.Errorf("del refresh cache: %w", err))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

type UserMeResp struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	UUID     uuid.UUID `json:"uuid"`
	IsSystem bool      `json:"is_system"`
}

func (b *UserBiz) Me(ctx context.Context, userUUID uuid.UUID) (*UserMeResp, error) {
	u, err := b.store.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get user by uuid: %w", err)
	}

	return &UserMeResp{
		Username: u.Username,
		Email:    u.Email,
		UUID:     u.UUID,
		IsSystem: u.IsSystem,
	}, nil
}

func (b *UserBiz) Refresh(ctx context.Context, refreshToken, ip, userAgent string) (*LoginResp, error) {
	identity := hashutil.Sum256([]byte(refreshToken))

	tok, err := b.tokenStore.GetIfValid(ctx, identity, enum.TokenTypeRefresh)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	online, err := b.tokenCache.ExistsRefresh(ctx, tok.SessionID)
	if err != nil {
		return nil, fmt.Errorf("check refresh token online: %w", err)
	}
	if !online {
		_ = b.tokenStore.RevokeBySessionID(ctx, tok.SessionID)
		return nil, ErrRefreshTokenExpired
	}

	u, err := b.store.Get(ctx, tok.UserID)
	if err != nil {
		if db.IsNotFound(err) {
			_ = b.invalidateSession(ctx, tok.SessionID)
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	if u.Status != enum.UserStatusActive {
		_ = b.invalidateSession(ctx, tok.SessionID)
		return nil, ErrAccountNotActive
	}

	if err := b.tokenStore.RevokeBySessionID(ctx, tok.SessionID); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}
	_ = b.invalidateSession(ctx, tok.SessionID)

	newJwtSessionID := uuid.New()
	jwtCfg := conf.Get().JWT

	accessToken, err := b.signer.SignAccessToken(ctx, u.UUID, newJwtSessionID, u.Username, jwtCfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	var newRefreshToken string
	if tok.ApplicationID == 0 {
		newRefreshToken, err = b.issueRefreshToken(ctx, u.ID, newJwtSessionID, ip, userAgent)
	}

	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	if err := b.tokenCache.SetAccess(ctx, newJwtSessionID, jwtCfg.AccessTokenTTL); err != nil {
		return nil, fmt.Errorf("cache access token: %w", err)
	}
	if err := b.tokenCache.SetRefresh(ctx, newJwtSessionID, jwtCfg.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("cache refresh token: %w", err)
	}

	return &LoginResp{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}

func (b *UserBiz) GetByUUID(ctx context.Context, userUUID uuid.UUID) (*db.User, error) {
	u, err := b.store.GetByUUID(ctx, userUUID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		logger.Error("get user by uuid failed: uuid=%s err=%v", userUUID, err)
		return nil, fmt.Errorf("get user by uuid: %w", err)
	}
	return u, nil
}

func (b *UserBiz) Password(ctx context.Context, id int, oldPassword, newPassword string) error {
	info, err := b.store.Get(ctx, id)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("user not found")
		}
		return fmt.Errorf("get user: %w", err)
	}

	if info.IsSystem {
		return errors.New("system user password cannot be changed via this endpoint; use the CLI tool")
	}

	ok, err := password.Validate(oldPassword, string(info.Password))
	if err != nil {
		return fmt.Errorf("validate password: %w", err)
	}
	if !ok {
		return ErrInvalidCredentials
	}

	hashed, err := password.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := b.store.UpdatePassword(ctx, id, hashed); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	_ = b.InvalidateAllSessions(ctx, id)
	return nil
}

func (b *UserBiz) InvalidateAllSessions(ctx context.Context, userID int) error {
	sys, err := b.store.GetSystem(ctx)
	if err == nil && sys != nil && sys.ID == userID {
		return errors.New("cannot revoke sessions for system user")
	}

	var errs []error

	sessionIDs, err := b.tokenStore.ListActiveSessionsByUserID(ctx, userID)
	if err != nil {
		errs = append(errs, fmt.Errorf("list active sessions: %w", err))
	}

	if err := b.tokenStore.RevokeByUserID(ctx, userID); err != nil {
		errs = append(errs, fmt.Errorf("revoke tokens: %w", err))
	}

	for _, sid := range sessionIDs {
		if err := b.invalidateSession(ctx, sid); err != nil {
			errs = append(errs, err)
		}
	}

	if err := b.sessionCache.DelAllByUser(ctx, userID); err != nil {
		errs = append(errs, fmt.Errorf("clear sessions: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

type UserListSessionsItemResp struct {
	SessionID     uuid.UUID  `json:"session_id"`
	ApplicationID int        `json:"application_id"`
	IP            string     `json:"ip"`
	UserAgent     string     `json:"user_agent"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
	LastActiveAt  *time.Time `json:"last_active_at"`
	IsCurrent     bool       `json:"is_current"`
}

func (b *UserBiz) ListSessions(ctx context.Context, currentSessionID uuid.UUID, userID, page, perPage int) ([]UserListSessionsItemResp, int, error) {
	list, total, err := b.tokenStore.List(ctx, userID, 0, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
	}

	views := make([]UserListSessionsItemResp, 0, len(list))
	for _, t := range list {
		views = append(views, UserListSessionsItemResp{
			SessionID:     t.SessionID,
			ApplicationID: t.ApplicationID,
			IP:            t.IP,
			UserAgent:     t.UserAgent,
			CreatedAt:     t.CreatedAt,
			ExpiresAt:     t.ExpiresAt,
			LastActiveAt:  t.LastActiveAt,
			IsCurrent:     t.SessionID == currentSessionID,
		})
	}
	return views, total, nil
}

func (b *UserBiz) RevokeSession(ctx context.Context, userID int, sessionID uuid.UUID) error {
	tok, err := b.tokenStore.GetBySessionID(ctx, sessionID)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("session not found")
		}
		return fmt.Errorf("get session: %w", err)
	}
	if tok.UserID != userID {
		return errors.New("session not found")
	}

	if err := b.tokenStore.RevokeBySessionID(ctx, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return b.invalidateSession(ctx, sessionID)
}
