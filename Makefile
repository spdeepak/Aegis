.PHONY: clean clean-frontend generate generate-frontend build build-backend build-frontend build-embedded dev-backend dev-frontend test-backend test-frontend

# Backend
clean:
	cd apps/server && find . -name "*.gen.go" -type f -delete

generate: clean
	cd apps/server && mkdir -p api
	cd apps/server && go tool oapi-codegen -config .oapi-codegen.yaml openapi.yaml
	cd apps/server && go tool sqlc generate
	cd apps/server && go tool mockery

build-backend:
	cd apps/server && go build -o ../../bin/server ./cmd/server

test-backend:
	cd apps/server && go test -p 1 ./...

test-backend-coverage:
	cd apps/server && go test -p 1 -coverprofile=coverage.out ./... && { head -n1 coverage.out; grep -vE '\.gen\.go:' coverage.out | tail -n +2; } > coverage.filtered.out && go tool cover -func=coverage.filtered.out | tail -n1 && rm coverage.out coverage.filtered.out

dev-backend:
	cd apps/server && go run ./cmd/server

# Frontend
clean-frontend:
	cd apps/web && find src -name "*.gen.ts" -type f -delete

generate-frontend: clean-frontend
	cd apps/web && npm run generate:api

build-frontend:
	npm run build --workspace=@aegis/web

dev-frontend:
	npm run dev --workspace=@aegis/web

test-frontend:
	npm run typecheck --workspaces

# API Client
generate-client:
	npm run generate --workspace=@aegis/api-client

# Embedded (single binary with frontend)
build-embedded: build-frontend
	cd apps/server && go build -o ../../bin/aegis ./cmd/server

# All
build: build-backend build-frontend

dev: dev-backend dev-frontend

test: test-backend test-frontend
