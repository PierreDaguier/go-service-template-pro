package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/config"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/handler"
	custommw "github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/middleware"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/telemetry"
	"github.com/freelance-engineer/go-service-template-pro/internal/store"
)

type RouterParams struct {
	Config        config.Config
	Logger        zerolog.Logger
	HealthHandler *handler.HealthHandler
	OpsHandler    *handler.OpsHandler
	Auth          *custommw.Auth
	RateLimit     *custommw.RateLimit
	Metrics       *telemetry.HTTPMetrics
	RequestStore  *store.RequestStore
	TraceStore    *store.TraceStore
}

func NewRouter(params RouterParams) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(custommw.RequestID)
	r.Use(custommw.CORS(params.Config.Origins()))
	r.Use(custommw.Observe(params.Logger, params.Metrics, params.RequestStore, params.TraceStore))

	r.Get("/health/live", params.HealthHandler.Liveness)
	r.Get("/health/ready", params.HealthHandler.Readiness)
	r.Get("/health", params.HealthHandler.Health)
	r.Handle("/metrics", params.Metrics.Handler())

	r.Route("/api/v1", func(api chi.Router) {
		api.Use(params.Auth.Handler)
		api.Use(params.RateLimit.Handler)
		api.Get("/overview", params.OpsHandler.GetOverview)
		api.Get("/metrics/live", params.OpsHandler.GetLiveMetrics)
		api.Get("/errors", params.OpsHandler.ListErrors)
		api.Post("/errors", params.OpsHandler.CreateError)
		api.Get("/traces", params.OpsHandler.ListTraces)
		api.Get("/config", params.OpsHandler.GetConfig)
		api.Get("/logs", params.OpsHandler.ListLogs)
	})

	return r
}
