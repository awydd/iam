//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/store"
	"github.com/awydd/iam/internal/transport/http/handler"
	"github.com/google/wire"
)

type App struct {
	Keypair *handler.KeypairHandler
}

var handlerSet = wire.NewSet(
	handler.NewKeypairHandler,
)

var bizSet = wire.NewSet(
	biz.NewKeypairBiz,
)

var storeSet = wire.NewSet(
	store.NewTransactor,
	wire.Bind(new(biz.Transactor), new(*store.Transactor)),

	store.NewKeypairStore,
	wire.Bind(new(biz.KeypairStore), new(*store.KeypairStore)),
)

var providerSet = wire.NewSet(
	handlerSet,
	bizSet,
	storeSet,

	wire.Struct(new(App), "*"),
)

func InitApp(client *db.Client) *App {
	wire.Build(providerSet)
	return &App{}
}
