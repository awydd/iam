package router

import (
	"net/http/pprof"
	"runtime"

	"github.com/awydd/iam/internal/logger"
	"github.com/gin-gonic/gin"
)

func registerPProf(r *gin.Engine, blockRate, mutexFraction int) {
	if blockRate > 0 {
		runtime.SetBlockProfileRate(blockRate)
		logger.Info("pprof: block profiling enabled, rate=%d", blockRate)
	}
	if mutexFraction > 0 {
		runtime.SetMutexProfileFraction(mutexFraction)
		logger.Info("pprof: mutex profiling enabled, fraction=%d", mutexFraction)
	}

	g := r.Group("/debug/pprof")
	g.GET("/", gin.WrapF(pprof.Index))
	g.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	g.GET("/profile", gin.WrapF(pprof.Profile))
	g.GET("/symbol", gin.WrapF(pprof.Symbol))
	g.POST("/symbol", gin.WrapF(pprof.Symbol))
	g.GET("/trace", gin.WrapF(pprof.Trace))

	g.GET("/:name", func(c *gin.Context) {
		pprof.Handler(c.Param("name")).ServeHTTP(c.Writer, c.Request)
	})
}
