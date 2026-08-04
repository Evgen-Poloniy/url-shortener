package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Evgen-Poloniy/url-shortener/internal/config"
	pg "github.com/Evgen-Poloniy/url-shortener/internal/repository/postgres"
	httpserver "github.com/Evgen-Poloniy/url-shortener/internal/server/http"
	"github.com/Evgen-Poloniy/url-shortener/internal/service/shortener"
	router "github.com/Evgen-Poloniy/url-shortener/internal/transport/http"
	v1 "github.com/Evgen-Poloniy/url-shortener/internal/transport/http/v1"
	"github.com/Evgen-Poloniy/url-shortener/pkg/database/postgres"
	"github.com/Evgen-Poloniy/url-shortener/pkg/logs"
	"github.com/sirupsen/logrus"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		logrus.Fatalf("error when loading env CONFIG_PATH")
	}

	config, err := config.LoadConfig(configPath)
	if err != nil {
		logrus.Fatalf("error when loading config: %v", err)
	}

	logger := logs.NewLogrusLogger(
		logs.WithLevel(config.Logger.Level),
		logs.WithFormat(config.Logger.Format),
	)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.Username,
		config.Postgres.Password,
		config.Postgres.DBName,
		config.Postgres.SSLMode,
	)

	db, err := postgres.NewPostgreSQL(
		dsn,
		postgres.WithMaxOpenConns(config.Postgres.MaxOpenConns),
		postgres.WithMaxIdleConns(config.Postgres.MaxIdleConns),
		postgres.WithConnMaxLifetime(config.Postgres.ConnMaxLifetime),
		postgres.WithConnMaxIdleLifetime(config.Postgres.ConnMaxIdleLifetime),
	)
	if err != nil {
		logger.Fatalf("database error: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("database error: %v", err)
		}
	}()

	// Layers initialization.
	repository := pg.NewPostgresRepository(db)
	service := shortener.NewShortenerService(repository)
	v1Handler := v1.NewHandler(service, &config.Auth)
	router := router.NewRouter(&config.CORS, logger)
	v1.NewRouter(router, v1Handler)

	httpServer := httpserver.NewServer(router,
		httpserver.WithAddr(config.Server.Host, config.Server.Port),
		httpserver.WithMaxHeaderBytes(config.Server.MaxHeaderBytes),
		httpserver.WithReadTimeout(config.Server.ReadTimeout),
		httpserver.WithWriteTimeout(config.Server.WriteTimeout),
		httpserver.WithReadHeaderTimeout(config.Server.ReadHeaderTimeout),
		httpserver.WithIdleTimeout(config.Server.IdleTimeout),
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Infof("http server is running on %s", fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port))
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("http server error: %v", err)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(
		context.Background(),
		config.Server.TimeForGracefulShutdown*time.Second,
	)
	defer cancel()

	logger.Info("shutting down server")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Errorf("server forced to shutdown: %v", err)
	}

	wg.Wait()

	logger.Info("server stop")
}
