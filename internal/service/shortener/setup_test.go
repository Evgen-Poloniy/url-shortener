package shortener_test

import (
	"testing"

	"github.com/Evgen-Poloniy/url-shortener/internal/service/shortener"
	mock_shortener_repository "github.com/Evgen-Poloniy/url-shortener/internal/service/shortener/mocks"
	"go.uber.org/mock/gomock"
)

func setupMock(t *testing.T) (*shortener.ShortenerService, *mock_shortener_repository.MockShortenerRepository) {
	ctrl := gomock.NewController(t)
	mockRepo := mock_shortener_repository.NewMockShortenerRepository(ctrl)
	svc := shortener.NewShortenerService(mockRepo)

	return svc, mockRepo
}
