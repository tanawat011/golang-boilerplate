package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a gin middleware for structured logging using zap
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request details
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Request error", append(fields, zap.String("error", e.Error()))...)
			}
		} else {
			statusCode := c.Writer.Status()
			switch {
			case statusCode >= 500:
				logger.Error("Server error", fields...)
			case statusCode >= 400:
				logger.Warn("Client error", fields...)
			default:
				logger.Info("Request completed", fields...)
			}
		}
	}
}
