package application

import (
	"context"
	"testing"
	"time"

	"github.com/freelance-engineer/go-service-template-pro/internal/domain"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/config"
	"github.com/freelance-engineer/go-service-template-pro/internal/store"
)

type mockRepo struct {
	incidents []domain.Incident
	counts    domain.IncidentCounts
}

func (m *mockRepo) CreateIncident(_ context.Context, incident domain.Incident) error {
	m.incidents = append(m.incidents, incident)
	return nil
}

func (m *mockRepo) ListIncidents(_ context.Context, _ domain.IncidentFilters) ([]domain.Incident, int, error) {
	return m.incidents, len(m.incidents), nil
}

func (m *mockRepo) GetOpenCounts(_ context.Context) (domain.IncidentCounts, error) {
	return m.counts, nil
}

func (m *mockRepo) Ping(_ context.Context) error { return nil }

func TestOverviewStatusDegradedWithCriticalIncident(t *testing.T) {
	repo := &mockRepo{counts: domain.IncidentCounts{OpenTotal: 4, CriticalOpen: 1}}
	service := NewService(repo, store.NewRequestStore(100), store.NewTraceStore(100), store.NewLogStore(100), config.Config{
		ServiceName:    "svc",
		Environment:    "test",
		ServiceVersion: "0.1.0",
	})

	service.requests.Add(store.RequestRecord{Timestamp: time.Now().UTC(), StatusCode: 200, DurationMs: 10})
	overview, err := service.Overview(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if overview.SystemStatus != "degraded" {
		t.Fatalf("expected degraded status, got %s", overview.SystemStatus)
	}
}

func TestCreateIncidentDefaults(t *testing.T) {
	repo := &mockRepo{}
	service := NewService(repo, store.NewRequestStore(100), store.NewTraceStore(100), store.NewLogStore(100), config.Config{
		ServiceName: "svc",
		Environment: "staging",
	})
	incident, err := service.CreateIncident(context.Background(), CreateIncidentInput{
		Severity: domain.SeverityHigh,
		Status:   domain.StatusOpen,
		Message:  "database response time elevated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if incident.Service != "svc" {
		t.Fatalf("expected default service name, got %s", incident.Service)
	}
	if incident.Environment != "staging" {
		t.Fatalf("expected default environment, got %s", incident.Environment)
	}
}
