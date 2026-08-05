package shortener

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/shortener_mocks.go -package=mock_shortener_repository

// ShortenerRepository represents interface for work with URL shortener repository
type ShortenerRepository interface {
	// SaveURL saves full and short URLs into repository.
	SaveURL(ctx context.Context, full, short string) error

	// GetFullURL gets full URL from repository.
	GetFullURL(ctx context.Context, short string) (string, error)
}

type ShortenerService struct {
	shortenerRepository ShortenerRepository
}

func NewShortenerService(shortenerRepository ShortenerRepository) *ShortenerService {
	return &ShortenerService{
		shortenerRepository: shortenerRepository,
	}
}
