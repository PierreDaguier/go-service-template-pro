SHELL := /bin/bash

.PHONY: help deps dev down backend-test backend-lint backend-build frontend-install frontend-lint frontend-build ci screenshot

help:
	@echo "Targets:"
	@echo "  deps             Install frontend dependencies"
	@echo "  dev              Start full stack with Docker Compose"
	@echo "  down             Stop compose stack"
	@echo "  backend-test     Run Go tests"
	@echo "  backend-lint     Run go vet + format check"
	@echo "  backend-build    Build backend binary"
	@echo "  frontend-install Install npm dependencies"
	@echo "  frontend-lint    Run frontend lint"
	@echo "  frontend-build   Build frontend"
	@echo "  ci               Run backend+frontend checks"
	@echo "  screenshot       Capture panel screenshots"

deps: frontend-install

frontend-install:
	cd web && npm install

frontend-lint:
	cd web && npm run lint

frontend-build:
	cd web && npm run build

backend-test:
	go test ./...

backend-lint:
	@test -z "$(shell gofmt -l $(shell find cmd internal -name '*.go'))" || (echo "Run gofmt" && exit 1)
	go vet ./...

backend-build:
	go build ./cmd/service

dev:
	docker compose up --build

down:
	docker compose down -v

ci: backend-test frontend-lint frontend-build

screenshot:
	./scripts/capture-screenshots.sh
