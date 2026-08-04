package v1

import (
	"github.com/Evgen-Poloniy/url-shortener/internal/config"
)

type UrlShortener interface {
}

type Handler struct {
	shortener  UrlShortener
	authConfig *config.AuthConfig
}

func NewHandler(shortener UrlShortener, authConfig *config.AuthConfig) *Handler {
	return &Handler{
		shortener:  shortener,
		authConfig: authConfig,
	}
}
