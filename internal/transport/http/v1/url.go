package v1

import (
	"net/http"
	"strings"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/gin-gonic/gin"
)

// CreateShortURL handles POST /urls requests to generate a shortened URL.
// @Summary      Create a short URL
// @Description  Generates a 10-character short URL from the provided full URL.
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        request  body      CreateShortURLReq  true  "Original URL payload"
// @Success      201      {object}  v1.DataResp{data=v1.ShortURLResp}
// @Failure      400      {object}  middleware.ResponseError
// @Failure      401      {object}  middleware.ResponseError
// @Failure      409      {object}  middleware.ResponseError
// @Failure      500      {object}  middleware.ResponseError
// @Security     ApiKeyAuth
// @Router       /urls [post]
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
//
// @Summary      Get original URL
// @Description  Retrieves the original full URL associated with the provided short URL.
// @Tags         urls
// @Produce      json
// @Param        short_url  path      string  true  "Short URL hash (10 characters)"
// @Success      200        {object}  v1.DataResp{data=v1.FullURLResp}
// @Failure      400        {object}  middleware.ResponseError
// @Failure      401        {object}  middleware.ResponseError
// @Failure      404        {object}  middleware.ResponseError
// @Failure      500        {object}  middleware.ResponseError
// @Security     ApiKeyAuth
// @Router       /urls/{short_url} [get]
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
