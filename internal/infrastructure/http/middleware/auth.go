package middleware

import (
	"net/http"

	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/http/response"
	"github.com/freelance-engineer/go-service-template-pro/internal/security"
)

type Auth struct {
	authenticator *security.Authenticator
}

func NewAuth(authenticator *security.Authenticator) *Auth {
	return &Auth{authenticator: authenticator}
}

func (m *Auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := m.authenticator.Authenticate(r)
		if err != nil {
			status, apiErr := response.FromDomain(err, RequestIDFromContext(r.Context()))
			response.Error(w, status, apiErr)
			return
		}
		next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), principal)))
	})
}
