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
}
