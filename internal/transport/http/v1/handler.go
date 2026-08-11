package v1

import (
	"context"

	"github.com/Evgen-Poloniy/url-shortener/internal/config"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/shortener_mocks.go -package=mock_shortener_service

// UrlShortener represents the interface for save and get URLs.
type UrlShortener interface {
	// CreateShortURL creates short URL and saves full and short URLs into storage.
	CreateShortURL(ctx context.Context, full string) (string, error)

	// GetFullURL gets full URL from storage.
	GetFullURL(ctx context.Context, short string) (string, error)
}

type Handler struct {
	shortenerService UrlShortener
	authConfig       *config.AuthConfig
	shortenerConfig  *config.ShortenerConfig
}

func NewHandler(shortenerService UrlShortener, authConfig *config.AuthConfig, shortenerConfig *config.ShortenerConfig) *Handler {
	return &Handler{
		shortenerService: shortenerService,
		authConfig:       authConfig,
		shortenerConfig:  shortenerConfig,
	}
}
