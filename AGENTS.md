# AGENTS.md

## Cursor Cloud specific instructions

### Ecosystem overview
Dujiao-Next is a self-hosted digital-goods (card/key) shop split across three sibling repos that run together for end-to-end testing:

| Repo | Role | Dev command | Port |
|------|------|-------------|------|
| `dujiao-next` (this repo) | Go backend REST API + async worker | `go run ./cmd/server` | 8080 |
| `admin` | Vue 3 + Vite admin console | `npm run dev` | 5174 |
| `user` | Vue 3 + Vite storefront | `npm run dev` | 5173 |

Both frontends' Vite dev servers proxy `/api` and `/uploads` to `http://localhost:8080`, so the backend must be running for them to function.

### Backend (this repo)
- Requires **Go 1.25+** (matches `go.mod`); it is preinstalled at `/usr/local/go` (on PATH as `go`). Standard build/run is in `README.md` (`go mod tidy`, `go run cmd/server/main.go`).
- Before first run, create the config: `cp config.yml.example config.yml`. `config.yml` and the `db/` SQLite dir are gitignored (local-only).
- Default `server.mode: debug` uses **SQLite** at `./db/dujiao.db` (auto-migrated; no external DB needed) and auto-seeds a default admin **`admin` / `admin123`**. Set `DJ_DEFAULT_ADMIN_USERNAME` / `DJ_DEFAULT_ADMIN_PASSWORD` to override.
- Health check: `GET /health` -> `{"status":"ok"}`.
- Start modes: default `-mode all` runs HTTP API **and** the async worker. `-mode api` runs the API only (use this if you want to avoid Redis).

### Redis is required for the default run mode (non-obvious)
- The default `-mode all` worker fails to start if `queue.enabled=true` but Redis is down, and cache/queue both default to Redis at `127.0.0.1:6379`.
- Redis is preinstalled but **not auto-started**. Start it with persistence disabled to avoid the `MISCONF` bgsave-to-disk error:
  ```bash
  redis-server --save "" --appendonly no --daemonize yes
  ```
- Gotcha: when Redis can't persist (RDB `MISCONF` error), backend write paths fail and the admin login UI shows a generic "用户服务不可用" (service unavailable) error. Disabling persistence as above prevents it.

### Tests / vet
- `go vet ./...` and `go build ./...` pass cleanly.
- Test suite: `go test ./...`.
