package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Evgen-Poloniy/url-shortener/internal/config"
	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/Evgen-Poloniy/url-shortener/internal/middleware"
	pg "github.com/Evgen-Poloniy/url-shortener/internal/repository/postgres"
	"github.com/Evgen-Poloniy/url-shortener/internal/service/shortener"
	router "github.com/Evgen-Poloniy/url-shortener/internal/transport/http"
	v1 "github.com/Evgen-Poloniy/url-shortener/internal/transport/http/v1"
	"github.com/Evgen-Poloniy/url-shortener/pkg/logs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS urls (
		id BIGSERIAL PRIMARY KEY,
		full_url TEXT NOT NULL UNIQUE,
		short_url VARCHAR(10) NOT NULL UNIQUE
	);`

	db.MustExec(schema)

	teardown := func() {
		_ = db.Close()
		termCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := pgContainer.Terminate(termCtx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return db, teardown
}

func TestURLShortener_IntegrationPostgres(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	apiKey := "12345"
	cfgAuth := config.AuthConfig{
		ApiKey: apiKey,
	}
	cfgCORS := config.CORSConfig{}
	cfgShortener := config.ShortenerConfig{
		AllowedProtocols: []string{"http", "https"},
	}

	logger := logs.NewLogrusLogger(
		logs.WithLevel("error"),
		logs.WithFormat("text"),
	)

	repository := pg.NewPostgresRepository(db)
	service := shortener.NewShortenerService(repository)
	v1Handler := v1.NewHandler(service, &cfgAuth, &cfgShortener)

	appRouter := router.NewRouter(&cfgCORS, logger)
	v1.NewRouter(appRouter, v1Handler)

	t.Run("Negative and Validation Cases", func(t *testing.T) {
		tests := []struct {
			name         string
			method       string
			path         string
			apiKeyHeader string
			contentType  string
			body         any
			expectedCode int
			expectedErr  domain.ErrCode
		}{
			{
				name:         "POST /urls - Unauthorized (Missing Key)",
				method:       http.MethodPost,
				path:         "/api/v1/urls",
				apiKeyHeader: "",
				contentType:  "application/json",
				body:         v1.CreateShortURLReq{URL: "https://example.com"},
				expectedCode: http.StatusUnauthorized,
				expectedErr:  domain.CodeMissingAuthHeaders,
			},
			{
				name:         "POST /urls - Unauthorized (Invalid Key)",
				method:       http.MethodPost,
				path:         "/api/v1/urls",
				apiKeyHeader: "wrong-key",
				contentType:  "application/json",
				body:         v1.CreateShortURLReq{URL: "https://example.com"},
				expectedCode: http.StatusUnauthorized,
				expectedErr:  domain.CodeInvalidAPIKey,
			},
			{
				name:         "POST /urls - Bad Request (Empty URL Field)",
				method:       http.MethodPost,
				path:         "/api/v1/urls",
				apiKeyHeader: apiKey,
				contentType:  "application/json",
				body:         v1.CreateShortURLReq{URL: ""},
				expectedCode: http.StatusBadRequest,
				expectedErr:  domain.CodeInvalidInput,
			},
			{
				name:         "POST /urls - Bad Request (Missing Protocol Scheme)",
				method:       http.MethodPost,
				path:         "/api/v1/urls",
				apiKeyHeader: apiKey,
				contentType:  "application/json",
				body:         v1.CreateShortURLReq{URL: "example.com/no-scheme"},
				expectedCode: http.StatusBadRequest,
				expectedErr:  domain.CodeInvalidURL,
			},
			{
				name:         "POST /urls - Bad Request (Unsupported Protocol)",
				method:       http.MethodPost,
				path:         "/api/v1/urls",
				apiKeyHeader: apiKey,
				contentType:  "application/json",
				body:         v1.CreateShortURLReq{URL: "ftp://example.com/file"},
				expectedCode: http.StatusBadRequest,
				expectedErr:  domain.CodeInvalidURL,
			},
			{
				name:         "POST /urls - Bad Request (Missing Host)",
				method:       http.MethodPost,
				path:         "/api/v1/urls",
				apiKeyHeader: apiKey,
				contentType:  "application/json",
				body:         v1.CreateShortURLReq{URL: "https://"},
				expectedCode: http.StatusBadRequest,
				expectedErr:  domain.CodeInvalidURL,
			},
			{
				name:         "POST /urls - Bad Request (Malformed JSON)",
				method:       http.MethodPost,
				path:         "/api/v1/urls",
				apiKeyHeader: apiKey,
				contentType:  "application/json",
				body:         "{invalid-json",
				expectedCode: http.StatusBadRequest,
				expectedErr:  domain.CodeInvalidInput,
			},
			{
				name:         "GET /urls/{short_url} - Not Found",
				method:       http.MethodGet,
				path:         "/api/v1/urls/notfound10",
				apiKeyHeader: apiKey,
				contentType:  "",
				body:         nil,
				expectedCode: http.StatusNotFound,
				expectedErr:  domain.CodeURLNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var reqBody []byte
				if strBody, ok := tt.body.(string); ok {
					reqBody = []byte(strBody)
				} else if tt.body != nil {
					reqBody, _ = json.Marshal(tt.body)
				}

				req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(reqBody))
				if tt.contentType != "" {
					req.Header.Set("Content-Type", tt.contentType)
				}
				if tt.apiKeyHeader != "" {
					req.Header.Set("X-API-Key", tt.apiKeyHeader)
				}

				w := httptest.NewRecorder()
				appRouter.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedCode, w.Code)

				var errResp middleware.ResponseError
				err := json.NewDecoder(w.Body).Decode(&errResp)
				require.NoError(t, err)
				assert.Equal(t, string(tt.expectedErr), errResp.Error.Code)
				assert.NotEmpty(t, errResp.Error.Message)
			})
		}
	})

	t.Run("Positive Flow Cases", func(t *testing.T) {
		targetURL := "https://example.com/flow-test-link"
		var createdHash string

		t.Run("1. Create Short URL Success", func(t *testing.T) {
			reqDTO := v1.CreateShortURLReq{URL: targetURL}
			reqBody, _ := json.Marshal(reqDTO)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", apiKey)

			w := httptest.NewRecorder()
			appRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var resp v1.DataResp
			err := json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			dataBytes, _ := json.Marshal(resp.Data)
			var shortResp v1.ShortURLResp
			err = json.Unmarshal(dataBytes, &shortResp)
			require.NoError(t, err)

			assert.Len(t, shortResp.ShortURL, 10)
			createdHash = shortResp.ShortURL
		})

		t.Run("2. Create Duplicate URL Returns Same Hash", func(t *testing.T) {
			reqDTO := v1.CreateShortURLReq{URL: targetURL}
			reqBody, _ := json.Marshal(reqDTO)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", apiKey)

			w := httptest.NewRecorder()
			appRouter.ServeHTTP(w, req)

			assert.True(t, w.Code == http.StatusCreated || w.Code == http.StatusOK)

			var resp v1.DataResp
			err := json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			dataBytes, _ := json.Marshal(resp.Data)
			var shortResp v1.ShortURLResp
			err = json.Unmarshal(dataBytes, &shortResp)
			require.NoError(t, err)

			assert.Equal(t, createdHash, shortResp.ShortURL)
		})

		t.Run("3. Get Original URL Redirect", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/urls/"+createdHash, nil)
			req.Header.Set("X-API-Key", apiKey)
			w := httptest.NewRecorder()

			appRouter.ServeHTTP(w, req)

			if w.Code == http.StatusTemporaryRedirect || w.Code == http.StatusFound {
				assert.Equal(t, targetURL, w.Header().Get("Location"))
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
				var resp v1.DataResp
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(resp.Data)
				var fullResp v1.FullURLResp
				err = json.Unmarshal(dataBytes, &fullResp)
				require.NoError(t, err)

				assert.Equal(t, targetURL, fullResp.URL)
			}
		})

		t.Run("4. Concurrent Creation of Same URL", func(t *testing.T) {
			const numWorkers = 50
			sharedURL := "https://example.com/shared-concurrent-link"
			var wg sync.WaitGroup
			wg.Add(numWorkers)

			hashes := make([]string, numWorkers)

			for i := 0; i < numWorkers; i++ {
				go func(workerID int) {
					defer wg.Done()

					reqDTO := v1.CreateShortURLReq{URL: sharedURL}
					reqBody, _ := json.Marshal(reqDTO)

					req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(reqBody))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("X-API-Key", apiKey)

					w := httptest.NewRecorder()
					appRouter.ServeHTTP(w, req)

					assert.True(t, w.Code == http.StatusCreated || w.Code == http.StatusOK, "Worker %d failed with code %d", workerID, w.Code)

					var resp v1.DataResp
					if err := json.NewDecoder(w.Body).Decode(&resp); err == nil {
						dataBytes, _ := json.Marshal(resp.Data)
						var shortResp v1.ShortURLResp
						_ = json.Unmarshal(dataBytes, &shortResp)
						hashes[workerID] = shortResp.ShortURL
					}
				}(i)
			}

			wg.Wait()

			firstHash := hashes[0]
			assert.NotEmpty(t, firstHash)
			for i := 1; i < numWorkers; i++ {
				assert.Equal(t, firstHash, hashes[i], "Worker %d got different short_url", i)
			}
		})

		t.Run("5. Concurrent Unique URL Creation", func(t *testing.T) {
			const numWorkers = 100
			var wg sync.WaitGroup
			wg.Add(numWorkers)

			for i := 0; i < numWorkers; i++ {
				go func(workerID int) {
					defer wg.Done()

					targetURL := fmt.Sprintf("https://example.com/concurrent-%d-%d", workerID, time.Now().UnixNano())
					reqDTO := v1.CreateShortURLReq{URL: targetURL}
					reqBody, _ := json.Marshal(reqDTO)

					req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(reqBody))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("X-API-Key", apiKey)

					w := httptest.NewRecorder()
					appRouter.ServeHTTP(w, req)

					assert.Equal(t, http.StatusCreated, w.Code, "Worker %d failed", workerID)

					var resp v1.DataResp
					if err := json.NewDecoder(w.Body).Decode(&resp); err == nil {
						dataBytes, _ := json.Marshal(resp.Data)
						var shortResp v1.ShortURLResp
						_ = json.Unmarshal(dataBytes, &shortResp)

						assert.Len(t, shortResp.ShortURL, 10)
					}
				}(i)
			}

			wg.Wait()
		})
	})
}
