# apps/server — Agent Guide

Extends the repo-root `AGENTS.md` — read that first for setup, env vars, Docker, and CI. Everything below is specific to the Go backend.

## Package map

```
internal/
├── users, roles, permissions, tokens, twoFA, jwt_secret   # domains — service.go (hand-written) +
│                                                            # db/models/querier.gen.go (sqlc) + mock (mockery)
├── config/       # viper-backed loader: configs/application.yaml + configs/secrets.json, hot-reloaded via fsnotify
├── db/           # init.go (pgxpool connection), migration.go (golang-migrate runner, called at startup)
├── error/        # package httperror — the error-code registry, see below
└── middleware/   # auth.go, error.go, gin.go, metrics.go — see request lifecycle below
pkg/
├── logging/      # custom slog.Handler — see Logging below
├── time/         # Now() → time.Now().UTC(), the forbidigo-enforced wrapper
├── ttlcache/     # generic in-memory TTL cache
└── util/         # PasswordValidator, UUID↔pgtype conversions, SliceSplit/HasAny/HasAll
```

## Request lifecycle

`main.go` wires one global middleware chain (order matters): `MetricHandler` → `RequestValidator` → `JWTAuthMiddleware` → `gin.Recovery()` → `ErrorMiddleware` → `GinLogger`. Two things are easy to miss:

- **`/metrics` skips all of it.** It's registered via `router.GET("/metrics", ...)` *before* `router.Use(...)` is called, so Gin never attaches the chain to it — it's not in any skip-list, it just never reaches the middleware at all.
- **`RequestValidator`** (oapi-codegen's gin-middleware + kin-openapi) validates every request against `openapi.yaml` *and* reads a custom `aegis-auth` OpenAPI extension off the matched operation (roles/permissions required for that route), stashing it in the Gin context for `JWTAuthMiddleware` to enforce. If you're adding a permission-gated endpoint, that gating is declared in `openapi.yaml`, not written in Go.
- **`JWTAuthMiddleware`** skips a hardcoded list — `/ready`, `/live`, `/api/v1/auth/{signup,login,refresh}` — in addition to `auth.skipPaths` from `configs/application.yaml`. Add new public routes to the config, not the hardcoded list.
- **Two different error paths.** Handlers call `_ = ctx.Error(err)` and return; `GinLogger` is what actually inspects `c.Errors`, picks Warn (4xx) vs Error (5xx) by unwrapping `httperror.HttpError`, and writes the JSON response. `ErrorMiddleware` is unrelated — it's a `recover()` wrapper that only fires on a panic, always returns 500, and doesn't run for ordinary returned errors.

## Adding an endpoint — the actual sequence

1. Add the path/schema to `openapi.yaml` (include the `aegis-auth` extension if it needs role/permission gating).
2. `make generate` — regenerates `api.ServerInterface` (you'll get a compile error until you implement the new method).
3. Implement the method on `Server` in `cmd/server/server.go`: bind the request, call a domain service, `ctx.Error(err)` and return on failure.
4. If it needs a new query, add it to `sqls/<domain>.sql` (or a new migration in `migrations/` if the schema changes), then `make generate` again — sqlc regenerates `Querier` and mockery regenerates its mock.
5. Implement the logic in the domain's `service.go`, returning `httperror.New(...)` / `httperror.NewWithMetadata(...)` for failures (add a new error code to `internal/error/error.go` if none fits).
6. Unit-test with `NewMockQuerier(t)` + `.EXPECT()...Return(...)` (mockery v3, testify template); add an integration test alongside the others if the behavior depends on real Postgres.

## Errors

`internal/error` (package `httperror`) is a registry, not ad hoc errors: each code (`JWT0001`…) maps to a fixed `StatusCode` + `Description` in a `map[string]HttpError`. Construct with `httperror.New(key)`, `httperror.NewWithMetadata(key, detail)` (detail goes in a `Metadata` field that's excluded from the JSON response — safe for internal error text), or `httperror.NewWithStatus(key, detail, status)` to override the status. An unrecognized key silently falls back to `UndefinedErrorCode` (500) — check `error.go` for the existing codes before adding a new one; there's likely already one close to what you need.

## Time

`pkg/time.Now()` just returns `time.Now().UTC()` — the forbidigo rule exists to keep every timestamp UTC, not for its own sake. The import alias isn't standardized: existing code uses `pkgTime`, `pkgtime`, `pkt`, and unaliased, inconsistently (roughly: implementation files lean `pkgtime`, test files lean `pkgTime`, but there are exceptions). Match whatever the file you're editing already uses; if a new file needs both stdlib `time` and this package, alias one of them to avoid a collision.

## Testing

- Unit tests live beside the code (`<domain>/service_test.go`) and mock `Querier` via mockery — no real DB. Integration tests (`*_integration_test.go`, e.g. `cmd/server`, `users`, `roles`, `permissions`) hit real Postgres at `localhost:5432` (`admin`/`admin`/`jwt_server`, matching `docker-compose.yaml`) with no build-tag separation — `docker compose up -d` first, or `go test ./...` just fails on connection errors.
- Naming convention: success/failure cases are usually separate functions suffixed `_OK` / `_NOK` (e.g. `TestUUIDToPgtypeUUID_OK`, `TestService_ValidateRefreshToken_OK`), not table-driven subtests.
- `gin.CreateTestContext(httptest.NewRecorder())` + `httptest.NewRequest(...)` is the standard way to build a fake `*gin.Context` for a handler/service test.
- `make test-backend` runs `lint-backend` first — a lint failure blocks the test run locally, though CI runs them as separate steps.

## Logging

`slog` with a custom JSON handler (`pkg/logging`) that injects `Correlation-Id`, `User-Agent`, and `User-Email` into every record — from context values if set, or straight off the incoming Gin request's headers otherwise. Level is controlled by the `LOG_LEVEL` env var (`DEBUG`/`INFO`/`WARN`/`ERROR`, defaults to `INFO`). Prefer `slog.InfoContext(ctx, ...)` / `ErrorContext` over the non-context variants so the correlation ID actually gets attached.

## Config

`internal/config` reads `configs/application.yaml` (path overridable via `CONFIG_FILE_PATH`) and `configs/secrets.json` (via `SECRETS_FILE_PATH`) through viper, and both are watched for live reload (`fsnotify`) — no restart needed for a config-only change. `validateConfig` runs `go-playground/validator` struct tags on startup and exits the process on any required field being missing, so an incomplete `application.yaml` fails fast rather than serving with zeroed config.