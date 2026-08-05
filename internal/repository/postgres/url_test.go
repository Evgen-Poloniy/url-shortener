package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/Evgen-Poloniy/url-shortener/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_SaveURL(t *testing.T) {
	query := "INSERT INTO urls (full_url, short_url) VALUES ($1, $2) ON CONFLICT (full_url) DO NOTHING;"

	testCases := []struct {
		name      string
		full      string
		short     string
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name:  "Success - New URL",
			full:  "https://example.com",
			short: "abc1234567",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(query).
					WithArgs("https://example.com", "abc1234567").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: nil,
		},
		{
			name:  "Success - Same URL Idempotency (ON CONFLICT DO NOTHING)",
			full:  "https://example.com",
			short: "abc1234567",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(query).
					WithArgs("https://example.com", "abc1234567").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: nil,
		},
		{
			name:  "Conflict - Short Key Collision for Different URL",
			full:  "https://another-domain.com",
			short: "abc1234567",
			mockSetup: func(mock sqlmock.Sqlmock) {
				pgErr := &pgconn.PgError{Code: pgErrUniqueViolation}
				mock.ExpectExec(query).
					WithArgs("https://another-domain.com", "abc1234567").
					WillReturnError(pgErr)
			},
			wantErr: domain.ErrURLConflict,
		},
		{
			name:  "Internal Database Error",
			full:  "https://example.com",
			short: "abc1234567",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(query).
					WithArgs("https://example.com", "abc1234567").
					WillReturnError(errors.New("db disconnect"))
			},
			wantErr: domain.ErrInternalStorage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			tc.mockSetup(mock)

			repo := postgres.NewPostgresRepository(db)
			err := repo.SaveURL(context.Background(), tc.full, tc.short)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_GetFullURL(t *testing.T) {
	query := "SELECT full_url FROM urls WHERE short_url = $1;"

	testCases := []struct {
		name      string
		short     string
		mockSetup func(mock sqlmock.Sqlmock)
		wantFull  string
		wantErr   error
	}{
		{
			name:  "Success",
			short: "abc1234567",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"full_url"}).AddRow("https://example.com")
				mock.ExpectQuery(query).
					WithArgs("abc1234567").
					WillReturnRows(rows)
			},
			wantFull: "https://example.com",
			wantErr:  nil,
		},
		{
			name:  "Not Found",
			short: "notfound00",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs("notfound00").
					WillReturnError(sql.ErrNoRows)
			},
			wantFull: "",
			wantErr:  domain.ErrURLNotFound,
		},
		{
			name:  "Internal Database Error",
			short: "abc1234567",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs("abc1234567").
					WillReturnError(errors.New("db disconnect"))
			},
			wantFull: "",
			wantErr:  domain.ErrInternalStorage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			tc.mockSetup(mock)

			repo := postgres.NewPostgresRepository(db)
			full, err := repo.GetFullURL(context.Background(), tc.short)

			assert.Equal(t, tc.wantFull, full)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
