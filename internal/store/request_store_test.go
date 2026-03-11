package store

import (
	"testing"
	"time"
)

func TestSnapshotAggregates(t *testing.T) {
	store := NewRequestStore(100)
	now := time.Now().UTC()
	store.Add(RequestRecord{Timestamp: now.Add(-1 * time.Minute), Method: "GET", Route: "GET /a", StatusCode: 200, DurationMs: 25})
	store.Add(RequestRecord{Timestamp: now.Add(-1 * time.Minute), Method: "GET", Route: "GET /a", StatusCode: 502, DurationMs: 220})

	snapshot := store.Snapshot(5 * time.Minute)
	if snapshot.Requests != 2 {
		t.Fatalf("expected 2 requests, got %d", snapshot.Requests)
	}
	if snapshot.ErrorRate <= 0 {
		t.Fatalf("expected non-zero error rate, got %f", snapshot.ErrorRate)
	}
	if len(snapshot.Series) == 0 {
		t.Fatal("expected at least one series point")
	}
}
