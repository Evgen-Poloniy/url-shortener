package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/Evgen-Poloniy/url-shortener/internal/middleware"
)

func TestAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expectedAPIKey := "secret-api-key"

	t.Run("Middleware Execution Flow", func(t *testing.T) {
		tests := []struct {
			name          string
			headerKey     string
			shouldProceed bool
		}{
			{
				name:          "Valid API Key",
				headerKey:     "secret-api-key",
				shouldProceed: true,
			},
			{
				name:          "Missing API Key Header",
				headerKey:     "",
				shouldProceed: false,
			},
			{
				name:          "Invalid API Key",
				headerKey:     "wrong-key",
				shouldProceed: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := gin.New()
				r.Use(middleware.APIKeyAuth(expectedAPIKey))

				handlerExecuted := false
				r.GET("/test", func(c *gin.Context) {
					handlerExecuted = true
					c.Status(http.StatusOK)
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				if tt.headerKey != "" {
					req.Header.Set("X-API-Key", tt.headerKey)
				}

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, tt.shouldProceed, handlerExecuted)
			})
		}
	})

	t.Run("Context Error Validation", func(t *testing.T) {
		tests := []struct {
			name         string
			headerKey    string
			expectedCode domain.ErrCode
		}{
			{
				name:         "Missing Header Sets CodeMissingAuthHeaders",
				headerKey:    "",
				expectedCode: domain.CodeMissingAuthHeaders,
			},
			{
				name:         "Invalid Header Sets CodeInvalidAPIKey",
				headerKey:    "wrong-key",
				expectedCode: domain.CodeInvalidAPIKey,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
				if tt.headerKey != "" {
					c.Request.Header.Set("X-API-Key", tt.headerKey)
				}

				authMiddleware := middleware.APIKeyAuth(expectedAPIKey)
				authMiddleware(c)

				assert.True(t, c.IsAborted())
				require.Len(t, c.Errors, 1)

				appErr, ok := c.Errors.Last().Err.(*domain.AppError)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, appErr.Code)
			})
		}
	})
}
