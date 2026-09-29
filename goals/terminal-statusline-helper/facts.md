# Facts: Instant Terminal Statusline & Prompt Composable Helper

## Architectural Invariants & Constraints
- Execution time must strictly remain under 5ms (target <3ms warm).
- No interactive UI, no terminal size detection overhead, no background daemons.
- Strict zero-allocation formatting for shell prompt embedding.

## File & Interface Contracts
- Relevant files:
  - `cmd/thermal/main.go`
  - `internal/render/statusline.go`
  - `internal/render/statusline_test.go`
  - `docs/MIGRATION.md`
