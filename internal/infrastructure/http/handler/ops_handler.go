package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/freelance-engineer/go-service-template-pro/internal/application"
	"github.com/freelance-engineer/go-service-template-pro/internal/domain"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/middleware"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/response"
)

type OpsHandler struct {
	service  *application.Service
	validate *validator.Validate
	logger   zerolog.Logger
}

type createIncidentRequest struct {
	Service     string                `json:"service" validate:"omitempty,min=2,max=64"`
	Environment string                `json:"environment" validate:"omitempty,min=2,max=32"`
	Severity    domain.Severity       `json:"severity" validate:"required,oneof=critical high medium low"`
	Status      domain.IncidentStatus `json:"status" validate:"required,oneof=open investigating resolved"`
	Message     string                `json:"message" validate:"required,min=8,max=320"`
	TraceID     string                `json:"traceId" validate:"omitempty,max=128"`
	Metadata    map[string]any        `json:"metadata"`
}

func NewOpsHandler(service *application.Service, logger zerolog.Logger) *OpsHandler {
	return &OpsHandler{
		service:  service,
		validate: validator.New(validator.WithRequiredStructEnabled()),
		logger:   logger,
	}
}

func (h *OpsHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.Overview(r.Context())
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": overview})
}

func (h *OpsHandler) GetLiveMetrics(w http.ResponseWriter, r *http.Request) {
	window := 15 * time.Minute
	if raw := strings.TrimSpace(r.URL.Query().Get("window")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 || parsed > 24*time.Hour {
			response.Error(w, http.StatusBadRequest, response.APIError{
				Code:      "invalid_window",
				Message:   "Query parameter 'window' must be a valid duration like 15m",
				RequestID: middleware.RequestIDFromContext(r.Context()),
			})
			return
		}
		window = parsed
	}
	snapshot := h.service.LiveMetrics(window)
	response.JSON(w, http.StatusOK, map[string]any{"data": snapshot})
}

func (h *OpsHandler) ListErrors(w http.ResponseWriter, r *http.Request) {
	filters := domain.IncidentFilters{
		Severity: strings.TrimSpace(r.URL.Query().Get("severity")),
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
		Query:    strings.TrimSpace(r.URL.Query().Get("q")),
		Page:     parseInt(r.URL.Query().Get("page"), 1),
		PageSize: parseInt(r.URL.Query().Get("page_size"), 20),
	}

	incidents, total, err := h.service.ListIncidents(r.Context(), filters)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"data": incidents,
		"pagination": map[string]int{
			"page":       filters.Page,
			"pageSize":   filters.PageSize,
			"total":      total,
			"totalPages": totalPages(total, filters.PageSize),
		},
	})
}

func (h *OpsHandler) CreateError(w http.ResponseWriter, r *http.Request) {
	var req createIncidentRequest
	if err := decodeStrict(r.Body, &req); err != nil {
		response.Error(w, http.StatusBadRequest, response.APIError{
			Code:      "invalid_payload",
			Message:   "Malformed JSON payload",
			Details:   err.Error(),
			RequestID: middleware.RequestIDFromContext(r.Context()),
		})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, response.APIError{
			Code:      "validation_error",
			Message:   "Payload validation failed",
			Details:   err.Error(),
			RequestID: middleware.RequestIDFromContext(r.Context()),
		})
		return
	}

	incident, err := h.service.CreateIncident(r.Context(), application.CreateIncidentInput{
		Service:     req.Service,
		Environment: req.Environment,
		Severity:    req.Severity,
		Status:      req.Status,
		Message:     req.Message,
		TraceID:     req.TraceID,
		Metadata:    req.Metadata,
	})
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"data": incident})
}

func (h *OpsHandler) ListTraces(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 80)
	if limit > 500 {
		limit = 500
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": h.service.ListTraces(limit)})
}

func (h *OpsHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{"data": h.service.ConfigStatus()})
}

func (h *OpsHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 120)
	if limit > 500 {
		limit = 500
	}
	level := strings.TrimSpace(r.URL.Query().Get("level"))
	response.JSON(w, http.StatusOK, map[string]any{"data": h.service.ListLogs(level, limit)})
}

func (h *OpsHandler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.Error().Err(err).Str("request_id", middleware.RequestIDFromContext(r.Context())).Msg("request failed")
	status, apiErr := response.FromDomain(err, middleware.RequestIDFromContext(r.Context()))
	response.Error(w, status, apiErr)
}

func decodeStrict(body io.ReadCloser, target any) error {
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.More() {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func parseInt(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func totalPages(total, pageSize int) int {
	if total == 0 || pageSize <= 0 {
		return 0
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	return pages
}
