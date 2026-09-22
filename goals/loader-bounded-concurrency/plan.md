# Plan: Bounded Worker Concurrency in Multi-File Loaders

## Execution Steps
- [x] **Task 1 — Audit Worker Loops**: Identify all multi-file log scanners using semaphore channels and goroutines across `internal/loaders/`.
- [x] **Task 2 — Restructure Semaphore Acquisition**: Move `sem <- struct{}{}` outside the spawned `go func()` across `agy.go`, `claude.go`, `codex.go`, `commandcode.go`, `droid.go`, `dsh.go`, and `grok.go`.
- [x] **Task 3 — Concurrency & Race Verification**: Run `go test -race ./internal/loaders/...` to ensure zero deadlocks and zero data races.
- [x] **Task 4 — Release Gate Verification**: Run `./scripts/simulated_user_gate.sh` to confirm end-to-end command fidelity.
- [x] **Task 5 — Continuity**: Record architectural lesson in sila (`fact_c387c399`).
