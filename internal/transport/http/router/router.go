package router

import (
	"net/http"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/transport/http/middleware"
	"github.com/awydd/iam/internal/wire"
	"github.com/gin-gonic/gin"
)

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func New(cfg *conf.HTTP, app *wire.App) *gin.Engine {
	isDev := conf.Get().IsDev()
	if isDev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogger())

	if len(cfg.TrustedProxies) > 0 {
		_ = r.SetTrustedProxies(cfg.TrustedProxies)
	} else {
		_ = r.SetTrustedProxies(nil)
	}

	if len(cfg.CORSAllowOrigins) > 0 {
		r.Use(middleware.CORS(cfg.CORSAllowOrigins))
	}

	r.GET("/healthz", healthCheck)

	if cfg.EnablePProf && isDev {
		registerPProf(r, cfg.BlockProfileRate, cfg.MutexProfileFraction)
	}

	registerRoutes(r, app)

	return r
}

func registerRoutes(r *gin.Engine, app *wire.App) {
	r.GET("/.well-known/jwks.json", app.Keypair.JWKS)

	api := r.Group(conf.Get().HTTP.APIBase())
	auth := api.Group("/auth")
	{
		auth.POST("/login", app.User.Login)
		auth.POST("/refresh", app.User.Refresh)
	}

	protected := api.Group("")
	protected.Use(middleware.Auth(app.JWTManager, app.TokenCache, app.TokenBiz))

	protected.POST("/auth/logout", app.User.Logout)
	protected.GET("/auth/me", app.User.Me)
	protected.PUT("/auth/me/password", app.User.Password)
	protected.GET("/auth/sessions", app.User.ListSessions)
	protected.DELETE("/auth/sessions/:session_id", app.User.RevokeSession)

	admin := protected.Group("")
	admin.Use(middleware.RequireSystem(app.UserBiz))
	{
		admin.GET("/keypairs", app.Keypair.List)
		admin.POST("/keypairs/rotate", app.Keypair.Rotate)
		admin.PUT("/keypairs/:kid/downgrade", app.Keypair.Downgrade)
		admin.PUT("/keypairs/:kid/retire", app.Keypair.Retire)

		admin.GET("/tokens", app.Token.List)
		admin.GET("/tokens/:id", app.Token.Info)
		admin.DELETE("/tokens/:id", app.Token.Revoke)

		admin.GET("/applications", app.Application.List)
		admin.GET("/applications/:id", app.Application.Info)
		admin.POST("/applications", app.Application.Create)
		admin.PUT("/applications/:id", app.Application.Update)
		admin.PUT("/applications/:id/status", app.Application.UpdateStatus)
		admin.PUT("/applications/:id/ttl", app.Application.UpdateTTL)
		admin.PUT("/applications/:id/secret", app.Application.UpdateSecret)
		admin.DELETE("/applications/:id", app.Application.Delete)
	}
}
