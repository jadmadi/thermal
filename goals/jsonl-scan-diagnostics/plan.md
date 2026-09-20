# Plan: A scan that stops early is reported, never silent

## Execution Steps
- [ ] **Task 1 — Facts**: the eight unchecked scan sites, the fatal-only warning channel, the stderr and JSON contract, and the chosen `Summary.Warnings` mechanism. (This file.)
- [ ] **Task 2 — Shared scanner and formatter**: the shared constructor from `jsonl-line-ceilings` carries the ceiling as a package variable so tests can lower it; a helper turns `scanner.Err()` into the warning text, naming the file and saying the rest of the file was skipped on `bufio.ErrTooLong`.
- [ ] **Task 3 — Wire every site**: all eight scan loops check `scanner.Err()` and record a warning; the worker pools merge per-file warnings after `wg.Wait()`; an `os.Open` failure records the same kind of warning instead of an empty result with no note.
- [ ] **Task 4 — Print under --verbose**: `Summary.Warnings` with `json:"-"`; `cmd/thermal` prints `thermal: warning: <tool>: <text>` to stderr only when `--verbose` is set.
- [ ] **Task 5 — Tests**: with the ceiling lowered, a file with an over-long line produces one warning naming the file and the ceiling, keeps the data read before it, and counts nothing after it; a clean file produces none; `--json` stays identical and non-verbose stderr stays empty.
- [ ] **Task 6 — Record the rule**: AGENTS.md: every JSONL scan checks `Err()`; diagnostics ride `Summary.Warnings` to verbose stderr, never stdout or JSON.
- [ ] **Task 7 — Commit and propose**: `feat(loaders): ... (goals/jsonl-scan-diagnostics/goal.md)`, pull request into `main`, no release.
