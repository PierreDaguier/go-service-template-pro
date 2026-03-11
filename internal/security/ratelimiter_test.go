package security

import (
	"testing"
	"time"
)

func TestRateLimiterAllowThenBlock(t *testing.T) {
	limiter := NewRateLimiter(1, 1, time.Minute)
	if !limiter.Allow("client-a") {
		t.Fatal("first request should pass")
	}
	if limiter.Allow("client-a") {
		t.Fatal("second immediate request should be blocked")
	}
}
