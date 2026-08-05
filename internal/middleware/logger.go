package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		entry := logger.WithFields(map[string]interface{}{
			"id":          requestID,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"ip":          c.ClientIP(),
			"latency":     fmt.Sprintf("%v", latency),
			"status_code": statusCode,
		})

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if val, exists := c.Get("code"); exists {
				if v, ok := val.(string); ok {
					entry = entry.WithField("code", v)
				}
			}

			if statusCode >= 500 {
				entry.Error(err.Error())
			} else {
				entry.Warn(err.Error())
			}

			return
		}

		switch {
		case statusCode >= 500:
			entry.Error("request completed with error")
		case statusCode >= 400:
			entry.Warn("request completed with warning")
		default:
			entry.Info("request completed successfully")
		}
	}
}
