package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains all runtime configuration loaded from environment variables.
type Config struct {
	ServiceName      string        `env:"SERVICE_NAME" envDefault:"go-service-template-pro"`
	Environment      string        `env:"APP_ENV" envDefault:"local"`
	ServiceVersion   string        `env:"SERVICE_VERSION" envDefault:"0.1.0"`
	HTTPPort         int           `env:"HTTP_PORT" envDefault:"8080"`
	HTTPReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"5s"`
	HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"10s"`
	HTTPIdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout  time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10s"`

	DatabaseURL      string `env:"DATABASE_URL" envDefault:"postgres://app:app@localhost:5432/go_service_template_pro?sslmode=disable"`
	DatabaseMaxConns int32  `env:"DATABASE_MAX_CONNS" envDefault:"20"`
	DatabaseMinConns int32  `env:"DATABASE_MIN_CONNS" envDefault:"2"`

	AuthAPIKeys string `env:"AUTH_API_KEYS" envDefault:"client-demo-key-2026"`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"change-me-in-production"`
	JWTIssuer   string `env:"JWT_ISSUER" envDefault:"go-service-template-pro"`
	JWTAudience string `env:"JWT_AUDIENCE" envDefault:"ops-clients"`

	RateLimitRPS   float64 `env:"RATE_LIMIT_RPS" envDefault:"12"`
	RateLimitBurst int     `env:"RATE_LIMIT_BURST" envDefault:"24"`

	OTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"otel-collector:4317"`
	OTLPInsecure bool   `env:"OTEL_EXPORTER_OTLP_INSECURE" envDefault:"true"`

	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	AllowedOrigins string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:5173,http://localhost:4173"`
	SeedDemoData   bool   `env:"SEED_DEMO_DATA" envDefault:"true"`
}

func Load() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RateLimitRPS <= 0 {
		return Config{}, fmt.Errorf("RATE_LIMIT_RPS must be > 0")
	}
	if cfg.RateLimitBurst <= 0 {
		return Config{}, fmt.Errorf("RATE_LIMIT_BURST must be > 0")
	}
	return cfg, nil
}

func (c Config) APIKeys() []string {
	parts := strings.Split(c.AuthAPIKeys, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			keys = append(keys, trimmed)
		}
	}
	return keys
}

func (c Config) Origins() []string {
	parts := strings.Split(c.AllowedOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
