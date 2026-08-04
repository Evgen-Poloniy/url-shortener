package router

import (
	"fmt"
	"net/http"

	"github.com/Evgen-Poloniy/url-shortener/internal/config"
	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/Evgen-Poloniy/url-shortener/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// NewRouter initializes HTTP API router and connects middlewares on target routers.
func NewRouter(corsConfig *config.CORSConfig, logger *logrus.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.HandleMethodNotAllowed = true

	router.NoRoute(func(c *gin.Context) {
		c.Error(domain.NewAppError(
			domain.CodeNotFound,
			fmt.Sprintf("endpoint with URL %s not found", c.Request.URL.Path),
			fmt.Errorf("endpoint with URL %s not found", c.Request.URL.Path),
		))
		c.Abort()
	})

	router.NoMethod(func(c *gin.Context) {
		c.Error(domain.NewAppError(
			domain.CodeMethodNotAllowed,
			fmt.Sprintf("method %s for URL %s not allowed", c.Request.Method, c.Request.URL.Path),
			fmt.Errorf("method %s for URL %s not allowed", c.Request.Method, c.Request.URL.Path),
		))
		c.Abort()
	})

	router.Use(
		gin.Recovery(),
		middleware.CORS(corsConfig),
		middleware.SecureHeaders(),
		middleware.Logger(logger),
		middleware.ErrorHandler(),
	)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
