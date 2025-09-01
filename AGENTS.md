AGENTS.md for maya-canteen

Build / Lint / Test
- Build backend: make build
- Run server: make run
- Run all Go tests: make test
- Run single Go test: go test ./... -run TestName -v (from repo root) or go test ./internal/... -run TestName -v
- DB integration tests: make itest
- Frontend dev: (cd frontend && pnpm dev) or npm/yarn as configured
- Lint frontend: (cd frontend && pnpm lint)

Code Style Guidelines
- Formatting: gofmt/gofmt -s / goimports for Go; run go fmt via make or go fmt ./...
- Imports: group std lib, third-party, internal (use goimports to auto-fix). In TS, group: builtins, external, aliases, local.
- Types: prefer concrete types in Go models; use interfaces for abstractions. In TS, type props and return values; prefer readonly where appropriate.
- Naming: Go uses camelCase for vars, PascalCase for exported symbols. TS uses camelCase for vars/props, PascalCase for React components and types.
- Error handling: return errors up, wrap with fmt.Errorf("context: %w", err) or use errors package. Check and handle DB errors explicitly.
- Logging: use structured logs in server (internal/server/logging.go). Avoid fmt.Println in production code.
- Tests: keep tests deterministic; use table-driven tests in Go; clean up DB state between tests.

Agent rules
- Follow .github/copilot-instructions.md architecture and workflows.
- No Cursor rules found in repository; if added, respect them.

Short checklist for agents
- Run go test -run <TestName> -v to execute a single test.
- Use goimports and gofmt before committing.
- When modifying frontend, run frontend lint and build.

