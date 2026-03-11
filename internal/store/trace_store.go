package store

import (
	"sort"
	"sync"
	"time"
)

type TraceRecord struct {
	Timestamp  time.Time `json:"timestamp"`
	TraceID    string    `json:"traceId"`
	Method     string    `json:"method"`
	Route      string    `json:"route"`
	StatusCode int       `json:"statusCode"`
	DurationMs float64   `json:"durationMs"`
	Source     string    `json:"source"`
}

type TraceStore struct {
	mu      sync.RWMutex
	records []TraceRecord
	max     int
}

func NewTraceStore(max int) *TraceStore {
	if max <= 0 {
		max = 1500
	}
	return &TraceStore{max: max, records: make([]TraceRecord, 0, max)}
}

func (s *TraceStore) Add(record TraceRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.records) >= s.max {
		copy(s.records, s.records[1:])
		s.records[len(s.records)-1] = record
		return
	}
	s.records = append(s.records, record)
}

func (s *TraceStore) List(limit int) []TraceRecord {
	if limit <= 0 {
		limit = 50
	}
	s.mu.RLock()
	items := make([]TraceRecord, len(s.records))
	copy(items, s.records)
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp.After(items[j].Timestamp)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *TraceStore) Seed(records []TraceRecord) {
	for _, item := range records {
		s.Add(item)
	}
}
