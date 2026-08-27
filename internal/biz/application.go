package biz

import (
	"context"
	"errors"
	"fmt"

	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/logger"
	"github.com/awydd/iam/pkg/hashutil"
	"github.com/awydd/iam/pkg/random"
)

type ApplicationStore interface {
	List(ctx context.Context, keyword string, page, perPage int) ([]*db.Application, int, error)

	Get(ctx context.Context, id int) (*db.Application, error)
	GetByClientID(ctx context.Context, clientID string) (*db.Application, error)
	GetSystem(ctx context.Context) (*db.Application, error)

	Duplicate(ctx context.Context, name, clientID string, id ...int) (bool, error)

	Create(ctx context.Context, name, clientID string, clientSecret []byte, redirectUris []string, clientType enum.ApplicationClientType, status enum.ApplicationStatus, accessTokenTTL, refreshTokenTTL int) (*db.Application, error)
	InitCreate(ctx context.Context, name, clientID string, clientSecretHash []byte, clientType enum.ApplicationClientType, redirectUris []string) (*db.Application, error)

	Update(ctx context.Context, id int, name, clientID string, redirectUris []string) (*db.Application, error)
	UpdateTTL(ctx context.Context, id int, accessTokenTTL, refreshTokenTTL int) error
	UpdateSecret(ctx context.Context, id int, secretHash []byte) error
	UpdateStatus(ctx context.Context, id int, status enum.ApplicationStatus) error

	Delete(ctx context.Context, id int) error
}

type ApplicationBiz struct {
	store      ApplicationStore
	tokenStore TokenStore
	tokenCache TokenCache
}

func NewApplicationBiz(store ApplicationStore, tokenStore TokenStore, tokenCache TokenCache) *ApplicationBiz {
	return &ApplicationBiz{store: store, tokenStore: tokenStore, tokenCache: tokenCache}
}

type ApplicationListItemResp struct {
	ID           int                        `json:"id"`
	Name         string                     `json:"name"`
	ClientID     string                     `json:"client_id"`
	RedirectUris []string                   `json:"redirect_uris"`
	Type         enum.ApplicationClientType `json:"type"`
	Status       enum.ApplicationStatus     `json:"status"`
}

func (b *ApplicationBiz) List(ctx context.Context, keyword string, page, perPage int) ([]ApplicationListItemResp, int, error) {
	list, count, err := b.store.List(ctx, keyword, page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to get list")
	}

	listRes := make([]ApplicationListItemResp, 0, len(list))
	for _, item := range list {
		listRes = append(listRes, ApplicationListItemResp{
			ID:           item.ID,
			Name:         item.Name,
			ClientID:     item.ClientID,
			RedirectUris: item.RedirectUris,
			Type:         item.Type,
			Status:       item.Status,
		})
	}

	return listRes, count, nil
}

type ApplicationInfoResp struct {
	ID           int                        `json:"id"`
	Name         string                     `json:"name"`
	ClientID     string                     `json:"client_id"`
	RedirectUris []string                   `json:"redirect_uris"`
	Type         enum.ApplicationClientType `json:"type"`
	Status       enum.ApplicationStatus     `json:"status"`
}

func (b *ApplicationBiz) Info(ctx context.Context, id int) (*ApplicationInfoResp, error) {
	info, err := b.store.Get(ctx, id)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("application not found")
		}
		return nil, fmt.Errorf("get application: %w", err)
	}
	return &ApplicationInfoResp{
		ID:           info.ID,
		Name:         info.Name,
		ClientID:     info.ClientID,
		RedirectUris: info.RedirectUris,
		Type:         info.Type,
		Status:       info.Status,
	}, nil
}

type ApplicationCreateResp struct {
	ClientSecret string `json:"client_secret"`
}

func (b *ApplicationBiz) Create(
	ctx context.Context,
	name, clientID string,
	redirectUris []string,
	clientType enum.ApplicationClientType,
	status enum.ApplicationStatus,
	accessTokenTTL, refreshTokenTTL int,
) (*ApplicationCreateResp, error) {
	exists, err := b.store.Duplicate(ctx, name, clientID)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, errors.New("name or client_id already exists")
	}

	var clientSecret string
	var clientSecretHash []byte

	if clientType == enum.ApplicationClientTypeConfidential {
		clientSecret, err = random.Unambiguous(32)
		if err != nil {
			return nil, fmt.Errorf("generate client_secret: %w", err)
		}
		clientSecretHash = hashutil.Sum256([]byte(clientSecret))
	}

	_, err = b.store.Create(ctx, name, clientID, clientSecretHash, redirectUris, clientType, status, accessTokenTTL, refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	return &ApplicationCreateResp{ClientSecret: clientSecret}, nil
}

func (b *ApplicationBiz) UpdateInfo(ctx context.Context, id int, name, clientID string, redirectUris []string) error {
	app, err := b.store.Get(ctx, id)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("application not found")
		}
		return fmt.Errorf("get application: %w", err)
	}
	if app.IsSystem {
		return errors.New("system application cannot be updated via this endpoint")
	}

	exists, err := b.store.Duplicate(ctx, name, clientID, id)
	if err != nil {
		return fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return errors.New("name or client_id already exists")
	}

	if _, err := b.store.Update(ctx, id, name, clientID, redirectUris); err != nil {
		if db.IsNotFound(err) {
			return errors.New("application not found")
		}
		return fmt.Errorf("update application: %w", err)
	}
	return nil
}

func (b *ApplicationBiz) UpdateTTL(ctx context.Context, id, accessTokenTTL, refreshTokenTTL int) error {
	app, err := b.store.Get(ctx, id)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("application not found")
		}
		return fmt.Errorf("get application: %w", err)
	}
	if app.IsSystem {
		return errors.New("system application cannot be updated via this endpoint")
	}

	if err := b.store.UpdateTTL(ctx, id, accessTokenTTL, refreshTokenTTL); err != nil {
		if db.IsNotFound(err) {
			return errors.New("application not found")
		}
		return fmt.Errorf("update application ttl: %w", err)
	}
	return nil
}

func (b *ApplicationBiz) UpdateStatus(ctx context.Context, id int, status enum.ApplicationStatus) error {
	app, err := b.store.Get(ctx, id)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("application not found")
		}
		return fmt.Errorf("get application: %w", err)
	}
	if app.IsSystem {
		return errors.New("system application cannot be updated via this endpoint")
	}

	if err := b.store.UpdateStatus(ctx, id, status); err != nil {
		if db.IsNotFound(err) {
			return errors.New("application not found")
		}
		return fmt.Errorf("update application status: %w", err)
	}

	if status != enum.ApplicationStatusActive {
		if err := b.invalidateApplicationSessions(ctx, id); err != nil {
			logger.Error("invalidate application sessions failed: application_id=%d err=%v", id, err)
		}
	}
	return nil
}

type ApplicationUpdateSecretResp struct {
	ClientSecret string `json:"client_secret"`
}

func (b *ApplicationBiz) UpdateSecret(ctx context.Context, id int) (*ApplicationUpdateSecretResp, error) {
	info, err := b.store.Get(ctx, id)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("application not found")
		}
		return nil, fmt.Errorf("get application: %w", err)
	}

	if info.IsSystem {
		return nil, errors.New("system application secret cannot be rotated via this endpoint")
	}

	if info.Type != enum.ApplicationClientTypeConfidential {
		return nil, errors.New("only confidential clients can have a client_secret")
	}

	if info.Status != enum.ApplicationStatusActive {
		return nil, errors.New("cannot rotate secret for a disabled application")
	}

	clientSecret, err := random.Unambiguous(32)
	if err != nil {
		return nil, fmt.Errorf("generate client_secret: %w", err)
	}

	clientSecretHash := hashutil.Sum256([]byte(clientSecret))
	if err := b.store.UpdateSecret(ctx, id, clientSecretHash); err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("application not found")
		}
		return nil, fmt.Errorf("update application secret: %w", err)
	}

	return &ApplicationUpdateSecretResp{
		ClientSecret: clientSecret,
	}, nil
}

func (b *ApplicationBiz) invalidateApplicationSessions(ctx context.Context, id int) error {
	sessionIDs, err := b.tokenStore.ListActiveSessionsByApplicationID(ctx, id)
	if err != nil {
		return fmt.Errorf("list active sessions: %w", err)
	}

	if err := b.tokenStore.RevokeByApplicationID(ctx, id); err != nil {
		return fmt.Errorf("revoke tokens: %w", err)
	}

	var errs []error
	for _, sid := range sessionIDs {
		if err := b.tokenCache.DelAccess(ctx, sid); err != nil {
			errs = append(errs, fmt.Errorf("del access cache: %w", err))
		}
		if err := b.tokenCache.DelRefresh(ctx, sid); err != nil {
			errs = append(errs, fmt.Errorf("del refresh cache: %w", err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (b *ApplicationBiz) Delete(ctx context.Context, id int) error {
	sys, err := b.store.GetSystem(ctx)
	if err == nil && sys != nil && sys.ID == id {
		return errors.New("cannot delete system application")
	}

	if err := b.store.Delete(ctx, id); err != nil {
		if db.IsNotFound(err) {
			return errors.New("application not found")
		}
		return fmt.Errorf("delete application: %w", err)
	}

	sessionIDs, err := b.tokenStore.ListActiveSessionsByApplicationID(ctx, id)
	if err != nil {
		return fmt.Errorf("list active sessions: %w", err)
	}

	if err := b.tokenStore.RevokeByApplicationID(ctx, id); err != nil {
		logger.Error("revoke tokens by application failed: application_id=%d err=%v", id, err)
	}

	for _, sid := range sessionIDs {
		if err := b.tokenCache.DelAccess(ctx, sid); err != nil {
			logger.Error("del access cache failed: session_id=%s err=%v", sid, err)
		}
		if err := b.tokenCache.DelRefresh(ctx, sid); err != nil {
			logger.Error("del refresh cache failed: session_id=%s err=%v", sid, err)
		}
	}

	return nil
}
