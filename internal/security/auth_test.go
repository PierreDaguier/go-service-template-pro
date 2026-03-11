package security

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthenticateAPIKey(t *testing.T) {
	auth, err := NewAuthenticator([]string{"abc123"}, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/overview", nil)
	req.Header.Set("X-API-Key", "abc123")

	principal, err := auth.Authenticate(req)
	if err != nil {
		t.Fatalf("expected no auth error, got %v", err)
	}
	if principal.Type != "api_key" {
		t.Fatalf("expected api_key principal, got %s", principal.Type)
	}
}

func TestAuthenticateJWT(t *testing.T) {
	secret := "top-secret"
	auth, err := NewAuthenticator(nil, secret, "issuer-a", "aud-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "issuer-a",
		Audience:  []string{"aud-a"},
		Subject:   "client-42",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/overview", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	principal, err := auth.Authenticate(req)
	if err != nil {
		t.Fatalf("expected no auth error, got %v", err)
	}
	if principal.Subject != "client-42" {
		t.Fatalf("unexpected subject: %s", principal.Subject)
	}
}
