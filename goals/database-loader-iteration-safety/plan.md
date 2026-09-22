# Plan: Robust SQLite Rows Iteration Safety and rows.Err Checking

## Execution Steps
- [x] **Task 1 — Identify Row Iteration Sites**: Audit all `for rows.Next()` call sites across SQLite loaders.
- [x] **Task 2 — Add Error Checks**: Insert `if err := rows.Err(); err != nil { return ..., err }` after the iteration loop in `codex.go`, `hermes.go`, and `muse.go`.
- [x] **Task 3 — Unit Test Validation**: Run all loader unit tests with race detection to verify error propagation and clean execution.
- [x] **Task 4 — Release Gate Verification**: Run `./scripts/simulated_user_gate.sh` to ensure zero regressions.
- [x] **Task 5 — Continuity**: Record architectural lesson in sila (`fact_78743714`).
