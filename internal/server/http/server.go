package httpserver

import (
	"context"
	"net/http"
	"time"
)

// Wrapper for http server
type Server struct {
	httpServer *http.Server
}

// Create new http server
func NewServer(handler http.Handler, opts ...Option) *Server {
	options := Options{
		Address:           "0.0.0.0:8080",
		MaxHeaderBytes:    1 << 20,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              options.Address,
			MaxHeaderBytes:    options.MaxHeaderBytes,
			Handler:           handler,
			ReadTimeout:       options.ReadTimeout,
			WriteTimeout:      options.WriteTimeout,
			ReadHeaderTimeout: options.ReadHeaderTimeout,
			IdleTimeout:       options.IdleTimeout,
		},
	}
}

// Launch http server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown http server with context
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
