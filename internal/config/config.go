package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath   = "configs/config.yml"
	PostgresStorageType = "postgres"
	MemoryStorageType   = "memory"
)

// Config with tags from cleanenv library.
type ServerConfig struct {
	Host                    string        `env:"API_HOST" env-required:"true"`
	Port                    int           `env:"API_PORT" env-required:"true"`
	MaxHeaderBytes          int           `yaml:"max_header_bytes" env-default:"1048576"`
	ReadTimeout             time.Duration `yaml:"read_timeout" env-default:"4s"`
	WriteTimeout            time.Duration `yaml:"write_timeout" env-default:"10s"`
	TimeForGracefulShutdown time.Duration `yaml:"time_for_graceful_shutdown" env-default:"10s"`
	ReadHeaderTimeout       time.Duration `yaml:"read_header_timeout" env-default:"2s"`
	IdleTimeout             time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// Logger config from config.yml.
type LoggerConfig struct {
	Level  string `yaml:"level" env-default:"info" validate:"oneof=trace debug info warn error panic fatal"`
	Format string `yaml:"format" env-default:"json" validate:"oneof=text json"`
}

// CORSConfig is the config for CORS policy.
type CORSConfig struct {
	AllowedOrigin    string        `yaml:"allowed_origin"`
	AllowCredentials bool          `yaml:"allow_credentials"`
	AllowedHeaders   []string      `yaml:"allowed_headers"`
	AllowedMethods   []string      `yaml:"allowed_methods"`
	MaxAge           time.Duration `yaml:"max_age" validate:"required,gte=1m,lte=24h"`
}

// PostgresConfig represents config from env and config.yaml.
type PostgresConfig struct {
	Host                string        `env:"DB_HOST" validate:"required"`
	Port                string        `env:"DB_PORT" validate:"required"`
	Username            string        `env:"DB_USER" validate:"required"`
	Password            string        `env:"DB_PASSWORD" validate:"required"`
	DBName              string        `env:"DB_NAME" validate:"required"`
	SSLMode             string        `env:"SSL_MODE" validate:"required,oneof=disable require"`
	MaxOpenConns        int           `yaml:"max_open_conns" env-default:"25"`
	MaxIdleConns        int           `yaml:"max_idle_conns" env-default:"25"`
	ConnMaxLifetime     time.Duration `yaml:"conn_max_lifetime" env-default:"5m"`
	ConnMaxIdleLifetime time.Duration `yaml:"conn_max_idle_lifetime" env-default:"2m"`
}

// AuthConfig represents authentication config.
type AuthConfig struct {
	ApiKey string `env:"API_KEY" env-required:"true"`
}

// Config represents dataclass with all configs.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Logger   LoggerConfig   `yaml:"logger"`
	CORS     CORSConfig     `yaml:"cors"`
	Postgres PostgresConfig `yaml:"postgres"`
	Auth     AuthConfig     `yaml:"auth"`
}

// Load config from config/config.yml.
func LoadConfig(path string, storageType string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}

	validate := validator.New()

	if err := validate.Struct(cfg.Logger); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}
	if err := validate.Struct(cfg.CORS); err != nil {
		return nil, fmt.Errorf("invalid CORS config:  %w", err)
	}

	if storageType == PostgresStorageType {
		if err := validate.Struct(cfg.Postgres); err != nil {
			return nil, fmt.Errorf("invalid postgres config: %w", err)
		}
	}

	return &cfg, nil
}
