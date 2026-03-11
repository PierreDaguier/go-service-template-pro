package domain

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrRateLimited  = errors.New("rate_limited")
	ErrValidation   = errors.New("validation_error")
	ErrNotFound     = errors.New("not_found")
)
