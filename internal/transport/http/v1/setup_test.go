package v1_test

import (
	"errors"
	"net/http"

	"github.com/Evgen-Poloniy/url-shortener/internal/config"
	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	v1 "github.com/Evgen-Poloniy/url-shortener/internal/transport/http/v1"
	mock_shortener_service "github.com/Evgen-Poloniy/url-shortener/internal/transport/http/v1/mocks"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupMockService(ctrl *gomock.Controller) (*v1.Handler, *mock_shortener_service.MockUrlShortener) {
	mockService := mock_shortener_service.NewMockUrlShortener(ctrl)
	handler := v1.NewHandler(mockService, nil, &config.ShortenerConfig{
		AllowedProtocols: []string{"http", "https"},
	})

	return handler, mockService
}

func setupTestRouter(handler *v1.Handler) *gin.Engine {
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var appErr *domain.AppError
			if errors.As(err, &appErr) {
				status := http.StatusInternalServerError
				switch appErr.Code {
				case domain.CodeInvalidInput, domain.CodeInvalidURL, domain.CodeValidationError:
					status = http.StatusBadRequest
				case domain.CodeURLNotFound:
					status = http.StatusNotFound
				case domain.CodeURLConflict:
					status = http.StatusConflict
				}
				c.JSON(status, gin.H{
					"error": appErr.Message,
					"code":  appErr.Code,
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}
		}
	})

	api := r.Group("/urls")
	{
		api.POST("", handler.CreateShortURL)
		api.GET("/:short_url", handler.GetFullURL)
	}

	return r
}
