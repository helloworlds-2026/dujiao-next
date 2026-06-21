# AGENTS.md

## Cursor Cloud specific instructions

This repo (`dujiao-next`) is the Go backend (Gin + GORM) for the Dujiao-Next
e-commerce ecosystem. The matching frontends live in sibling repos `admin`
(management console) and `user` (storefront), both of which proxy `/api` to this
service on port `8080`.

Dependencies (Go modules) are refreshed automatically by the startup update
script; do not reinstall them here. Standard commands are in `README.md`.

### Redis is required (and needs special start flags in this sandbox)
`redis.enabled` and `queue.enabled` default to `true`, so the API/worker need a
running Redis on `127.0.0.1:6379`. In this VM, Redis RDB `bgsave` fails and, with
the default `stop-writes-on-bgsave-error yes`, Redis then **blocks all writes** —
which silently breaks admin login (the login rate-limiter writes to Redis).
Start Redis with RDB snapshots disabled:

```bash
redis-server --daemonize yes --save '' --stop-writes-on-bgsave-error no
```

If Redis is already running and login fails with write errors, fix it live with:
`redis-cli CONFIG SET stop-writes-on-bgsave-error no`.

### Running the backend
From the repo root: `go run ./cmd/server` (default mode `all` = HTTP API + asynq
worker). No `config.yml` is required — `config.Load()` falls back to defaults
(SQLite at `./db/dujiao.db`, Redis on localhost). To customize, copy
`config.yml.example` to `config.yml`. Health check: `GET /health`.

### Default admin
On first run in `debug` mode a super-admin is auto-created and logged at startup:
username `admin`, password `admin123`. Admin login API:
`POST /api/v1/admin/login`. The SQLite DB persists at `./db/dujiao.db`, so the
admin survives restarts.

### Tests / vet
`go test ./...` (the `internal/service` suite alone takes ~30s) and
`go vet ./...`. There is no separate lint config.
