package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/freelance-engineer/go-service-template-pro/internal/application"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/config"
	dbinfra "github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/db"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/repository"
	"github.com/freelance-engineer/go-service-template-pro/internal/logging"
	"github.com/freelance-engineer/go-service-template-pro/internal/store"
)

func newMux(service *application.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"alive"}`))
	})
	mux.HandleFunc("/api/v1/overview", func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"service_unavailable"}`))
			return
		}
		overview, err := service.Overview(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"overview_unavailable"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": overview})
	})
	return mux
}

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

	repo := repository.NewIncidentRepository(dbPool)
	service := application.NewService(repo, store.NewRequestStore(5000), store.NewTraceStore(2000), logStore, cfg)

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:      newMux(service),
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
