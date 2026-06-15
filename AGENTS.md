# AGENTS.md

## Cursor Cloud specific instructions

This repo is the **Dujiao-Next API**: a single Go (1.25.x) backend (Gin + GORM). It produces
one server process (`cmd/server`) that runs in `-mode all|api|worker` (default `all` = API + async
worker), plus an `admin-tool` CLI (`cmd/admin-tool`). The admin/user SPA frontends live in
separate repos and are only embedded in `-tags fullstack` release builds, so there is no frontend
to run here.

Standard commands are documented in `README.md`. Build tags and release flow live in `Dockerfile`
and `.goreleaser.yaml`. Notes below are only the non-obvious bits.

### Runtime dependencies / config
- **`config.yml` is required at runtime and is git-ignored.** Create it once with
  `cp config.yml.example config.yml`. Viper also reads env-var overrides (`.`→`_`, e.g.
  `SERVER_PORT`, `DATABASE_DRIVER`). The update script does NOT create it (it must not clobber a
  customized one), so create it manually if missing.
- **Default DB is SQLite** (`./db/dujiao.db`, auto-created). No external DB needed. GORM
  auto-migrates on every startup — there is no manual migration step.
- **Redis is needed for `worker`/`all` mode** (asynq queue on logical DB 1) and for cache /
  login-rate-limit / 2FA-replay (DB 0). It is NOT pre-started and NOT installed by the update
  script (system dependency). Start it yourself before running in `all`/`worker` mode:
  `redis-server --daemonize yes` (or run it in a tmux session). To run API-only with no Redis,
  set `redis.enabled: false` AND `queue.enabled: false` in `config.yml` (worker mode then errors
  out by design).
- In the example config `email.enabled: true` points at a fake SMTP host; set it to `false` for
  local dev unless you have a real SMTP server.
- To get a usable admin account for testing, set `bootstrap.default_admin_username` /
  `bootstrap.default_admin_password` in `config.yml` before first startup (the admin is seeded
  once on initialization). Alternatively use `go run cmd/admin-tool/main.go list-admins`.
- `server.mode: release` makes the app **fatal** on weak/default `app.secret_key`, `jwt.secret`,
  `user_jwt.secret`. Keep `mode: debug` locally (the example config already does).

### Run / verify
- Run everything: `go run cmd/server/main.go` (listens on `0.0.0.0:8080`). Health check:
  `curl http://127.0.0.1:8080/health` → `{"status":"ok"}`.
- Public storefront routes are under `/api/v1/public/...`; admin routes under `/api/v1/admin/...`
  (login at `POST /api/v1/admin/login`, then `Authorization: Bearer <token>`).

### Lint / test / build
- Lint: `go vet ./...` and `gofmt -l .` (there is no golangci-lint config). Note
  `internal/web/handler_test.go` is reported by `gofmt -l` on a clean checkout — pre-existing, not
  something you introduced.
- Test: `go test ./...` — passes with only SQLite/in-memory; needs no external services
  (`internal/service` is the slow package, ~30s). Postgres integration tests are gated behind
  `-tags=integration` + `TEST_POSTGRES_DSN`.
- Build: `go build -o /tmp/dujiao-api ./cmd/server`. The `fullstack` build tag additionally
  requires `internal/web/dist/{admin,user}` (populated from the other repos in CI) and will fail
  without them.
