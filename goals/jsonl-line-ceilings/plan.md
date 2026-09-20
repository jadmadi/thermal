# Plan: JSONL line ceilings that fit real logs

## Execution Steps
- [ ] **Task 1 — Pin the facts and the repro**: record the live file, the byte counts, the before numbers from `thermal command-code --json`, and the scanner table. Facts already hold the commands; repeat them after the fix.
- [ ] **Task 2 — Add the shared ceiling**: one constructor and constant (32 MiB) in `internal/loaders`, adopted at every JSONL scan site: `codex.go` (`readLastTokenBreakdown`), `commandcode.go`, `grok.go`, `claude.go`, `droid.go`, `agy.go` (two sites), `jsonl.go`.
- [ ] **Task 3 — Tests that would have caught this**: generate a transcript with a line over the old 256 KiB cap and messages after it; the command-code loader must count every message, and the jsonl fallback must read a line over the 64 KiB default. Run the test once against the old ceiling and record the failure (red before green).
- [ ] **Task 4 — Verify on real data**: the file's 320 lines all count; total 10,707 -> 10,745; 2026-08-01 294 -> 308; 2026-08-02 113 -> 137. `go test -race ./...` and a build.
- [ ] **Task 5 — Record the rule**: AGENTS.md loader standards gain one line: every JSONL scan sets an explicit ceiling that fits real data, 32 MiB today; the diagnostics half is the sibling goal.
- [ ] **Task 6 — Commit and propose**: `fix(loaders): ... (goals/jsonl-line-ceilings/goal.md)`, pull request into `main`, no release.
