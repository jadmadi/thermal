# Plan: Nous Hermes Agent Loader

## Tasks
- [x] **Phase 1 — Schema Probing & Contracts**: Inspect and document table structure of `~/.hermes/state.db` (sessions, turns, model usage).
- [x] **Phase 2 — Ingestion Engine**: Implement `internal/loaders/hermes.go` with read-only SQLite connection, memory mapping, and delta cache.
- [x] **Phase 3 — Registry & CLI Integration**: Register `ToolHermes` in `types.go`, `registry.go`, and update `--help` in `cmd/thermal/main.go`.
- [x] **Phase 4 — Unit Tests**: Create `internal/loaders/hermes_test.go` with mock database fixtures, covering multi-day sessions and model breakdowns.
- [x] **Phase 5 — Full Test Suite Pass**: Verify with `go test -v -race ./...` and `golangci-lint`.
- [x] **Phase 6 — Documentation Parity**: Update `README.md`, `AGENTS.md`, `index.html`, and `404.html` with tool 14.
