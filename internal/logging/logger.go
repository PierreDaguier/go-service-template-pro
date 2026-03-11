package logging

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"

	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/config"
	"github.com/freelance-engineer/go-service-template-pro/internal/store"
)

func New(cfg config.Config, logStore *store.LogStore) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	level := parseLevel(cfg.LogLevel)
	zerolog.SetGlobalLevel(level)

	writers := []io.Writer{os.Stdout}
	if logStore != nil {
		writers = append(writers, logStore)
	}

	return zerolog.New(io.MultiWriter(writers...)).With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("environment", cfg.Environment).
		Logger()
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
