package domain

import "time"

type Severity string

type IncidentStatus string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

const (
	StatusOpen          IncidentStatus = "open"
	StatusInvestigating IncidentStatus = "investigating"
	StatusResolved      IncidentStatus = "resolved"
)

type Incident struct {
	ID          string         `json:"id"`
	Service     string         `json:"service"`
	Environment string         `json:"environment"`
	Severity    Severity       `json:"severity"`
	Status      IncidentStatus `json:"status"`
	Message     string         `json:"message"`
	TraceID     string         `json:"traceId,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	ResolvedAt  *time.Time     `json:"resolvedAt,omitempty"`
}

type IncidentFilters struct {
	Severity string
	Status   string
	Query    string
	Page     int
	PageSize int
}

type IncidentCounts struct {
	OpenTotal    int `json:"openTotal"`
	CriticalOpen int `json:"criticalOpen"`
	HighOpen     int `json:"highOpen"`
}
