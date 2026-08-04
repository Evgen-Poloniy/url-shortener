package middleware

import (
	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/gin-gonic/gin"
)

// APIKeyAuth checks request on API-Key availability and validity.
func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.Error(domain.NewAppError(
				domain.CodeMissingAuthHeaders,
				domain.ErrMissingAuthHeader.Error(),
				domain.ErrMissingAuthHeader,
			))
			c.Abort()
			return
		}

		if key != apiKey {
			c.Error(domain.NewAppError(
				domain.CodeInvalidAPIKey,
				domain.ErrInvalidAPIKey.Error(),
				domain.ErrInvalidAPIKey,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
