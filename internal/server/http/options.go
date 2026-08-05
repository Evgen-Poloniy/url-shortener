package httpserver

import (
	"fmt"
	"time"
)

// Options holds configuration parameters for the HTTP server.
type Options struct {
	Address           string
	MaxHeaderBytes    int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
}

// Option defines a function for configuring Options.
type Option func(*Options)

// WithAddr configures the network address (host and port).
func WithAddr(host string, port int) Option {
	return func(o *Options) {
		o.Address = fmt.Sprintf("%s:%d", host, port)
	}
}

// WithReadTimeout sets the maximum duration for reading the entire request.
func WithReadTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the maximum duration before timing out writes of the response.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = timeout
	}
}

// WithReadHeaderTimeout sets the amount of time allowed to read request headers.
func WithReadHeaderTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadHeaderTimeout = timeout
	}
}

// WithIdleTimeout sets the maximum amount of time to wait for the next request.
func WithIdleTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.IdleTimeout = timeout
	}
}

// WithMaxHeaderBytes sets the maximum number of bytes the server will read parsing request headers.
func WithMaxHeaderBytes(maxBytes int) Option {
	return func(o *Options) {
		o.MaxHeaderBytes = maxBytes
	}
}
