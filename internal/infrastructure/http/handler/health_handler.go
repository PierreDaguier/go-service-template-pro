package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/response"
)

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	checker ReadinessChecker
	logger  zerolog.Logger
}

func NewHealthHandler(checker ReadinessChecker, logger zerolog.Logger) *HealthHandler {
	return &HealthHandler{checker: checker, logger: logger}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{"status": "alive"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.checker.Ping(ctx); err != nil {
		h.logger.Error().Err(err).Msg("readiness check failed")
		response.JSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "reason": err.Error()})
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.Readiness(w, r)
}
