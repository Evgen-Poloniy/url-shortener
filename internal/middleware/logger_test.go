package middleware_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/Evgen-Poloniy/url-shortener/internal/middleware"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		handler        gin.HandlerFunc
		expectedLevel  string
		expectedMsg    string
		expectedCode   string
		expectedStatus string
	}{
		{
			name: "Success 200 OK",
			handler: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
			expectedLevel:  "level=info",
			expectedMsg:    "msg=\"request completed successfully\"",
			expectedStatus: "status_code=200",
		},
		{
			name: "Client Error 400 With Error Context",
			handler: func(c *gin.Context) {
				c.Set("code", "VALIDATION_ERROR")
				_ = c.Error(errors.New("invalid request body"))
				c.Status(http.StatusBadRequest)
			},
			expectedLevel:  "level=warning",
			expectedMsg:    "msg=\"invalid request body\"",
			expectedCode:   "code=VALIDATION_ERROR",
			expectedStatus: "status_code=400",
		},
		{
			name: "Server Error 500 With Error Context",
			handler: func(c *gin.Context) {
				c.Set("code", "INTERNAL_ERROR")
				_ = c.Error(errors.New("db connection timeout"))
				c.Status(http.StatusInternalServerError)
			},
			expectedLevel:  "level=error",
			expectedMsg:    "msg=\"db connection timeout\"",
			expectedCode:   "code=INTERNAL_ERROR",
			expectedStatus: "status_code=500",
		},
		{
			name: "Server Error 500 Without Errors Slice",
			handler: func(c *gin.Context) {
				c.Status(http.StatusInternalServerError)
			},
			expectedLevel:  "level=error",
			expectedMsg:    "msg=\"request completed with error\"",
			expectedStatus: "status_code=500",
		},
		{
			name: "Client Error 404 Without Errors Slice",
			handler: func(c *gin.Context) {
				c.Status(http.StatusNotFound)
			},
			expectedLevel:  "level=warning",
			expectedMsg:    "msg=\"request completed with warning\"",
			expectedStatus: "status_code=404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.TextFormatter{
				DisableTimestamp: true,
			})

			r := gin.New()
			r.Use(middleware.Logger(logger))
			r.GET("/test", tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			logOutput := buf.String()

			assert.Contains(t, logOutput, tt.expectedLevel)
			assert.Contains(t, logOutput, tt.expectedMsg)
			assert.Contains(t, logOutput, tt.expectedStatus)
			assert.Contains(t, logOutput, "path=/test")
			assert.Contains(t, logOutput, "method=GET")

			if tt.expectedCode != "" {
				assert.Contains(t, logOutput, tt.expectedCode)
			}
		})
	}
}
