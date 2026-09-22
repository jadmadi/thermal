# Goal: Nous Hermes Agent Loader

## Goal Description
Implement an ingestion loader for **Nous Hermes** (`hermes`), an open-source, local-first persistent coding agent developed by Nous Research.

Hermes stores all local conversation histories, session metadata, tool executions, and turn token counts in a local SQLite database located at `~/.hermes/state.db`.

In Tyler Southwick's (@btsouth) multi-tool AI coding history scan using StackReplay, Hermes was one of the five primary local harnesses analyzed (alongside Claude Code, Codex, Command Code, and OpenCode).

Adding Hermes expands Thermal's supported tools from 13 to 14, providing complete coverage of all local agent harnesses tracked in current developer workflows.

This goal adds:
1. `internal/loaders/hermes.go` reading `~/.hermes/state.db` in strictly read-only mode (`?mode=ro`) with memory-mapped I/O (`PRAGMA mmap_size`).
2. Disjoint token mapping (Input, Output, Reasoning, CacheRead, CacheWrite), turn counts, and model tracking.
3. Project attribution via session working directories normalized through `thermal.ProjectKey`.
4. CLI registration (`hermes`, aliases: `nous`, `nous-hermes`) in `internal/thermal/types.go`, `internal/loaders/registry.go`, and `cmd/thermal/main.go`.
5. Unit tests with in-memory SQLite fixtures in `internal/loaders/hermes_test.go`.
6. Documentation updates across `README.md`, `AGENTS.md`, and HTML docs.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/hermes-loader/facts.md`](facts.md)
- **Execution Plan**: [`goals/hermes-loader/plan.md`](plan.md)

## Done Condition
1. `hermes.go` successfully scans `~/.hermes/state.db` (or custom `--db` path) in read-only mode without file locking.
2. `thermal --tool hermes` renders streak, heatmap, session count, tokens, and model breakdown cleanly.
3. Unit tests in `hermes_test.go` achieve >90% coverage for the loader with edge case handling (missing DB, zero-token sessions, schema variations).
4. `go test -v -race ./...` passes with zero race conditions.
5. `golangci-lint` passes with zero issues.
6. Documentation (`README.md`, `AGENTS.md`, `docs/pages/index.html`, `docs/pages/404.html`) updated with full parity.
