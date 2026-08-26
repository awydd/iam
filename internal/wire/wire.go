//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/infra/cache/redis"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/store"
	"github.com/awydd/iam/internal/jwt"
	"github.com/awydd/iam/internal/transport/http/handler"
	"github.com/google/wire"
)

type App struct {
	Keypair *handler.KeypairHandler
	User    *handler.UserHandler
}

var handlerSet = wire.NewSet(
	handler.NewKeypairHandler,
	handler.NewUserHandler,
)

var bizSet = wire.NewSet(
	biz.NewKeypairBiz,
	biz.NewUserBiz,
)

var storeSet = wire.NewSet(
	store.NewTransactor,
	wire.Bind(new(biz.Transactor), new(*store.Transactor)),

	store.NewKeypairStore,
	wire.Bind(new(biz.KeypairStore), new(*store.KeypairStore)),

	store.NewUserStore,
	wire.Bind(new(biz.UserStore), new(*store.UserStore)),

	store.NewTokenStore,
	wire.Bind(new(biz.TokenStore), new(*store.TokenStore)),
)

var cacheSet = wire.NewSet(
	redis.NewTokenCache,
	wire.Bind(new(biz.TokenCache), new(*redis.TokenCache)),

	redis.NewSessionCache,
	wire.Bind(new(biz.SessionCache), new(*redis.SessionCache)),

	redis.NewLoginAttemptCache,
	wire.Bind(new(biz.LoginAttemptCache), new(*redis.LoginAttemptCache)),
)

func provideJWTManager(keypairBiz *biz.KeypairBiz) *jwt.Manager {
	cfg := conf.Get().JWT
	return jwt.New(keypairBiz, jwt.WithIssuer(cfg.Issuer), jwt.WithAudience(cfg.Audience))
}

var helperSet = wire.NewSet(
	provideJWTManager,
	wire.Bind(new(biz.TokenSigner), new(*jwt.Manager)),
)

var providerSet = wire.NewSet(
	handlerSet,
	bizSet,
	storeSet,
	cacheSet,
	helperSet,

	wire.Struct(new(App), "*"),
)

func InitApp(client *db.Client) *App {
	wire.Build(providerSet)
	return &App{}
}
