package middleware

import (
	"time"

	"github.com/awydd/iam/internal/logger"
	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if len(c.Errors) > 0 {
			logger.Error("%s %s %d %s client=%s err=%s",
				c.Request.Method, path, status, latency, c.ClientIP(), c.Errors.String())
			return
		}

		logger.Info("%s %s %d %s client=%s",
			c.Request.Method, path, status, latency, c.ClientIP())
	}
}
