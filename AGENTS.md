# maya-canteen

Full-stack canteen management: Go backend (gorilla/mux, SQLite) + React/Vite frontend (shadcn/ui, Tailwind v4). Integrates with ZK fingerprint devices and WhatsApp.

## Build & Run

- `make build` — builds Go binary at `./main`. Requires `CGO_ENABLED=1` (sqlite3 driver).
- `make run` — starts backend (port 8080) and frontend dev server concurrently.
- `make watch` — hot-reload backend via `air`.
- `pnpm build` (in `frontend/`) — `tsc -b && vite build`. Must run **before** `make build` for production (frontend embedded via `//go:embed dist`).
- `pnpm dev` (in `frontend/`) — Vite dev server with proxy: `/api` → `localhost:8080`, `/socket.io` → WS.
- `pnpm lint` (in `frontend/`) — ESLint.

## Test

- `make test` (or `go test ./... -v`) — all Go tests.
- `make itest` — DB integration tests (requires Docker for SQLite container).
- Single test: `go test ./internal/... -run TestName -v`

## Architecture

| Layer | Path | Key details |
|-------|------|-------------|
| Entrypoint | `cmd/api/main.go` | Starts ZK device, WhatsApp, HTTP server |
| Routing | `internal/server/routes/` | gorilla/mux + middleware chain (CORS, Logger, Recover) |
| Handlers | `internal/handlers/` | Domain-grouped (product, transaction, user, whatsapp, websocket) |
| DB | `internal/database/` | SQLite via mattn/go-sqlite3, repository pattern in `repository/` |
| Models | `internal/models/` | User, Product, Transaction, TransactionProduct, UserBalance |
| Frontend | `frontend/src/` | React 19, React Router v7, TanStack React Query, shadcn/ui |
| Errors | `internal/errors/` | Centralized error types |
| Middleware | `internal/middleware/` | CORS, Logger, Recover |
| ZK device | `internal/gozk/` + `internal/handlers/zk_device.go` | Fingerprint scanner over TCP |
| WhatsApp | `internal/handlers/whatsapp.go` | whatsmeow library, session in `whatsapp-store.db` |

## Environment

Key vars (see `.env.example`):
- `PORT` (default: 8080), `BLUEPRINT_DB_URL` (default: `D:/database/canteen.db`)
- `ZK_PORT` (4370), `ZK_IP` (192.168.1.153), `ZK_TIMEZONE`
- `VITE_API_BASE_URL`, `VITE_WS_URL` — for frontend

`godotenv/autoload` loads `.env` automatically. `mise.toml` pins Go, Node, pnpm versions.

## Conventions

- **Go**: `gofmt` / `goimports` (std → third-party → internal groups). Log with logrus, not `fmt.Println`. Wrap errors with `fmt.Errorf("context: %w", err)`.
- **Tests**: table-driven, deterministic, clean DB state between tests.
- **Frontend**: PascalCase for components/types, camelCase for vars/props. Components colocated by domain under `components/`. Pages in `pages/`. App state in `context/`.
- **API**: gorilla/mux router, handlers return JSON via `common.JSONResponse`. SPA fallback for non-API routes.
- Version injected at build via ldflags: `maya-canteen/internal/version`.
- Graceful shutdown handles ZK → WhatsApp → HTTP server tear-down.
