package v1

import (
	"github.com/Evgen-Poloniy/url-shortener/internal/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter(router *gin.Engine, handler *Handler) {
	v1 := router.Group("/api/v1")

	protected := v1.Group("")
	protected.Use(
		middleware.APIKeyAuth(handler.authConfig.ApiKey),
	)
	{
		protected.POST("/urls")
		protected.GET("/urls/:short_url")
	}
}
