package domain

import "context"

type IncidentRepository interface {
	CreateIncident(ctx context.Context, incident Incident) error
	ListIncidents(ctx context.Context, filters IncidentFilters) ([]Incident, int, error)
	GetOpenCounts(ctx context.Context) (IncidentCounts, error)
	Ping(ctx context.Context) error
}
