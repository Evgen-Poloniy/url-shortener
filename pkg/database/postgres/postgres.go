package postgres

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// NewPostgreSQL opens connection with the database.
func NewPostgreSQL(dsn string, opts ...Option) (*sqlx.DB, error) {
	options := &Options{
		maxOpenConns:        25,
		maxIdleConns:        25,
		connMaxLifetime:     5 * time.Minute,
		connMaxIdleLifetime: 2 * time.Minute,
	}

	for _, opt := range opts {
		opt(options)
	}

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("error when open connection with the database: %w", err)
	}

	db.SetMaxOpenConns(options.maxOpenConns)
	db.SetMaxIdleConns(options.maxIdleConns)
	db.SetConnMaxLifetime(options.connMaxLifetime)
	db.SetConnMaxIdleTime(options.connMaxIdleLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("the database is not available: %w", err)
	}

	return db, nil
}
