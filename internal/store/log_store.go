package store

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields"`
}

type LogStore struct {
	mu      sync.RWMutex
	entries []LogEntry
	max     int
}

func NewLogStore(max int) *LogStore {
	if max <= 0 {
		max = 3000
	}
	return &LogStore{max: max, entries: make([]LogEntry, 0, max)}
}

func (s *LogStore) Write(p []byte) (int, error) {
	lines := bytes.Split(p, []byte("\n"))
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		entry := decodeLine(line)
		s.add(entry)
	}
	return len(p), nil
}

func decodeLine(line []byte) LogEntry {
	fallback := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "info",
		Message:   strings.TrimSpace(string(line)),
		Fields:    map[string]any{},
	}

	payload := map[string]any{}
	if err := json.Unmarshal(line, &payload); err != nil {
		return fallback
	}

	entry := LogEntry{Fields: map[string]any{}}
	for k, v := range payload {
		switch k {
		case "time":
			if str, ok := v.(string); ok {
				entry.Timestamp = str
			}
		case "level":
			if str, ok := v.(string); ok {
				entry.Level = str
			}
		case "message":
			if str, ok := v.(string); ok {
				entry.Message = str
			}
		default:
			entry.Fields[k] = v
		}
	}

	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if entry.Level == "" {
		entry.Level = "info"
	}
	if entry.Message == "" {
		entry.Message = "log entry"
	}
	return entry
}

func (s *LogStore) add(entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) >= s.max {
		copy(s.entries, s.entries[1:])
		s.entries[len(s.entries)-1] = entry
		return
	}
	s.entries = append(s.entries, entry)
}

func (s *LogStore) List(limit int, level string) []LogEntry {
	if limit <= 0 {
		limit = 100
	}
	level = strings.TrimSpace(strings.ToLower(level))

	s.mu.RLock()
	items := make([]LogEntry, 0, len(s.entries))
	for _, item := range s.entries {
		if level == "" || strings.ToLower(item.Level) == level {
			items = append(items, item)
		}
	}
	s.mu.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp > items[j].Timestamp
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}
