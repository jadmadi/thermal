# Plan: Make transcript verification independent of ambient tool-root overrides

## Execution Steps
- [x] **Step 1 — Enumerate the blast radius**: Re-run the poisoned-environment command from `facts.md` and confirm the current failing set (`TestScanSessionReceipts_Codex` for `CODEX_HOME`; `TestLoadUserConfig_*` and `TestAllTools_ConfigOverrides` for `THERMAL_CONFIG`). Extend the sweep with the remaining overrides (`GROK_HOME`, `DSH_HOME`, `HERMES_HOME`, `AGY_HOME`, `ANTIGRAVITY_APP_DATA_DIR`, `OPENCODE_HOME`, `ZED_HOME`, `APPDATA`, `LOCALAPPDATA`, `XDG_*`) to catch anything the first pass misses.
  Acceptance and verification: failure table checked into the goal evidence with the exact command.
- [x] **Step 2 — Isolate the receipt test**: In `internal/thermal/receipt_test.go`, make `TestScanSessionReceipts_Codex` (and any sibling scan test) set the ambient override explicitly (`t.Setenv("CODEX_HOME", "")` for tests that use their own home, or point it at the test fixture when the override is the subject). Audit the rest of the file for other env-reading paths.
  Acceptance and verification: `CODEX_HOME=/tmp/empty go test ./internal/thermal -run TestScanSessionReceipts_Codex` passes; the assertion still requires exactly one Codex receipt from the fixture.
- [x] **Step 3 — Isolate the config tests**: Apply the same treatment for `THERMAL_CONFIG` in `internal/loaders` (`t.Setenv("THERMAL_CONFIG", "")` or an explicit fixture path per test).
  Acceptance and verification: `THERMAL_CONFIG=/nonexistent go test ./internal/loaders -run 'LoadUserConfig|AllTools_ConfigOverrides'` passes.
- [x] **Step 4 — Re-run the poisoned sweep**: Rerun the full poisoned command and the normal `go test -race ./...`.
  Acceptance and verification: both exit 0; no test was deleted or relaxed (review the diff for assertion changes).
- [x] **Step 5 — Attribution commit**: Commit with a `test:` subject ending in `(goals/test-env-isolation/goal.md)` plus a `chore:` closure commit if the tracker ignores `test:`; then `sila goals`. Document the poisoned-run command in the goal evidence.
  Acceptance and verification: commit attributed; poisoned command recorded with exit 0.

## Verification
- Poisoned sweep: `env CODEX_HOME=… GROK_HOME=… DSH_HOME=… HERMES_HOME=… ANTIGRAVITY_APP_DATA_DIR=… THERMAL_CONFIG=… go test ./... -count=1`
- `go test -race ./internal/thermal ./internal/loaders`
- Common final checks from `tasks/plan.md`

## Notes
- `sila` closure ignores `test:` and `docs:` commits; pair the test change with a `fix(...)`/`chore(...)` commit so the goal can close, or use the tracker's `--no-build` commit path.
