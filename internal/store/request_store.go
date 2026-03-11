package store

import (
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

type RequestRecord struct {
	Timestamp  time.Time
	Method     string
	Route      string
	StatusCode int
	DurationMs float64
	TraceID    string
}

type TimePoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Throughput int       `json:"throughput"`
	ErrorRate  float64   `json:"errorRate"`
	P95Latency float64   `json:"p95LatencyMs"`
}

type StatusCount struct {
	Code  int `json:"code"`
	Count int `json:"count"`
}

type RouteCount struct {
	Route string `json:"route"`
	Count int    `json:"count"`
}

type MetricsSnapshot struct {
	WindowSeconds int           `json:"windowSeconds"`
	Requests      int           `json:"requests"`
	RPS           float64       `json:"rps"`
	ErrorRate     float64       `json:"errorRate"`
	P95LatencyMs  float64       `json:"p95LatencyMs"`
	Series        []TimePoint   `json:"series"`
	StatusCodes   []StatusCount `json:"statusCodes"`
	TopRoutes     []RouteCount  `json:"topRoutes"`
}

type RequestStore struct {
	mu      sync.RWMutex
	records []RequestRecord
	max     int
}

func NewRequestStore(max int) *RequestStore {
	if max <= 0 {
		max = 5000
	}
	return &RequestStore{max: max, records: make([]RequestRecord, 0, max)}
}

func (s *RequestStore) Add(record RequestRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.records) >= s.max {
		copy(s.records, s.records[1:])
		s.records[len(s.records)-1] = record
		return
	}
	s.records = append(s.records, record)
}

func (s *RequestStore) Snapshot(window time.Duration) MetricsSnapshot {
	if window <= 0 {
		window = 15 * time.Minute
	}
	cutoff := time.Now().Add(-window)

	s.mu.RLock()
	records := make([]RequestRecord, 0, len(s.records))
	for _, item := range s.records {
		if item.Timestamp.After(cutoff) {
			records = append(records, item)
		}
	}
	s.mu.RUnlock()

	windowSeconds := int(window.Seconds())
	if windowSeconds <= 0 {
		windowSeconds = 1
	}

	snapshot := MetricsSnapshot{WindowSeconds: windowSeconds, Requests: len(records)}
	if len(records) == 0 {
		return snapshot
	}

	durations := make([]float64, 0, len(records))
	statusMap := map[int]int{}
	routeMap := map[string]int{}
	bucketMap := map[time.Time][]RequestRecord{}
	errorCount := 0

	for _, item := range records {
		durations = append(durations, item.DurationMs)
		statusMap[item.StatusCode]++
		routeMap[item.Route]++
		if item.StatusCode >= 500 {
			errorCount++
		}
		bucket := item.Timestamp.Truncate(time.Minute)
		bucketMap[bucket] = append(bucketMap[bucket], item)
	}

	snapshot.RPS = float64(len(records)) / float64(windowSeconds)
	snapshot.ErrorRate = float64(errorCount) / float64(len(records))
	snapshot.P95LatencyMs = percentile(durations, 95)

	snapshot.StatusCodes = make([]StatusCount, 0, len(statusMap))
	for code, count := range statusMap {
		snapshot.StatusCodes = append(snapshot.StatusCodes, StatusCount{Code: code, Count: count})
	}
	sort.Slice(snapshot.StatusCodes, func(i, j int) bool { return snapshot.StatusCodes[i].Code < snapshot.StatusCodes[j].Code })

	topRoutes := make([]RouteCount, 0, len(routeMap))
	for route, count := range routeMap {
		topRoutes = append(topRoutes, RouteCount{Route: route, Count: count})
	}
	sort.Slice(topRoutes, func(i, j int) bool { return topRoutes[i].Count > topRoutes[j].Count })
	if len(topRoutes) > 6 {
		topRoutes = topRoutes[:6]
	}
	snapshot.TopRoutes = topRoutes

	bucketKeys := make([]time.Time, 0, len(bucketMap))
	for key := range bucketMap {
		bucketKeys = append(bucketKeys, key)
	}
	sort.Slice(bucketKeys, func(i, j int) bool { return bucketKeys[i].Before(bucketKeys[j]) })

	series := make([]TimePoint, 0, len(bucketKeys))
	for _, key := range bucketKeys {
		bucketRecords := bucketMap[key]
		errorN := 0
		latencies := make([]float64, 0, len(bucketRecords))
		for _, item := range bucketRecords {
			if item.StatusCode >= 500 {
				errorN++
			}
			latencies = append(latencies, item.DurationMs)
		}
		series = append(series, TimePoint{
			Timestamp:  key,
			Throughput: len(bucketRecords),
			ErrorRate:  float64(errorN) / float64(len(bucketRecords)),
			P95Latency: percentile(latencies, 95),
		})
	}
	snapshot.Series = series

	return snapshot
}

func (s *RequestStore) Seed(records []RequestRecord) {
	for _, record := range records {
		s.Add(record)
	}
}

func percentile(values []float64, pct float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sort.Float64s(values)
	idx := int(math.Ceil((pct/100)*float64(len(values)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(values) {
		idx = len(values) - 1
	}
	return values[idx]
}

func NormalizeRoute(method, route string) string {
	trimmed := strings.TrimSpace(route)
	if trimmed == "" {
		trimmed = "/unknown"
	}
	return strings.ToUpper(strings.TrimSpace(method)) + " " + trimmed
}
