package postgres

import "time"

type Options struct {
	maxOpenConns        int
	maxIdleConns        int
	connMaxLifetime     time.Duration
	connMaxIdleLifetime time.Duration
}

type Option func(*Options)

// WithMaxOpenConns sets max opened connections with the database
func WithMaxOpenConns(n int) Option {
	return func(o *Options) {
		o.maxOpenConns = n
	}
}

// WithMaxIdleConns sets max idle connections with the database
func WithMaxIdleConns(n int) Option {
	return func(o *Options) {
		o.maxIdleConns = n
	}
}

// WithConnMaxLifetime sets max connections time with the database
func WithConnMaxLifetime(d time.Duration) Option {
	return func(o *Options) {
		o.connMaxLifetime = d
	}
}

// WithConnMaxIdleLifetime sets max connections idle time with the database
func WithConnMaxIdleLifetime(d time.Duration) Option {
	return func(o *Options) {
		o.connMaxIdleLifetime = d
	}
}
