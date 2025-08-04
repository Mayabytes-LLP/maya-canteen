# Copilot Instructions for maya-canteen

## Overview

maya-canteen is a full-stack canteen management system with a Go backend and a React/TypeScript (Vite) frontend. The backend handles business logic, data persistence, and integrations (e.g., WhatsApp, WebSocket), while the frontend provides a modern UI for users and admins.

## Architecture & Key Components

- **Backend (Go):**
  - Entry point: `cmd/api/main.go`
  - Business logic: `internal/handlers/`, `internal/models/`, `internal/database/`
  - Routing: `internal/server/routes/`
  - Middleware: `internal/middleware/`
  - Integrations: WhatsApp (`internal/handlers/whatsapp_handler.go`), WebSocket (`internal/handlers/websocket_handler.go`)
  - Data: SQLite DB (`db/canteen.db`), repository pattern in `internal/database/repository/`
- **Frontend (React + Vite):**
  - Source: `frontend/src/`
  - UI components: `frontend/src/components/`
  - Pages: `frontend/src/pages/`
  - State/context: `frontend/src/context/`
  - Utilities: `frontend/src/lib/`

## Developer Workflows

- **Build & Run (Backend):**
  - `make build` — Build Go backend
  - `make run` — Run backend server
  - `make watch` — Live reload backend
  - `make docker-run` / `make docker-down` — Start/stop DB container
  - `make itest` — DB integration tests
  - `make test` — Run Go tests
  - `make clean` — Clean build artifacts
- **Frontend:**
  - Uses Vite for dev/build. See `frontend/README.md` for ESLint and plugin details.
  - Main entry: `frontend/src/main.tsx`, app: `frontend/src/App.tsx`

## Project Conventions & Patterns

- **Backend:**
  - Repository pattern for DB access (`internal/database/repository/`)
  - Handlers are grouped by domain (product, transaction, user, etc.)
  - Models in `internal/models/` map directly to DB tables
  - Errors centralized in `internal/errors/`
  - Use middleware for cross-cutting concerns
- **Frontend:**
  - Components are colocated by domain (e.g., `dashboard/`, `product/`)
  - Context for app-wide state in `src/context/`
  - Use React Query for data fetching (`src/lib/react-query.ts`)
  - UI built with modern React patterns (hooks, context, composition)

## Integration Points

- **WhatsApp:** Automated notifications via `internal/handlers/whatsapp_handler.go`
- **WebSocket:** Real-time updates via `internal/handlers/websocket_handler.go`
- **Database:** SQLite, managed via Go repository pattern

## Examples

- Add a new API route: see `internal/server/routes/`
- Add a new frontend page: create in `frontend/src/pages/` and route via main app
- Add a new DB model: define in `internal/models/`, add repository, handler, and route

## References
- Main backend entry: `cmd/api/main.go`
- Main frontend entry: `frontend/src/main.tsx`
- Build/test commands: see root `README.md`

---

For questions, check the relevant `README.md` or ask for examples from the codebase.
