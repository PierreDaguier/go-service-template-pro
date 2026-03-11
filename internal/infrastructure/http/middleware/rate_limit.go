package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/freelance-engineer/go-service-template-pro/internal/domain"
	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/response"
	"github.com/freelance-engineer/go-service-template-pro/internal/security"
)

type RateLimit struct {
	limiter *security.RateLimiter
}

func NewRateLimit(limiter *security.RateLimiter) *RateLimit {
	return &RateLimit{limiter: limiter}
}

func (m *RateLimit) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := identityFromRequest(r)
		if !m.limiter.Allow(identity) {
			status, apiErr := response.FromDomain(domain.ErrRateLimited, RequestIDFromContext(r.Context()))
			response.Error(w, status, apiErr)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func identityFromRequest(r *http.Request) string {
	if principal, ok := PrincipalFromContext(r.Context()).(security.Principal); ok {
		return principal.Type + ":" + principal.Subject
	}
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "anonymous"
}
