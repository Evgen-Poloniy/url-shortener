package shortener_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/Evgen-Poloniy/url-shortener/internal/service/shortener"
	mock_shortener_repository "github.com/Evgen-Poloniy/url-shortener/internal/service/shortener/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestShortenerService_CreateShortURL(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name      string
		fullURL   string
		mockSetup func(m *mock_shortener_repository.MockShortenerRepository, fullURL string)
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:    "Success",
			fullURL: "https://example.com",
			mockSetup: func(m *mock_shortener_repository.MockShortenerRepository, fullURL string) {
				expectedShort := shortener.GenerateShortURL(fullURL)
				m.EXPECT().
					SaveURL(ctx, fullURL, expectedShort).
					Return(nil).
					Times(1)
			},
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "Conflict Error - Short URL Collision",
			fullURL: "https://conflict.com",
			mockSetup: func(m *mock_shortener_repository.MockShortenerRepository, fullURL string) {
				expectedShort := shortener.GenerateShortURL(fullURL)
				m.EXPECT().
					SaveURL(ctx, fullURL, expectedShort).
					Return(domain.ErrURLConflict).
					Times(1)
			},
			checkErr: func(t *testing.T, err error) {
				var appErr *domain.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, domain.CodeURLConflict, appErr.Code)
				assert.ErrorIs(t, appErr.Err, domain.ErrURLConflict)
			},
		},
		{
			name:    "Internal Storage Error",
			fullURL: "https://error.com",
			mockSetup: func(m *mock_shortener_repository.MockShortenerRepository, fullURL string) {
				expectedShort := shortener.GenerateShortURL(fullURL)
				m.EXPECT().
					SaveURL(ctx, fullURL, expectedShort).
					Return(errors.New("db connection timeout")).
					Times(1)
			},
			checkErr: func(t *testing.T, err error) {
				var appErr *domain.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, domain.CodeInternalStorage, appErr.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRepo := mock_shortener_repository.NewMockShortenerRepository(ctrl)

			tc.mockSetup(mockRepo, tc.fullURL)

			svc := shortener.NewShortenerService(mockRepo)
			short, err := svc.CreateShortURL(ctx, tc.fullURL)

			tc.checkErr(t, err)

			if err == nil {
				expectedShort := shortener.GenerateShortURL(tc.fullURL)
				assert.NotEmpty(t, short)
				assert.Equal(t, expectedShort, short)
			}
		})
	}
}

func TestShortenerService_GetFullURL(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name      string
		shortURL  string
		mockSetup func(m *mock_shortener_repository.MockShortenerRepository, shortURL string)
		wantURL   string
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:     "Success",
			shortURL: "abc1234567",
			mockSetup: func(m *mock_shortener_repository.MockShortenerRepository, shortURL string) {
				m.EXPECT().
					GetFullURL(ctx, shortURL).
					Return("https://example.com", nil).
					Times(1)
			},
			wantURL: "https://example.com",
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "Error - URL Not Found",
			shortURL: "notfound00",
			mockSetup: func(m *mock_shortener_repository.MockShortenerRepository, shortURL string) {
				m.EXPECT().
					GetFullURL(ctx, shortURL).
					Return("", domain.ErrURLNotFound).
					Times(1)
			},
			wantURL: "",
			checkErr: func(t *testing.T, err error) {
				var appErr *domain.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, domain.CodeURLNotFound, appErr.Code)
				assert.ErrorIs(t, appErr.Err, domain.ErrURLNotFound)
			},
		},
		{
			name:     "Internal Storage Error",
			shortURL: "dberror000",
			mockSetup: func(m *mock_shortener_repository.MockShortenerRepository, shortURL string) {
				m.EXPECT().
					GetFullURL(ctx, shortURL).
					Return("", errors.New("unexpected error")).
					Times(1)
			},
			wantURL: "",
			checkErr: func(t *testing.T, err error) {
				var appErr *domain.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, domain.CodeInternalStorage, appErr.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRepo := mock_shortener_repository.NewMockShortenerRepository(ctrl)

			tc.mockSetup(mockRepo, tc.shortURL)

			svc := shortener.NewShortenerService(mockRepo)
			gotURL, err := svc.GetFullURL(ctx, tc.shortURL)

			assert.Equal(t, tc.wantURL, gotURL)
			tc.checkErr(t, err)
		})
	}
}

func TestGenerateShortURL(t *testing.T) {
	testCases := []struct {
		name    string
		fullURL string
	}{
		{
			name:    "Standard URL",
			fullURL: "https://example.com",
		},
		{
			name:    "URL with path and params",
			fullURL: "https://site.ru/path/to/page?param=123&sort=asc",
		},
		{
			name:    "Short domain",
			fullURL: "https://ya.cc",
		},
		{
			name:    "Video service URL",
			fullURL: "https://youtu.be/dQw4w9WgXcQ",
		},
	}

	allowedAlphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key1 := shortener.GenerateShortURL(tc.fullURL)
			key2 := shortener.GenerateShortURL(tc.fullURL)

			assert.Len(t, key1, 10, "Key length must be exactly 10")

			assert.Equal(t, key1, key2, "GenerateShortURL must be deterministic")

			for _, char := range key1 {
				assert.Contains(t, allowedAlphabet, string(char), "Key contains invalid character: %c", char)
			}
		})
	}
}
