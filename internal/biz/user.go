package biz

import (
	"context"

	"github.com/awydd/iam/internal/infra/ent/db"
)

type UserStore interface {
	GetSystem(ctx context.Context) (*db.User, error)
}
