package v1

import (
	"github.com/Evgen-Poloniy/url-shortener/internal/middleware"
	"github.com/gin-gonic/gin"

	_ "github.com/Evgen-Poloniy/url-shortener/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(router *gin.Engine, handler *Handler) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")

	protected := v1.Group("")
	protected.Use(
		middleware.APIKeyAuth(handler.authConfig.ApiKey),
	)
	{
		protected.POST("/urls", handler.CreateShortURL)
		protected.GET("/urls/:short_url", handler.GetFullURL)
	}
}
