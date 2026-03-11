package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLivenessEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rr := httptest.NewRecorder()

	newMux().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
