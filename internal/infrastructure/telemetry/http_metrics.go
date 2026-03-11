package telemetry

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPMetrics struct {
	registry        *prometheus.Registry
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	inFlight        prometheus.Gauge
}

func NewHTTPMetrics() *HTTPMetrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_http_requests_total",
			Help: "Total HTTP requests handled by the service.",
		},
		[]string{"route", "method", "status"},
	)
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_http_request_duration_ms",
			Help:    "HTTP request duration in milliseconds.",
			Buckets: []float64{5, 10, 25, 50, 75, 100, 150, 200, 300, 500, 1000, 2000},
		},
		[]string{"route", "method"},
	)
	inFlight := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "service_http_in_flight_requests",
		Help: "Current number of in-flight requests.",
	})

	registry.MustRegister(requestsTotal, requestDuration, inFlight)

	return &HTTPMetrics{
		registry:        registry,
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
		inFlight:        inFlight,
	}
}

func (m *HTTPMetrics) Observe(route, method string, statusCode int, durationMs float64) {
	status := strconv.Itoa(statusCode)
	m.requestsTotal.WithLabelValues(route, method, status).Inc()
	m.requestDuration.WithLabelValues(route, method).Observe(durationMs)
}

func (m *HTTPMetrics) IncInFlight() {
	m.inFlight.Inc()
}

func (m *HTTPMetrics) DecInFlight() {
	m.inFlight.Dec()
}

func (m *HTTPMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *HTTPMetrics) Summary() string {
	metricFamilies, _ := m.registry.Gather()
	return fmt.Sprintf("%d metrics exported", len(metricFamilies))
}
