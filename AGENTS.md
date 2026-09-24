# AGENTS.md

## Commands (use Makefile)

- `make dev` — live reload via air (binary `./tmp/main`, config `.air.toml`)
- `make dev-all` — backend (air) + frontend (`bun run dev` in `ui/`) concurrently (Ctrl-C stops both)
- `make run` / `make build` — run or build `./cmd/api`
- `make test` / `go test ./internal/auth/ -run TestSignupEmail -v` — all / focused tests
- `make check` (= `go vet ./...` + `golangci-lint run ./...` + `go test ./...`)
- `make lint` / `make lint-fix` — lint / lint with auto-fix
- `make ent-new NAME=Group` — new schema in `ent/schema/`; then `make ent-generate`
- `make setup` — copies `.env.example` → `.env` if missing
- `make up` / `make down` / `make logs` — observability stack (stop `make dev` first, both want `:8080`)

## Finish gate (required)

- Before finishing any task: `make check` AND `make build` must both pass with fresh runs. No success claims without that evidence.

## Setup gotcha

- `config.New()` returns an error if `DATABASE_URL` or `JWT_SECRET` is unset. Server needs `.env` present (godotenv loads from repo root). Unit tests never call it — they construct services directly (`NewService(client, ...)`, no config needed).

## Echo v5 quirks (verified)

- Handler signature is `func (h *X) Y(c *echo.Context) error` — pointer context.
- Always set `e.Validator = validation.New()` (`pkg/validation`, validator/v10 wrapper).
- No `e.Shutdown` / `e.Start` for graceful shutdown — use `echo.StartConfig{Address, GracefulTimeout}.Start(sigCtx, e)` with `signal.NotifyContext` (see `cmd/api/main.go`).
- Routing standard: Scoped route groups (`public` vs `protected := api.Group("", auth.Middleware(secret))`). Domain handlers implement `RegisterRoutes(public, protected *echo.Group)`. Never use global auth middleware with hardcoded URL path maps.

## Ent + Postgres

- Schemas live in `ent/schema/`; everything else under `ent/` is generated — never hand-edit.
- `internal/platform/db.go`: opens with pgx stdlib (`sql.Open("pgx", url)`, `dialect.Postgres`). `AutoMigrate` uses `WithDropIndex/WithDropColumn` and is gated to `APP_ENV=development` in `main.go` — never call it in prod paths.
- Connection pooling: configured in `platform.Open` with explicit max open/idle connections and lifetime. Multi-entity transactions use `platform.WithTx(ctx, client, func(tx *ent.Tx) error)`.
- Configuration lives in `internal/config/config.go`. Loaded via `config.New()`.

## Auth pattern

- `internal/auth/`: `auth_model.go` (request/response + `validate` tags) → `auth_service.go` (`NewService(db, secret, accessTTL, refreshTTL)` holds `*ent.Client` directly, no repo layer) → `auth_handler.go` (`response.Bind` → service → `response.Error`). Handlers mount public/protected endpoints via `RegisterRoutes(public, protected)`.
- Shared handler helpers: `response.Bind(c, &req) bool` (bind+validate, writes 400), `response.CurrentUserID(c)` (writes 401, reads `response.UserIDKey`), `auth.ViewerID(r, secret)` (optional auth, `""` when anonymous). Cookie build/clear in `auth_cookies.go`; `secure` + `env` injected into `NewHandler` (no `GetAppEnv`, no config in services).
- Passwords: bcrypt, `max=72` in the validate tag is the bcrypt limit — keep it. Never return the hash; use `ToUserResponse` (`profile.ToResponse` builds on it).
- Duplicate email/username surfaces as `ent.IsConstraintError` → 409.

## Logging (slog, stdlib only)

- Build via `logger.New(env, level)` (`pkg/logger`): JSON to stdout in production, pretty single-line text elsewhere (`18:04:12.345 INFO POST /path → 200 · 3ms ...`), `AddSource` always on (file:line on every record). `main.go` calls `slog.SetDefault(log)` so `response.Error` can log without plumbing. No new logging deps.
- Dev colors are ANSI and TTY-gated (`NO_COLOR`/`TERM=dumb` respected; pipes/tests stay plain) via a custom `slog.Handler` (`pkg/logger/pretty.go`) — level color doubles as status color since access logs map status class to level.
- Handlers and services hold no logger. Only two places log: `response.Error` (5xx only, via `slog.Default`) and `main.go` startup paths (`log.Error(...)` then `os.Exit(1)` — slog has no Fatal). 4xx are never logged per-request; the access log already records them.
- Request ID: `appmiddleware.Register(e, log)` (`internal/middleware/middleware.go`) owns the whole stack — Recover, RequestID, slog RequestLogger, CORS. `middleware.RequestID` must stay **before** `RequestLogger`. Access logs use a custom `RequestLoggerConfig.LogValuesFunc` in `internal/middleware` (Info <400 / Warn 4xx / Error 5xx; human summary `METHOD path → status · ms` in `msg`, details as attrs). Every logged value needs its `LogMethod/LogURIPath/...` opt-in flag or it arrives as a zero value. Keep `internal/config` on stdlib errors — the logger doesn't exist at config-load time.
- `main.go` sets `e.Logger = log`: Echo's own logs (banner, startup, HTTP errors) default to JSON and must be routed through our handler. `source` paths are trimmed to two segments (`auth/auth_service.go:56`).
- Level override: `LOG_LEVEL` env (`debug/info/warn/error`, default `info`) via `Config.LogLevel`. No vendor shipper SDK — stdout JSON is the shipping contract.

## Where things go

- `internal/<domain>/` — one package per product area (`auth`, `profile`): owns its routes, service, models. New area = new package + schema + `RegisterRoutes` + tests, never a placeholder dir.
- `internal/filestore/` — upload infra behind `New(token, dir)` + `Upload(ctx, name, data)` (`Result{URL,Key,Name,Size}`); UploadThing with local fallback, `MaxUploadBytes` shared with handlers. Any domain needing uploads uses this.
- `internal/middleware/` — global Echo stack only (Recover, RequestID, access log, CORS, metrics). Domain auth (`auth.Middleware`, `auth.ViewerID`) stays in `internal/auth` next to token verification — middleware must never import a domain.
- `internal/platform/` — DB open/pooling/transactions; `internal/config/` — env only, returns `(*Config, error)`, `IsProd()` for prod checks.
- `pkg/response` — HTTP helpers (`Map`, `Bind`, `Error`, `Unauthorized`, `BadRequest`, `CurrentUserID`, `UserIDKey`); `pkg/{apperror,validation,logger}` stay framework-free shared libs.
- `ent/schema/` is the source of truth; everything else under `ent/` is generated.
- Rules: no domain→domain imports for helpers (use `pkg/response`); one-way model reuse (`profile` builds on `auth.ToUserResponse`) is allowed, revisit at 3 domains. No empty placeholder dirs — `social/`, `ws/`, `test/integration/` were deleted for exactly this reason.

## File naming

- Domain files: `<domain>_<layer>.go` for model/service/handler (`auth_service.go`, `profile_model.go`), `<domain>_<topic>.go` for scoped extras (`auth_cookies.go`, `auth_request.go`) — tests mirror (`auth_service_test.go`).
- Single-purpose packages: bare name — `pkg/logger/logger.go`, `pkg/validation/validator.go`, `internal/platform/db.go`, `internal/filestore/filestore.go`.

## Tests

- testify, co-located `*_test.go` next to code (hybrid layout).
- DB tests use `enttest.Open(t, "sqlite3", "file:<name>?mode=memory&cache=shared&_fk=1")` — requires the blank `_ "github.com/mattn/go-sqlite3"` import. Use a unique `file:<name>` per test file to avoid shared-cache cross-talk.
- Handler tests call `h.SignupEmail(e.NewContext(req, rec))` directly, not over HTTP.

## Lint (golangci-lint v2, pinned v2.11.4 — match it locally)

- Config `.golangci.yml` (`version: "2"`, `default: standard` + `errorlint`/`misspell`; formatters `gofmt`+`goimports` with local prefix). Barebone on purpose: no `revive`/`gocritic`/`gosec` — comments only where they explain why, style via `gofmt`. CI (`.github/workflows/ci.yml`, `lint` + `test` jobs, main-push/PR only) runs the same version.
- `ent/` excluded via `generated: lax` — never fix generated code, fix the schema or config instead. Test files skip `errcheck`.
- No doc-comment-per-export rule. Comment the why (reuse-revokes-all, timing guard, committed-response), not the what. Use `errors.As`, never `err.(Type)`.
- `gofumpt` is omitted: this golangci version reports nondeterministic gofumpt findings (standalone gofumpt is clean). Re-evaluate after upgrading past v2.11.4.

## Ignored

- `tmp/` (air builds) and `docs/superpowers/` are gitignored.

## Observability

- `observability/` holds Prometheus/Loki/Alloy/Grafana configs; app exposes `GET /metrics` (`townhall_http_requests_total`, `townhall_http_request_duration_seconds`) wired in `internal/middleware` (`RegisterMetrics`, excluded from its own instrumentation).
- Compose DB is separate from native Postgres.app on `:5432` (compose port unpublished to avoid the clash); Compose app uses `db:5432` and `APP_ENV=production` for Loki-friendly JSON logs. Non-destructive schema create runs only with explicit `AUTO_MIGRATE=true` (`platform.Migrate`); destructive `AutoMigrate` stays dev-only. DB dial retries 15×2s at startup for container ordering.

## Frontend (ui/, bun + React + Vite + TS)

- Commands (run in `ui/`): `bun run dev` (vite `:5173`, `/api` proxied to `:8080`) / `bun run build` (`tsc -b && vite build`) / `bun run lint` / `bun run typecheck`.
- Finish gate (UI tasks): `bun run typecheck` + `bun run lint` + `bun run build`, all green. Backend `make check`/`make build` unaffected (no backend changes in UI tasks).
- Structure: feature folders — `src/features/<feature>/` (`components/`, `*Api.ts`, `*Slice.ts`, `*Page.tsx`); shared `src/app/` (store, router, hooks), `src/lib/`, `src/components/`. All domain page views live inside their respective feature folder (e.g. `src/features/messages/MessagesPage.tsx`).
- Naming: components/pages PascalCase matching the default export (`LoginPage.tsx`); hooks `use*.ts`; slices/apis/store camelCase (`authSlice.ts`); no barrel `index.ts` re-exports — import directly via `@/` alias (`tsconfig.app.json` paths + `vite.config.ts` resolve.alias).
- RTK Query: single central `baseApi` in `src/lib/baseApi.ts` (`reducerPath: 'api'`) with shared `tagTypes`. Features inject endpoints via `baseApi.injectEndpoints({ ... })`. `store.ts` only ever mounts `baseApi.reducer` and `baseApi.middleware`.
- Route code-splitting: Route page components in `src/app/router.tsx` must be loaded dynamically via `React.lazy()` with `<Suspense fallback={<PageLoadingSkeleton />}>` to keep the main bundle under 400 kB.
- Auth rules: access token lives in Redux memory only — never `localStorage`/`sessionStorage`, never logged; refresh travels by HttpOnly cookie (`credentials: "include"`); 401s funnel through `baseQueryWithReauth` (refresh-then-retry-once, else `clearCredentials`).
- Backend contract mirror: TS types in `*Api.ts` must match Go model JSON tags exactly (`access_token`, `expires_in`, `email_verified`, …). Changing the Go contract means updating the TS types in the same task.
