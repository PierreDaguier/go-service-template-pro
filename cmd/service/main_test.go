package main

import (
	"testing"

	"github.com/freelance-engineer/go-service-template-pro/internal/infrastructure/config"
)

func TestDefaultConfigLoads(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}
	if cfg.ServiceName == "" {
		t.Fatal("expected non-empty default service name")
	}
}
