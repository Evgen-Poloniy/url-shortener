package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgErrUniqueViolation = "23505"

// SaveURL saves full and short URLs into repository.
func (r *PostgresRepository) SaveURL(ctx context.Context, full, short string) error {
	query := `INSERT INTO urls (full_url, short_url) VALUES ($1, $2);`

	_, err := r.db.ExecContext(ctx, query, full, short)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgErrUniqueViolation {
				return domain.ErrURLConflict
			}
		}

		return fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	return nil
}

// GetFullURL gets full URL from repository.
func (r *PostgresRepository) GetFullURL(ctx context.Context, short string) (string, error) {
	query := `SELECT full_url FROM urls WHERE short_url = $1;`

	var full string
	err := r.db.GetContext(ctx, &full, query, short)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrURLNotFound
		}
		return "", fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	return full, nil
}
