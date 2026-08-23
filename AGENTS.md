# Aegis — Agent Guide

## What this is

Aegis is a JWT authentication and authorization server in Go, built contract-first from an OpenAPI spec. It handles signup/login, refresh-token rotation, IP/device/user-agent token fingerprinting, TOTP-based 2FA, and RBAC (roles + permissions). `apps/web` is a React admin console for managing users, roles, and permissions. It's a monorepo: one Go module for the backend, one npm workspace for the frontend.

## Repo layout

```
Aegis/
├── apps/
│   ├── server/            # Go backend — module github.com/spdeepak/aegis/server, go 1.25
│   │   ├── cmd/server/     # main.go (DI/wiring, graceful shutdown), server.go (HTTP handlers)
│   │   ├── internal/       # domain packages: users, roles, permissions, tokens, twoFA,
│   │   │                   # jwt_secret — plus cross-cutting config, db, error, middleware
│   │   ├── pkg/            # logging, time (clock wrapper), ttlcache, util
│   │   ├── api/            # GENERATED (oapi-codegen) — gitignored, never edit directly
│   │   ├── migrations/     # golang-migrate SQL — also sqlc's schema source
│   │   ├── sqls/           # hand-written queries, one file per domain, feeds sqlc
│   │   ├── configs/        # application.yaml, secrets.json, prometheus/grafana config
│   │   └── openapi.yaml    # API contract — source of truth for both codegen paths below
│   └── web/                # React 19 + Vite 6 + TS + Tailwind v4 admin console
│       └── src/lib/api.gen.ts   # GENERATED (orval) — gitignored, never edit directly
├── k8s/                    # local kind-cluster manifests + deployment walkthrough (README)
├── .github/workflows/go.yml
├── Dockerfile              # backend → distroless nonroot image
├── Dockerfile.web          # frontend → nginx image
├── docker-compose.yaml     # local dev stack: postgres, mailpit, prometheus, grafana
└── Makefile                # single entry point for dev commands, see below
```

The Go module path is lowercase (`github.com/spdeepak/aegis/server`) even though the repo directory is `Aegis` — don't assume casing matches when writing imports or Docker image tags (published as `spdeepak/aegis` / `spdeepak/aegis-web`, also lowercase).

## Setup

Prerequisites: Go 1.25, Node 22, Docker with Compose.

```bash
docker compose up -d   # postgres:18, mailpit, prometheus, grafana
make generate           # oapi-codegen + sqlc + mockery — required before the first build
npm ci                  # frontend deps (npm workspaces, root of repo)
```

The server needs env vars to start: one of `JWT_MASTER_KEY` / `JWT_SECRET_KEY`, plus `DEFAULT_ADMIN_EMAIL` and `DEFAULT_ADMIN_PASSWORD` (all base64-encoded — see README for the full table). Postgres credentials come from `apps/server/configs/secrets.json` (already `admin`/`admin` to match `docker-compose.yaml`) or the `POSTGRES_USER_NAME` / `POSTGRES_PASSWORD` env vars as a fallback.

## Common commands

Run from the repo root via `Makefile`:

| Command | What it does |
|---|---|
| `make generate` | Regenerate everything: oapi-codegen, sqlc, mockery |
| `make dev-backend` / `make dev-frontend` | `go run` the server / Vite dev server |
| `make build` | Build backend and frontend separately |
| `make build-embedded` | Build frontend, then a single Go binary that embeds it |
| `make lint-backend` | `golangci-lint run ./...` |
| `make test-backend` | Runs `lint-backend` first, then `go test -p 1 ./...` |
| `make test-backend-coverage` | Same, plus a filtered coverage report |
| `make generate-frontend` | Regenerate the frontend's typed API client (orval) |
| `make test-frontend` | `tsc --noEmit` across workspaces — there's no frontend unit-test runner yet |

`apps/web` also has `npm run lint` (ESLint) — it isn't wired into any Makefile target or CI step, so run it manually after frontend changes. `make generate-client` targets an `@aegis/api-client` npm workspace that doesn't exist in the repo yet; skip it for now.

**Tests need the Compose stack running.** Integration tests (`*_integration_test.go`, e.g. under `cmd/server`, `users`, `roles`, `permissions`) connect straight to `localhost:5432` using the `admin`/`admin`/`jwt_server` credentials from `docker-compose.yaml`. There's no build-tag split from unit tests and no testcontainers, so plain `go test ./...` will fail with connection errors if `docker compose up -d` hasn't been run first. The `-p 1` flag runs packages serially — tests share that one database, so parallel package runs would collide.

## Code generation — don't hand-edit generated code

Anything matching `*.gen.go` or `*.gen.ts` is generated and gitignored. Change the source and regenerate instead:

- **API shape** (routes, request/response types) → edit `apps/server/openapi.yaml`, then `make generate` (backend) and `make generate-frontend` (frontend).
- **A query** → edit `apps/server/sqls/<domain>.sql`, then `make generate`. sqlc emits `db.gen.go`, `models.gen.go`, and a `Querier` interface (`querier.gen.go`) into `internal/<domain>/`.
- **Schema** → add a new `NNNN_name.up.sql` / `.down.sql` pair to `apps/server/migrations/`, then `make generate` (sqlc reads migrations as its schema source).
- **Mocks** → generated automatically from each domain's `Querier` interface via mockery (testify-style, `<name>_mock.gen.go`). Don't hand-write mocks.

## Conventions worth knowing

- **Never call `time.Now()` directly** in backend code — `.golangci.yml` enforces this with a `forbidigo` rule (`analyze-types` on, so it catches indirect uses too), except inside `_test.go` files and `pkg/time` itself. Use the wrapper in `pkg/time` instead (imported as `pkgTime` in existing code).
- **Per-domain vertical slices**: each of `users`, `roles`, `permissions`, `tokens`, `twoFA`, `jwt_secret` under `internal/` owns its `service.go` (business logic) plus its generated data-access layer. `cmd/server/server.go`'s `Server` struct implements the generated `api.ServerInterface` and stays thin — bind/validate the request, call a service, map the error.
- **Errors** flow through `internal/error` (imported as `httperror`) and `middleware.ErrorMiddleware` — avoid ad hoc `ctx.JSON` error responses.
- Two separate mechanisms make a route public. `middleware.JWTAuthMiddleware` hardcodes a skip-list (`/ready`, `/live`, `/api/v1/auth/signup`, `/api/v1/auth/login`, `/api/v1/auth/refresh`) on top of whatever's in `auth.skipPaths` (`configs/application.yaml`) — use the latter for anything new. `/metrics` bypasses auth for a different reason: it's registered in `main.go` before the global middleware chain is attached, so it never reaches `JWTAuthMiddleware` at all.

## CI (`.github/workflows/go.yml`)

On every push/PR: bring up the Compose stack → `make generate` → `go build ./...` → `make lint-backend` → `go test -p 1 ./...` with coverage (excluding mocks/generated/`cmd/main.go`) → `npm ci` → `make generate-frontend` → `make test-frontend` → upload coverage to Codecov. On `main`, it also builds and pushes multi-arch (`amd64`/`arm64`) Docker images — `spdeepak/aegis` and `spdeepak/aegis-web` — to Docker Hub; other branches build-test the same images without pushing. Match these steps locally before pushing if you want confidence CI will pass.