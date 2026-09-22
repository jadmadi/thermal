# Goal: Robust SQLite Rows Iteration Safety and rows.Err Checking

## Goal Description
Database queries in `database/sql` iterate results using `for rows.Next() { ... }`. When an iteration ends, `rows.Next()` returns `false` both when all rows have been successfully exhausted and when an underlying network, serialization, or SQLite error truncates the query. Without checking `rows.Err()`, database loaders silently proceed with partial records and drop remaining sessions.

This goal adds explicit `if err := rows.Err(); err != nil` inspection to database loaders that lacked it (`codex`, `hermes`, `muse`), preventing silent session truncation and guaranteeing data integrity.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- `internal/loaders/codex.go`
- `internal/loaders/hermes.go`
- `internal/loaders/muse.go`
- `AGENTS.md`

## References
- **Shared Understanding & Fact Sheet**: [`goals/database-loader-iteration-safety/facts.md`](facts.md)
- **Execution Plan**: [`goals/database-loader-iteration-safety/plan.md`](plan.md)

## Done Condition
1. All SQLite database loaders inspect `rows.Err()` following row iteration loops.
2. Any query truncation or SQLite read error halts loader execution and propagates the non-nil error.
3. Unit tests pass cleanly with race detection: `go test -race ./internal/loaders/...`.
4. Release gate passes: `./scripts/simulated_user_gate.sh`.
