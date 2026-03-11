package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/freelance-engineer/go-service-template-pro/internal/security"
)

func TestAuthMiddlewareRejectsMissingCredentials(t *testing.T) {
	authenticator, err := security.NewAuthenticator([]string{"valid-key"}, "", "", "")
	if err != nil {
		t.Fatalf("setup authenticator: %v", err)
	}
	auth := NewAuth(authenticator)

	h := auth.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRateLimitMiddlewareBlocksAfterBurst(t *testing.T) {
	limiter := security.NewRateLimiter(1, 1, 10*time.Second)
	rate := NewRateLimit(limiter)

	h := rate.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	req1.RemoteAddr = "127.0.0.1:10001"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	req2.RemoteAddr = "127.0.0.1:10001"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)

	if rr1.Code != http.StatusNoContent {
		t.Fatalf("first request should pass, got %d", rr1.Code)
	}
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request should be rate limited, got %d", rr2.Code)
	}
}
