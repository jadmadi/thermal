# Plan: Defensive Timestamp Parsing and Malformed Date Guards in Streak Computation

## Execution Steps
- [x] **Task 1 — Date Parse Guards**: In `internal/thermal/streak.go`, check `err` from `time.Parse` in the historical run loop (`if err != nil { continue }`).
- [x] **Task 2 — Cursor Parse Guard**: Guard `cursor, err := time.Parse(...)` in the current streak walkback loop.
- [x] **Task 3 — Unit Test Validation**: Verify `go test -race ./internal/thermal/...` to confirm streaks and thresholds tests pass.
- [x] **Task 4 — Simulated User Gate**: Run `./scripts/simulated_user_gate.sh` to ensure release gate integrity.
