# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Full-stack canteen management app: Go backend (gorilla/mux, SQLite via mattn/go-sqlite3) + React 19/Vite frontend (shadcn/ui, Tailwind v4, React Router v7, TanStack Query). Integrates with a ZK fingerprint device (TCP) and WhatsApp (whatsmeow). See `AGENTS.md` for the architecture table, env vars, and code conventions.

## Commands

Backend (repo root):
- `make build` — builds `./main` (needs `CGO_ENABLED=1`; injects version via ldflags into `maya-canteen/internal/version`)
- `make run` — backend (:8080) + frontend dev server; `make watch` — hot reload with `air`
- `make test` / `go test ./... -v`; single test: `go test ./internal/handlers -run TestName -v`
- `make itest` — DB integration tests (requires Docker); `make build-windows` — cross-build with `go-winres`

Frontend (`frontend/`, pnpm; versions pinned in `mise.toml`):
- `pnpm dev` — Vite with proxy `/api` → `localhost:8080`, `/socket.io` → WS
- `pnpm build` — `tsc -b && vite build`; **must run before `make build` for production** because the Go binary embeds `frontend/dist` via `//go:embed`
- `pnpm lint` — ESLint

## Architecture notes (beyond the file layout)

- **Startup wiring**: `cmd/api/main.go` starts the ZK device and WhatsApp client, then `server.NewServer(whatsappClient)` → `routes.RegisterRoutes(db, whatsappClient)`. Shutdown (`internal/server/shutdown.go`) tears down ZK → WhatsApp → HTTP in that order.
- **WhatsApp client is swappable at runtime**: pairing/reconnect replaces the client, so `Server.UpdateWhatsAppClient` also pushes it into `routes.GlobalWebSocketHandler` (a package-level global). Handlers depend on the `handlers.WhatsAppClient` interface, which is what the `whatsapp_*_test.go` files mock.
- **WebSocket** (`internal/handlers/websocket_handler.go`) carries WhatsApp pairing flow (e.g. `pair_phone` messages with the phone nested in `payload`) and status events to the frontend; the frontend connects via `VITE_WS_URL`.
- **DB path**: resolved once at startup by `server.SetupDBPath()` (`internal/server/db_path.go`) and exported as `BLUEPRINT_DB_URL`; all DB access reads that env var. Default is a `db/canteen.db` beside the executable (the `.env.example` default differs, check both when debugging). WhatsApp session lives separately in `whatsapp-store.db`.
- **Data access**: handlers → `internal/database` service → `database/repository/*` (repository pattern). Business logic for transactions (balances, products) lives in `transaction_handler.go`.
- **Frontend**: all API calls go through `frontend/src/services/` (currently `transaction-service.ts`); global state in `context/app-provider.tsx`. SPA fallback on the Go side serves non-`/api` routes from the embedded `dist`.
- `internal/gozk/` is a vendored ZK protocol library (own LICENSE/README); treat as third-party-ish code.

## Conventions

- Go: `gofmt`/`goimports` (std → third-party → internal), logrus not `fmt.Println`, wrap errors with `%w`, JSON responses via `common.JSONResponse`.
- Go tests: table-driven, deterministic, clean DB state between tests.
- Frontend: PascalCase components/types, camelCase vars/props; components grouped by domain under `components/`, pages under `pages/`.
- Worktrees go under `~/projects/worktrees/maya-canteen/<name>/` (workspace rule).
