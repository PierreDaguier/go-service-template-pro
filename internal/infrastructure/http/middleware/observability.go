package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/telemetry"
	"github.com/freelance-engineer/go-service-template-pro/internal/store"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func Observe(logger zerolog.Logger, metrics *telemetry.HTTPMetrics, requestStore *store.RequestStore, traceStore *store.TraceStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.IncInFlight()
			defer metrics.DecInFlight()

			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(recorder, r)

			routePattern := "unknown"
			if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
				if pattern := routeCtx.RoutePattern(); pattern != "" {
					routePattern = pattern
				}
			}
			normalizedRoute := store.NormalizeRoute(r.Method, routePattern)
			durationMs := float64(time.Since(start).Milliseconds())
			if durationMs == 0 {
				durationMs = 1
			}

			spanCtx := trace.SpanContextFromContext(r.Context())
			traceID := ""
			if spanCtx.HasTraceID() {
				traceID = spanCtx.TraceID().String()
			}

			metrics.Observe(routePattern, r.Method, recorder.statusCode, durationMs)
			requestStore.Add(store.RequestRecord{
				Timestamp:  start.UTC(),
				Method:     r.Method,
				Route:      normalizedRoute,
				StatusCode: recorder.statusCode,
				DurationMs: durationMs,
				TraceID:    traceID,
			})
			traceStore.Add(store.TraceRecord{
				Timestamp:  start.UTC(),
				TraceID:    traceID,
				Method:     r.Method,
				Route:      routePattern,
				StatusCode: recorder.statusCode,
				DurationMs: durationMs,
				Source:     "http_server",
			})

			logEvent := logger.Info()
			if recorder.statusCode >= 500 {
				logEvent = logger.Error()
			} else if recorder.statusCode >= 400 {
				logEvent = logger.Warn()
			}
			logEvent.
				Str("method", r.Method).
				Str("route", routePattern).
				Int("status", recorder.statusCode).
				Float64("duration_ms", durationMs).
				Str("trace_id", traceID).
				Str("request_id", RequestIDFromContext(r.Context())).
				Msg("request completed")
		})
	}
}
