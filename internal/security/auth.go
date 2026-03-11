package security

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/freelance-engineer/go-service-template-pro/internal/domain"
)

type Principal struct {
	Type    string `json:"type"`
	Subject string `json:"subject"`
}

type Authenticator struct {
	apiKeys  [][]byte
	jwtKey   []byte
	issuer   string
	audience string
}

func NewAuthenticator(apiKeys []string, jwtSecret, issuer, audience string) (*Authenticator, error) {
	keyBytes := make([][]byte, 0, len(apiKeys))
	for _, key := range apiKeys {
		trimmed := strings.TrimSpace(key)
		if trimmed != "" {
			keyBytes = append(keyBytes, []byte(trimmed))
		}
	}
	if len(keyBytes) == 0 && strings.TrimSpace(jwtSecret) == "" {
		return nil, fmt.Errorf("no auth mechanism configured")
	}

	return &Authenticator{
		apiKeys:  keyBytes,
		jwtKey:   []byte(jwtSecret),
		issuer:   issuer,
		audience: audience,
	}, nil
}

func (a *Authenticator) Authenticate(r *http.Request) (Principal, error) {
	if principal, ok := a.authenticateJWT(r); ok {
		return principal, nil
	}
	if principal, ok := a.authenticateAPIKey(r); ok {
		return principal, nil
	}
	return Principal{}, domain.ErrUnauthorized
}

func (a *Authenticator) authenticateAPIKey(r *http.Request) (Principal, bool) {
	provided := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if provided == "" {
		return Principal{}, false
	}
	bytesProvided := []byte(provided)
	for _, key := range a.apiKeys {
		if subtle.ConstantTimeCompare(key, bytesProvided) == 1 {
			return Principal{Type: "api_key", Subject: "client"}, true
		}
	}
	return Principal{}, false
}

func (a *Authenticator) authenticateJWT(r *http.Request) (Principal, bool) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return Principal{}, false
	}
	if len(a.jwtKey) == 0 {
		return Principal{}, false
	}
	tokenStr := strings.TrimSpace(authHeader[7:])
	if tokenStr == "" {
		return Principal{}, false
	}

	claims := jwt.RegisteredClaims{}
	parserOptions := []jwt.ParserOption{}
	if a.issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(a.issuer))
	}
	if a.audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(a.audience))
	}
	parser := jwt.NewParser(parserOptions...)
	token, err := parser.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.jwtKey, nil
	})
	if err != nil || !token.Valid {
		return Principal{}, false
	}
	subject := claims.Subject
	if strings.TrimSpace(subject) == "" {
		subject = "jwt-client"
	}
	return Principal{Type: "jwt", Subject: subject}, true
}
