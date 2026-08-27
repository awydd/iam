package store

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/ent/db/application"
)

type ApplicationStore struct {
	*baseStore
}

func NewApplicationStore(client *db.Client) *ApplicationStore {
	return &ApplicationStore{baseStore: newBaseStore(client)}
}

var _ biz.ApplicationStore = (*ApplicationStore)(nil)

func (s *ApplicationStore) List(ctx context.Context, keyword string, page, perPage int) ([]*db.Application, int, error) {
	q := s.Client(ctx).Application.Query().
		Where(application.DeletedAtIsNil())

	if keyword != "" {
		q = q.Where(
			application.Or(
				application.NameContainsFold(keyword),
				application.ClientIDContainsFold(keyword),
			),
		)
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}
	if total == 0 {
		return []*db.Application{}, 0, nil
	}

	offset, limit, ok := paginate(total, page, perPage)
	if !ok {
		return []*db.Application{}, total, nil
	}

	list, err := q.
		Order(db.Desc(application.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list applications: %w", err)
	}

	return list, total, nil
}

func (s *ApplicationStore) ListByIds(ctx context.Context, ids []int) ([]*db.Application, error) {
	if len(ids) == 0 {
		return []*db.Application{}, nil
	}

	list, err := s.Client(ctx).Application.Query().
		Where(application.IDIn(ids...), application.DeletedAtIsNil()).
		Order(db.Desc(application.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list applications by ids: %w", err)
	}
	return list, nil
}

func (s *ApplicationStore) Get(ctx context.Context, id int) (*db.Application, error) {
	app, err := s.Client(ctx).Application.Query().
		Where(application.IDEQ(id), application.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get application by id %d: %w", id, err)
	}
	return app, nil
}

func (s *ApplicationStore) GetByClientID(ctx context.Context, clientID string) (*db.Application, error) {
	app, err := s.Client(ctx).Application.Query().
		Where(application.ClientIDEQ(clientID), application.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get application by client_id %s: %w", clientID, err)
	}
	return app, nil
}

func (s *ApplicationStore) GetSystem(ctx context.Context) (*db.Application, error) {
	app, err := s.Client(ctx).Application.Query().
		Where(application.IsSystem(true), application.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get system application: %w", err)
	}
	return app, nil
}

func (s *ApplicationStore) Duplicate(ctx context.Context, name, clientID string, id ...int) (bool, error) {
	q := s.Client(ctx).Application.Query().
		Where(
			application.DeletedAtIsNil(),
			application.Or(
				application.NameEQ(name),
				application.ClientIDEQ(clientID),
			),
		)

	if len(id) > 0 && id[0] > 0 {
		q = q.Where(application.IDNEQ(id[0]))
	}

	exist, err := q.Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("check application duplicate: %w", err)
	}
	return exist, nil
}

func (s *ApplicationStore) Create(
	ctx context.Context,
	name, clientID string,
	clientSecret []byte,
	redirectUris []string,
	clientType enum.ApplicationClientType,
	status enum.ApplicationStatus,
	accessTokenTTL, refreshTokenTTL int,
) (*db.Application, error) {
	res, err := s.Client(ctx).Application.Create().
		SetName(name).
		SetClientID(clientID).
		SetClientSecret(clientSecret).
		SetRedirectUris(redirectUris).
		SetType(clientType).
		SetStatus(status).
		SetAccessTokenTTL(accessTokenTTL).
		SetRefreshTokenTTL(refreshTokenTTL).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	return res, nil
}

func (s *ApplicationStore) InitCreate(
	ctx context.Context,
	name, clientID string,
	clientSecretHash []byte,
	clientType enum.ApplicationClientType,
	redirectUris []string,
) (*db.Application, error) {
	builder := s.Client(ctx).Application.Create().
		SetName(name).
		SetClientID(clientID).
		SetType(clientType).
		SetRedirectUris(redirectUris).
		SetIsSystem(true)

	if len(clientSecretHash) > 0 {
		builder = builder.SetClientSecret(clientSecretHash)
	}

	app, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create system application: %w", err)
	}
	return app, nil
}

func (s *ApplicationStore) Update(ctx context.Context, id int, name, clientID string, redirectUris []string) (*db.Application, error) {
	res, err := s.Client(ctx).Application.UpdateOneID(id).
		SetName(name).
		SetClientID(clientID).
		SetRedirectUris(redirectUris).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update application %d: %w", id, err)
	}
	return res, nil
}

func (s *ApplicationStore) UpdateTTL(ctx context.Context, id int, accessTokenTTL, refreshTokenTTL int) error {
	if err := s.Client(ctx).Application.UpdateOneID(id).
		SetAccessTokenTTL(accessTokenTTL).
		SetRefreshTokenTTL(refreshTokenTTL).
		Exec(ctx); err != nil {
		return fmt.Errorf("update application ttl %d: %w", id, err)
	}
	return nil
}

func (s *ApplicationStore) UpdateSecret(ctx context.Context, id int, secretHash []byte) error {
	if err := s.Client(ctx).Application.UpdateOneID(id).
		SetClientSecret(secretHash).
		Exec(ctx); err != nil {
		return fmt.Errorf("update application secret %d: %w", id, err)
	}
	return nil
}

func (s *ApplicationStore) UpdateStatus(ctx context.Context, id int, status enum.ApplicationStatus) error {
	if err := s.Client(ctx).Application.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx); err != nil {
		return fmt.Errorf("update application status %d: %w", id, err)
	}
	return nil
}

func (s *ApplicationStore) Delete(ctx context.Context, id int) error {
	app, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if app.IsSystem {
		return fmt.Errorf("system application cannot be deleted")
	}

	if err := s.Client(ctx).Application.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx); err != nil {
		return fmt.Errorf("delete application %d: %w", id, err)
	}
	return nil
}
