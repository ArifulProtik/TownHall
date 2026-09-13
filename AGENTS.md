# AGENTS.md

## Commands (use Makefile)

- `make dev` — live reload via air (binary `./tmp/main`, config `.air.toml`)
- `make run` / `make build` — run or build `./cmd/api`
- `make test` / `go test ./internal/auth/ -run TestSignupEmail -v` — all / focused tests
- `make check` (= `go vet ./...` + `golangci-lint run ./...` + `go test ./...`)
- `make lint` / `make lint-fix` — lint / lint with auto-fix
- `make ent-new NAME=Group` — new schema in `ent/schema/`; then `make ent-generate`
- `make setup` — copies `.env.example` → `.env` if missing

## Finish gate (required)

- Before finishing any task: `make check` AND `make build` must both pass with fresh runs. No success claims without that evidence.

## Setup gotcha

- `config.New()` calls `log.Fatalf` if `DATABASE_URL` or `JWT_SECRET` is unset. Server/tests that boot the app need `.env` present (godotenv loads from repo root). Unit tests avoid this by constructing `&config.Config{...}` directly.

## Echo v5 quirks (verified)

- Handler signature is `func (h *X) Y(c *echo.Context) error` — pointer context.
- Always set `e.Validator = validation.New()` (`pkg/validation`, validator/v10 wrapper).
- No `e.Shutdown` / `e.Start` for graceful shutdown — use `echo.StartConfig{Address, GracefulTimeout}.Start(sigCtx, e)` with `signal.NotifyContext` (see `cmd/api/main.go`).

## Ent + Postgres

- Schemas live in `ent/schema/`; everything else under `ent/` is generated — never hand-edit.
- `internal/platform/db.go`: opens with pgx stdlib (`sql.Open("pgx", url)`, `dialect.Postgres`). `AutoMigrate` uses `WithDropIndex/WithDropColumn` and is gated to `APP_ENV=development` in `main.go` — never call it in prod paths.
- Note typo'd filename: config lives in `internal/config/conifg.go`. Don't create a second `config.go`.

## Auth pattern

- `internal/auth/`: `auth_model.go` (request/response + `validate` tags) → `auth_service.go` (`NewService(cfg, db, log)` holds `*ent.Client` directly, no repo layer) → `auth_handler.go` (bind → validate → service, maps `apperror.AppError` to status).
- Passwords: bcrypt, `max=72` in the validate tag is the bcrypt limit — keep it. Never return the hash; use `ToUserResponse`.
- Duplicate email/username surfaces as `ent.IsConstraintError` → 409.

## Logging (slog, stdlib only)

- Build via `logger.New(env, level)` (`pkg/logger`): JSON to stdout in production, pretty single-line text elsewhere (`18:04:12.345 INFO POST /path → 200 · 3ms ...`), `AddSource` always on (file:line on every record). No new logging deps.
- Dev colors are ANSI and TTY-gated (`NO_COLOR`/`TERM=dumb` respected; pipes/tests stay plain) via a custom `slog.Handler` (`pkg/logger/pretty.go`) — level color doubles as status color since access logs map status class to level.
- Never wrap the logger in helper funcs — wrappers break call-site source lines. Pass `*slog.Logger` via constructors (`NewService(cfg, db, log)`), never globals.
- Levels: `Error` = unexpected failures/5xx + fatal startup paths (`log.Error(...)` then `os.Exit(1)` — slog has no Fatal); `Warn` = 4xx, validation failures, duplicate-email; `Info` = lifecycle + access logs.
- Request ID: `appmiddleware.Register(e, log)` (`internal/middleware/middleware.go`) owns the whole stack — Recover, RequestID, slog RequestLogger, CORS. `middleware.RequestID` must stay **before** `RequestLogger`. Stash with `c.Set(logger.RequestIDKey, rid)`, thread into services via `logger.ContextWithRequestID(ctx, rid)`, log with `logger.WithContext(ctx, s.log)`.
- Access logs use a custom `RequestLoggerConfig.LogValuesFunc` in `internal/middleware` (Info <400 / Warn 4xx / Error 5xx; human summary `METHOD path → status · ms` in `msg`, details as attrs). Every logged value needs its `LogMethod/LogURIPath/...` opt-in flag or it arrives as a zero value. Keep `internal/config` on stdlib `log` — the logger doesn't exist at config-load time.
- `main.go` sets `e.Logger = log`: Echo's own logs (banner, startup, HTTP errors) default to JSON and must be routed through our handler. `source` paths are trimmed to two segments (`auth/auth_service.go:56`).
- Level override: `LOG_LEVEL` env (`debug/info/warn/error`, default `info`) via `Config.LogLevel`. No vendor shipper SDK — stdout JSON is the shipping contract.

## File naming

- Domain files: `<domain>_<layer>.go` snake_case — `auth_service.go`, `auth_handler.go`, `auth_model.go` (tests mirror: `auth_service_test.go`).
- Single-purpose packages: bare name — `pkg/logger/logger.go`, `pkg/validation/validator.go`, `internal/platform/db.go`.

## Tests

- testify, co-located `*_test.go` next to code (hybrid layout); `test/integration/` is an empty placeholder.
- DB tests use `enttest.Open(t, "sqlite3", "file:<name>?mode=memory&cache=shared&_fk=1")` — requires the blank `_ "github.com/mattn/go-sqlite3"` import. Use a unique `file:<name>` per test file to avoid shared-cache cross-talk.
- Handler tests call `h.SignupEmail(e.NewContext(req, rec))` directly, not over HTTP.

## Lint (golangci-lint v2, pinned v2.11.4 — match it locally)

- Config `.golangci.yml` (`version: "2"`, `default: none` + curated defect-finders; formatters `gofmt`+`goimports` with local prefix). CI (`.github/workflows/ci.yml`, `lint` + `test` jobs, main-push/PR only) runs the same version.
- `ent/` excluded via `generated: lax` — never fix generated code, fix the schema or config instead. Test files skip `errcheck`/`gosec`.
- `revive` demands doc comments on all exported identifiers and forbids stutter (`auth.Service`, not `auth.AuthService`). Use `errors.As`, never `err.(Type)`. `//nolint` must name the linter + reason.
- `gofumpt` is omitted: this golangci version reports nondeterministic gofumpt findings (standalone gofumpt is clean). Re-evaluate after upgrading past v2.11.4.

## Ignored

- `tmp/` (air builds) and `docs/superpowers/` are gitignored.
