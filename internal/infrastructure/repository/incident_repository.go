package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/freelance-engineer/go-service-template-pro/internal/domain"
)

type IncidentRepository struct {
	pool *pgxpool.Pool
}

func NewIncidentRepository(pool *pgxpool.Pool) *IncidentRepository {
	return &IncidentRepository{pool: pool}
}

func (r *IncidentRepository) CreateIncident(ctx context.Context, incident domain.Incident) error {
	meta, err := json.Marshal(incident.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO incidents (
			id, service, environment, severity, status, message, trace_id, metadata, created_at, resolved_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		incident.ID,
		incident.Service,
		incident.Environment,
		string(incident.Severity),
		string(incident.Status),
		incident.Message,
		incident.TraceID,
		meta,
		incident.CreatedAt,
		incident.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("insert incident: %w", err)
	}
	return nil
}

func (r *IncidentRepository) ListIncidents(ctx context.Context, filters domain.IncidentFilters) ([]domain.Incident, int, error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}

	where, args := buildWhere(filters)
	query := `
		SELECT id, service, environment, severity, status, message, trace_id, metadata, created_at, resolved_at
		FROM incidents
	` + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, filters.PageSize, (filters.Page-1)*filters.PageSize)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query incidents: %w", err)
	}
	defer rows.Close()

	incidents := make([]domain.Incident, 0, filters.PageSize)
	for rows.Next() {
		var (
			incident domain.Incident
			metaRaw  []byte
		)
		if err := rows.Scan(
			&incident.ID,
			&incident.Service,
			&incident.Environment,
			&incident.Severity,
			&incident.Status,
			&incident.Message,
			&incident.TraceID,
			&metaRaw,
			&incident.CreatedAt,
			&incident.ResolvedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan incident: %w", err)
		}
		incident.Metadata = map[string]any{}
		if len(metaRaw) > 0 {
			_ = json.Unmarshal(metaRaw, &incident.Metadata)
		}
		incidents = append(incidents, incident)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate incidents: %w", err)
	}

	countQuery := "SELECT COUNT(*) FROM incidents " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count incidents: %w", err)
	}

	return incidents, total, nil
}

func (r *IncidentRepository) GetOpenCounts(ctx context.Context) (domain.IncidentCounts, error) {
	counts := domain.IncidentCounts{}
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status IN ('open', 'investigating')),
			COUNT(*) FILTER (WHERE status IN ('open', 'investigating') AND severity = 'critical'),
			COUNT(*) FILTER (WHERE status IN ('open', 'investigating') AND severity = 'high')
		FROM incidents
	`
	if err := r.pool.QueryRow(ctx, query).Scan(&counts.OpenTotal, &counts.CriticalOpen, &counts.HighOpen); err != nil {
		return domain.IncidentCounts{}, fmt.Errorf("query open counts: %w", err)
	}
	return counts, nil
}

func (r *IncidentRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func buildWhere(filters domain.IncidentFilters) (string, []any) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)

	if severity := strings.TrimSpace(filters.Severity); severity != "" {
		args = append(args, severity)
		clauses = append(clauses, fmt.Sprintf("severity = $%d", len(args)))
	}
	if status := strings.TrimSpace(filters.Status); status != "" {
		args = append(args, status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}
	if query := strings.TrimSpace(filters.Query); query != "" {
		args = append(args, "%"+query+"%")
		clauses = append(clauses, fmt.Sprintf("(message ILIKE $%d OR service ILIKE $%d)", len(args), len(args)))
	}

	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}
