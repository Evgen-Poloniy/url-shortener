package v1

import (
	"fmt"
	"net/http"
	"net/url"
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

	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		_ = c.Error(domain.NewAppError(
			domain.CodeInvalidInput,
			"url is required",
			domain.ErrEmptyURL,
		))
		return
	}

	fullURL, err := h.validateFormatURL(rawURL)
	if err != nil {
		_ = c.Error(err)
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
	if len(shortURL) != 10 {
		_ = c.Error(domain.NewAppError(
			domain.CodeInvalidInput,
			domain.ErrLenURL.Error(),
			domain.ErrLenURL,
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

// validateFormatURL validates URL format and allowed protocols from config.
func (h *Handler) validateFormatURL(rawURL string) (string, error) {
	if !strings.Contains(rawURL, "://") {
		return "", domain.NewAppError(
			domain.CodeInvalidURL,
			"URL must include protocol scheme (://)",
			domain.ErrMissingScheme,
		)
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", domain.NewAppError(
			domain.CodeInvalidURL,
			"invalid URL format",
			domain.ErrInvalidURLFormat,
		)
	}

	if u.Host == "" {
		return "", domain.NewAppError(
			domain.CodeInvalidURL,
			"URL host cannot be empty",
			domain.ErrMissingHost,
		)
	}

	currentScheme := strings.ToLower(u.Scheme)
	allowed := false

	for _, proto := range h.shortenerConfig.AllowedProtocols {
		if strings.ToLower(strings.TrimSpace(proto)) == currentScheme {
			allowed = true
			break
		}
	}

	if !allowed {
		return "", domain.NewAppError(
			domain.CodeInvalidURL,
			fmt.Sprintf("protocol '%s' is not allowed", u.Scheme),
			domain.ErrUnsupportedScheme,
		)
	}

	return u.String(), nil
}
