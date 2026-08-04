package shortener

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/go-playground/validator/v10"
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
	validate  *validator.Validate
	shortener ShortenerRepository
}

func NewShortenerService(shortener ShortenerRepository) *ShortenerService {
	return &ShortenerService{
		validate:  validator.New(),
		shortener: shortener,
	}
}

func (s *ShortenerService) validateData(ctx context.Context, data any) error {
	if err := s.validate.StructCtx(ctx, data); err != nil {
		var validationErrors validator.ValidationErrors

		if errors.As(err, &validationErrors) {
			messages := make([]string, 0, len(validationErrors))
			errMessages := make([]string, 0, len(validationErrors))

			for _, fieldErr := range validationErrors {
				var baseMsg string
				if fieldErr.Param() != "" {
					baseMsg = fmt.Sprintf("field=%s | tag=%s | value='%v' | expected=%s", fieldErr.Field(), fieldErr.Tag(), fieldErr.Value(), fieldErr.Param())
				} else {
					baseMsg = fmt.Sprintf("field=%s | tag=%s | value='%v'", fieldErr.Field(), fieldErr.Tag(), fieldErr.Value())
				}

				messages = append(messages, fmt.Sprintf("[%s]", baseMsg))
				errMessages = append(errMessages, fmt.Sprintf("[%s | err=%s]", baseMsg, fieldErr.Error()))
			}

			return domain.NewAppError(
				domain.CodeValidationError,
				fmt.Sprintf("validation failed: %s", strings.Join(messages, " | ")),
				fmt.Errorf("validation error details: %s", strings.Join(errMessages, " | ")),
			)
		}

		return domain.NewAppError(
			domain.CodeValidationError,
			"validation error",
			fmt.Errorf("validation error: %w", err),
		)
	}

	return nil
}
