package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/Evgen-Poloniy/url-shortener/internal/middleware"
)

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		errToRegister  error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "AppError URL Conflict",
			errToRegister:  domain.NewAppError(domain.CodeURLConflict, "url already exists", errors.New("conflict")),
			expectedStatus: http.StatusConflict,
			expectedCode:   string(domain.CodeURLConflict),
			expectedMsg:    "url already exists",
		},
		{
			name:           "AppError Missing Auth Header",
			errToRegister:  domain.NewAppError(domain.CodeMissingAuthHeaders, "missing header", domain.ErrMissingAuthHeader),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   string(domain.CodeMissingAuthHeaders),
			expectedMsg:    "missing header",
		},
		{
			name:           "AppError Unmapped Code Defaults To 500",
			errToRegister:  domain.NewAppError(domain.ErrCode("CUSTOM_UNMAPPED_CODE"), "custom error", errors.New("custom")),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "CUSTOM_UNMAPPED_CODE",
			expectedMsg:    "custom error",
		},
		{
			name:           "Unknown Native Error",
			errToRegister:  errors.New("unexpected internal issue"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "UNKNOWN_ERROR",
			expectedMsg:    "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(middleware.ErrorHandler())
			r.GET("/test", func(c *gin.Context) {
				_ = c.Error(tt.errToRegister)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var res middleware.ResponseError
			err := json.Unmarshal(w.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCode, res.Error.Code)
			assert.Equal(t, tt.expectedMsg, res.Error.Message)
		})
	}

	t.Run("No Errors Proceed Normally", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.ErrorHandler())
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})
}
