package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	v1 "github.com/Evgen-Poloniy/url-shortener/internal/transport/http/v1"
	mock_shortener_service "github.com/Evgen-Poloniy/url-shortener/internal/transport/http/v1/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_CreateShortURL(t *testing.T) {
	type mockBehavior func(m *mock_shortener_service.MockUrlShortener, url string)

	tests := []struct {
		name         string
		reqBody      any
		urlInput     string
		mockBehavior mockBehavior
		expectedCode int
	}{
		{
			name:     "Success",
			reqBody:  v1.CreateShortURLReq{URL: "https://example.com/very/long/link"},
			urlInput: "https://example.com/very/long/link",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, url string) {
				m.EXPECT().
					CreateShortURL(gomock.Any(), url).
					Return("aB1_xY9z0A", nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Invalid JSON Body",
			reqBody:      "{invalid_json}",
			urlInput:     "",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, url string) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty URL Field",
			reqBody:      v1.CreateShortURLReq{URL: "   "},
			urlInput:     "",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, url string) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing Protocol Scheme",
			reqBody:      v1.CreateShortURLReq{URL: "example.com/path"},
			urlInput:     "",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, url string) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing Host",
			reqBody:      v1.CreateShortURLReq{URL: "https://"},
			urlInput:     "",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, url string) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Unsupported Protocol",
			reqBody:      v1.CreateShortURLReq{URL: "ftp://example.com/file"},
			urlInput:     "",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, url string) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:     "URL Conflict",
			reqBody:  v1.CreateShortURLReq{URL: "https://example.com/existing"},
			urlInput: "https://example.com/existing",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, url string) {
				m.EXPECT().
					CreateShortURL(gomock.Any(), url).
					Return("", domain.NewAppError(domain.CodeURLConflict, "url already exists", domain.ErrURLConflict))
			},
			expectedCode: http.StatusConflict,
		},
		{
			name:     "Internal Service Error",
			reqBody:  v1.CreateShortURLReq{URL: "https://example.com/err"},
			urlInput: "https://example.com/err",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, url string) {
				m.EXPECT().
					CreateShortURL(gomock.Any(), url).
					Return("", domain.NewAppError(domain.CodeInternalStorage, "db error", domain.ErrInternalStorage))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockService := setupMockService(ctrl)
			tt.mockBehavior(mockService, tt.urlInput)

			router := setupTestRouter(handler)

			var bodyBytes []byte
			if str, ok := tt.reqBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.reqBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/urls", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestHandler_GetFullURL(t *testing.T) {
	type mockBehavior func(m *mock_shortener_service.MockUrlShortener, shortURL string)

	tests := []struct {
		name          string
		shortURLParam string
		mockBehavior  mockBehavior
		expectedCode  int
	}{
		{
			name:          "Success",
			shortURLParam: "aB1_xY9z0A",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, shortURL string) {
				m.EXPECT().
					GetFullURL(gomock.Any(), shortURL).
					Return("https://example.com/very/long/link", nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Invalid Short URL Length (Handler Validation)",
			shortURLParam: "invalid_len",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, shortURL string) {
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:          "Invalid Input from Service",
			shortURLParam: "invalid100",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, shortURL string) {
				m.EXPECT().
					GetFullURL(gomock.Any(), shortURL).
					Return("", domain.NewAppError(domain.CodeInvalidInput, "invalid short url format", domain.ErrInvalidURLFormat))
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:          "URL Not Found",
			shortURLParam: "notfound10",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, shortURL string) {
				m.EXPECT().
					GetFullURL(gomock.Any(), shortURL).
					Return("", domain.NewAppError(domain.CodeURLNotFound, "url not found", domain.ErrURLNotFound))
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:          "Internal Service Error",
			shortURLParam: "err_code10",
			mockBehavior: func(m *mock_shortener_service.MockUrlShortener, shortURL string) {
				m.EXPECT().
					GetFullURL(gomock.Any(), shortURL).
					Return("", domain.NewAppError(domain.CodeInternalStorage, "db error", domain.ErrInternalStorage))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockService := setupMockService(ctrl)
			tt.mockBehavior(mockService, tt.shortURLParam)

			router := setupTestRouter(handler)

			req := httptest.NewRequest(http.MethodGet, "/urls/"+tt.shortURLParam, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
