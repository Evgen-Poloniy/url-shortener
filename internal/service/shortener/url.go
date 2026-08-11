package shortener

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
)

const (
	alphabet    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	shortUrlLen = 10
	base        = uint64(len(alphabet))
)

// CreateShortURL creates short URL and saves full and short URLs into storage.
func (s *ShortenerService) CreateShortURL(ctx context.Context, full string) (string, error) {
	short := GenerateShortURL(full)

	if err := s.shortenerRepository.SaveURL(ctx, full, short); err != nil {
		if errors.Is(err, domain.ErrURLConflict) {
			return "", domain.NewAppError(
				domain.CodeURLConflict,
				"failed to save URL",
				domain.ErrURLConflict,
			)
		}

		return "", domain.NewAppError(
			domain.CodeInternalStorage,
			"internal storage error",
			err,
		)
	}

	return short, nil
}

// GetFullURL gets full URL from repository.
func (s *ShortenerService) GetFullURL(ctx context.Context, short string) (string, error) {
	fullURL, err := s.shortenerRepository.GetFullURL(ctx, short)
	if err != nil {
		if errors.Is(err, domain.ErrURLNotFound) {
			return "", domain.NewAppError(
				domain.CodeURLNotFound,
				fmt.Sprintf("failed to get full url by short url %s", short),
				err,
			)
		}

		return "", domain.NewAppError(
			domain.CodeInternalStorage,
			"internal storage error",
			err,
		)
	}

	return fullURL, nil
}

// GenerateShortURL generates short URL from alphabet a-z, A-Z, 0-9, _
func GenerateShortURL(url string) string {
	hash := sha256.Sum256([]byte(url))

	num := binary.BigEndian.Uint64(hash[:8])

	var result [shortUrlLen]byte

	for i := 0; i < shortUrlLen; i++ {
		result[i] = alphabet[num%base]
		num /= base
	}

	return string(result[:])
}
