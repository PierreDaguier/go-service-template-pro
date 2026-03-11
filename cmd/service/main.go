package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/freelance-engineer/go-service-template-pro/internal/application"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/config"
	dbinfra "github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/db"
	httpinfra "github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/handler"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/middleware"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/repository"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/telemetry"
	"github.com/freelance-engineer/go-service-template-pro/internal/logging"
	"github.com/freelance-engineer/go-service-template-pro/internal/security"
	"github.com/freelance-engineer/go-service-template-pro/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	logStore := store.NewLogStore(4000)
	logger := logging.New(cfg, logStore)

	dbPool, err := dbinfra.Connect(ctx, cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("database connection failed")
	}
	defer dbPool.Close()

	if err := dbinfra.RunMigrations(ctx, dbPool); err != nil {
		logger.Fatal().Err(err).Msg("database migrations failed")
	}

	shutdownTracing, err := telemetry.InitTracing(ctx, cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("telemetry setup failed")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("failed to shutdown tracing")
		}
	}()

	incidentRepo := repository.NewIncidentRepository(dbPool)
	requestStore := store.NewRequestStore(5000)
	traceStore := store.NewTraceStore(2000)
	opsService := application.NewService(incidentRepo, requestStore, traceStore, logStore, cfg)
	if cfg.SeedDemoData {
		opsService.SeedDemoData()
	}

	authenticator, err := security.NewAuthenticator(cfg.APIKeys(), cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		logger.Fatal().Err(err).Msg("auth setup failed")
	}
	rateLimiter := security.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst, 10*time.Minute)
	httpMetrics := telemetry.NewHTTPMetrics()

	healthHandler := handler.NewHealthHandler(incidentRepo, logger)
	opsHandler := handler.NewOpsHandler(opsService, logger)
	apiRouter := httpinfra.NewRouter(httpinfra.RouterParams{
		Config:        cfg,
		Logger:        logger,
		HealthHandler: healthHandler,
		OpsHandler:    opsHandler,
		Auth:          middleware.NewAuth(authenticator),
		RateLimit:     middleware.NewRateLimit(rateLimiter),
		Metrics:       httpMetrics,
		RequestStore:  requestStore,
		TraceStore:    traceStore,
	})

	rootHandler := otelhttp.NewHandler(apiRouter, "http.server")
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:      rootHandler,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	go func() {
		logger.Info().Int("port", cfg.HTTPPort).Msg("service started")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("http server failed")
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-sigCtx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("server shutdown failed")
	}
	logger.Info().Msg("service stopped")
}
