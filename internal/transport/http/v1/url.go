package v1

import (
	"net/http"
	"strings"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/gin-gonic/gin"
)

// CreateShortURL handles POST /urls requests to generate a shortened URL.
func (h *Handler) CreateShortURL(c *gin.Context) {
	var req CreateShortURLReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domain.NewAppError(
			domain.CodeInvalidInput,
			"invalid request body format",
			err,
		))
		return
	}

	fullURL := strings.TrimSpace(req.URL)
	if fullURL == "" {
		_ = c.Error(domain.NewAppError(
			domain.CodeInvalidInput,
			"url is required",
			domain.ErrEmptyURL,
		))
		return
	}

	shortURL, err := h.shortenerService.CreateShortURL(c.Request.Context(), fullURL)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, DataResp{
		Data: ShortURLResp{
			ShortURL: shortURL,
		},
	})
}

// GetFullURL handles GET /urls/:short_url requests to retrieve the original URL.
func (h *Handler) GetFullURL(c *gin.Context) {
	shortURL := c.Param("short_url")
	if shortURL == "" {
		_ = c.Error(domain.NewAppError(
			domain.CodeInvalidInput,
			"short_url parameter is required",
			domain.ErrEmptyURL,
		))
		return
	}

	fullURL, err := h.shortenerService.GetFullURL(c.Request.Context(), shortURL)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, DataResp{
		Data: FullURLResp{
			URL: fullURL,
		},
	})
}
