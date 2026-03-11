package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/freelance-engineer/go-service-template-pro/internal/domain"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/config"
	"github.com/freelance-engineer/go-service-template-pro/internal/store"
)

type Service struct {
	repo      domain.IncidentRepository
	requests  *store.RequestStore
	traces    *store.TraceStore
	logs      *store.LogStore
	cfg       config.Config
	startedAt time.Time
	now       func() time.Time
}

type CreateIncidentInput struct {
	Service     string
	Environment string
	Severity    domain.Severity
	Status      domain.IncidentStatus
	Message     string
	TraceID     string
	Metadata    map[string]any
}

type OverviewResponse struct {
	ServiceName       string                `json:"serviceName"`
	Environment       string                `json:"environment"`
	Version           string                `json:"version"`
	UptimeSeconds     int64                 `json:"uptimeSeconds"`
	OpenIncidents     int                   `json:"openIncidents"`
	CriticalIncidents int                   `json:"criticalIncidents"`
	Traffic           store.MetricsSnapshot `json:"traffic"`
	SystemStatus      string                `json:"systemStatus"`
}

type ConfigStatusResponse struct {
	Service struct {
		Name        string `json:"name"`
		Environment string `json:"environment"`
		Version     string `json:"version"`
	} `json:"service"`
	Auth struct {
		APIKeyEnabled bool `json:"apiKeyEnabled"`
		JWTEnabled    bool `json:"jwtEnabled"`
	} `json:"auth"`
	Telemetry struct {
		OTLPEndpoint string `json:"otlpEndpoint"`
	} `json:"telemetry"`
	Database struct {
		Configured bool `json:"configured"`
	} `json:"database"`
	RateLimit struct {
		RequestsPerSecond float64 `json:"requestsPerSecond"`
		Burst             int     `json:"burst"`
	} `json:"rateLimit"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

func NewService(
	repo domain.IncidentRepository,
	requests *store.RequestStore,
	traces *store.TraceStore,
	logs *store.LogStore,
	cfg config.Config,
) *Service {
	return &Service{
		repo:      repo,
		requests:  requests,
		traces:    traces,
		logs:      logs,
		cfg:       cfg,
		startedAt: time.Now().UTC(),
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) CreateIncident(ctx context.Context, input CreateIncidentInput) (domain.Incident, error) {
	incident := domain.Incident{
		ID:          uuid.NewString(),
		Service:     input.Service,
		Environment: input.Environment,
		Severity:    input.Severity,
		Status:      input.Status,
		Message:     input.Message,
		TraceID:     input.TraceID,
		Metadata:    input.Metadata,
		CreatedAt:   s.now(),
	}

	if incident.Service == "" {
		incident.Service = s.cfg.ServiceName
	}
	if incident.Environment == "" {
		incident.Environment = s.cfg.Environment
	}

	if err := s.repo.CreateIncident(ctx, incident); err != nil {
		return domain.Incident{}, fmt.Errorf("create incident: %w", err)
	}
	return incident, nil
}

func (s *Service) ListIncidents(ctx context.Context, filters domain.IncidentFilters) ([]domain.Incident, int, error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 || filters.PageSize > 100 {
		filters.PageSize = 20
	}
	return s.repo.ListIncidents(ctx, filters)
}

func (s *Service) Overview(ctx context.Context) (OverviewResponse, error) {
	counts, err := s.repo.GetOpenCounts(ctx)
	if err != nil {
		return OverviewResponse{}, fmt.Errorf("get open counts: %w", err)
	}
	traffic := s.requests.Snapshot(15 * time.Minute)
	overview := OverviewResponse{
		ServiceName:       s.cfg.ServiceName,
		Environment:       s.cfg.Environment,
		Version:           s.cfg.ServiceVersion,
		UptimeSeconds:     int64(s.now().Sub(s.startedAt).Seconds()),
		OpenIncidents:     counts.OpenTotal,
		CriticalIncidents: counts.CriticalOpen,
		Traffic:           traffic,
		SystemStatus:      "healthy",
	}
	if counts.CriticalOpen > 0 || traffic.ErrorRate > 0.05 {
		overview.SystemStatus = "degraded"
	}
	return overview, nil
}

func (s *Service) LiveMetrics(window time.Duration) store.MetricsSnapshot {
	return s.requests.Snapshot(window)
}

func (s *Service) ListTraces(limit int) []store.TraceRecord {
	return s.traces.List(limit)
}

func (s *Service) ListLogs(level string, limit int) []store.LogEntry {
	return s.logs.List(limit, level)
}

func (s *Service) ConfigStatus() ConfigStatusResponse {
	status := ConfigStatusResponse{}
	status.Service.Name = s.cfg.ServiceName
	status.Service.Environment = s.cfg.Environment
	status.Service.Version = s.cfg.ServiceVersion
	status.Auth.APIKeyEnabled = len(s.cfg.APIKeys()) > 0
	status.Auth.JWTEnabled = s.cfg.JWTSecret != ""
	status.Telemetry.OTLPEndpoint = s.cfg.OTLPEndpoint
	status.Database.Configured = s.cfg.DatabaseURL != ""
	status.RateLimit.RequestsPerSecond = s.cfg.RateLimitRPS
	status.RateLimit.Burst = s.cfg.RateLimitBurst
	status.AllowedOrigins = s.cfg.Origins()
	return status
}

func (s *Service) SeedDemoData() {
	now := s.now()
	for i := 30; i >= 1; i-- {
		ts := now.Add(-time.Duration(i) * time.Minute)
		status := 200
		duration := 25 + float64(i%9)*7.2
		route := "/api/v1/overview"
		if i%5 == 0 {
			route = "/api/v1/errors"
		}
		if i%9 == 0 {
			status = 502
			duration = 210 + float64(i)
		}
		traceID := uuid.NewString()
		s.requests.Add(store.RequestRecord{
			Timestamp:  ts,
			Method:     "GET",
			Route:      store.NormalizeRoute("GET", route),
			StatusCode: status,
			DurationMs: duration,
			TraceID:    traceID,
		})
		s.traces.Add(store.TraceRecord{
			Timestamp:  ts,
			TraceID:    traceID,
			Method:     "GET",
			Route:      route,
			StatusCode: status,
			DurationMs: duration,
			Source:     "demo_seed",
		})
	}
}
