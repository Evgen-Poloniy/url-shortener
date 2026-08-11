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

	"github.com/Evgen-Poloniy/url-shortener/internal/config"
	"github.com/Evgen-Poloniy/url-shortener/internal/repository/memory"
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

func Run(storageType string) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = config.DefaultConfigPath
	}

	if storageType == "" {
		storageType = config.MemoryStorageType
		logrus.Warn(`flag "storage-type" is empty. Is used "storage-memory"`)
	}

	cfg, err := config.LoadConfig(configPath, storageType)
	if err != nil {
		logrus.Fatalf("error when loading config: %v", err)
	}

	logger := logs.NewLogrusLogger(
		logs.WithLevel(cfg.Logger.Level),
		logs.WithFormat(cfg.Logger.Format),
	)

	// Layers initialization.

	var repository shortener.ShortenerRepository

	// Repository type choice.
	switch storageType {
	case config.PostgresStorageType:
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.Username,
			cfg.Postgres.Password,
			cfg.Postgres.DBName,
			cfg.Postgres.SSLMode,
		)

		db, err := postgres.NewPostgreSQL(
			dsn,
			postgres.WithMaxOpenConns(cfg.Postgres.MaxOpenConns),
			postgres.WithMaxIdleConns(cfg.Postgres.MaxIdleConns),
			postgres.WithConnMaxLifetime(cfg.Postgres.ConnMaxLifetime),
			postgres.WithConnMaxIdleLifetime(cfg.Postgres.ConnMaxIdleLifetime),
		)
		if err != nil {
			logger.Fatalf("database error: %v", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				logger.Errorf("database error: %v", err)
			}
		}()

		repository = pg.NewPostgresRepository(db)
		logger.Info("is used postgres repository type")

	case "memory":
		repository = memory.NewMemoryRepository()
		logger.Info("is used memory repository type")

	default:
		logger.Fatalln("unknown repository type")
	}

	service := shortener.NewShortenerService(repository)
	v1Handler := v1.NewHandler(service, &cfg.Auth)
	router := router.NewRouter(&cfg.CORS, logger)
	v1.NewRouter(router, v1Handler)

	httpServer := httpserver.NewServer(router,
		httpserver.WithAddr(cfg.Server.Host, cfg.Server.Port),
		httpserver.WithMaxHeaderBytes(cfg.Server.MaxHeaderBytes),
		httpserver.WithReadTimeout(cfg.Server.ReadTimeout),
		httpserver.WithWriteTimeout(cfg.Server.WriteTimeout),
		httpserver.WithReadHeaderTimeout(cfg.Server.ReadHeaderTimeout),
		httpserver.WithIdleTimeout(cfg.Server.IdleTimeout),
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Infof("http server is running on %s", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
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
		cfg.Server.TimeForGracefulShutdown,
	)
	defer cancel()

	logger.Info("shutting down server")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Errorf("server forced to shutdown: %v", err)
	}

	wg.Wait()

	logger.Info("server stop")
}
