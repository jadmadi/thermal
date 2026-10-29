# Evidence: cli-argument-parser-strictness

Date: 2026-10-29
Project: thermal-streak
Goal: cli-argument-parser-strictness

## Implementation Summary
- Added `cmd/thermal/args.go`: `normalizeArgs`, the `flagMetas` value table, `usageError`/`retiredError`, `parseArgsFrom`, `nearestName`, and `levenshtein`.
- Replaced the split stdlib-flag/manual-scanner parser in `cmd/thermal/main.go` with `parseArgsFrom`. Flags are now interspersed with positionals and parsed once.
- Added `LastSet` and `TopSet` presence fields to `thermal.Options`.
- Invalid values now fail before any data read: `--last abc`, `--last 1.5`, `--last 0`, `--weeks abc`, `--weeks 3`, `--top xyz`, `--top -1`, `--interval abc`, `--interval 0`, `--flame-ceiling abc`, `--flame-ceiling 0`.
- Unknown flags are rejected with a nearest-match suggestion; `--metirc` suggests `--metric`.
- Positional/flag conflicts (`opencode --tool codex`), empty values (`--tool=`), extra positionals (`daily extra third`), and reversed date ranges (`--since` after `--until`) are rejected.
- Usage errors exit `2` with at most four lines: cause, fix/example, and help pointer. No 84-line manual is printed for value errors.
- Updated `cmd/thermal/simulated_user_test.go`, `scripts/simulated_user_gate.sh`, `scripts/output_gallery_gate.sh`, and regenerated `docs/OUTPUT-GALLERY.md`.
- Added README argument-validation notes, a v0.22.0 migration entry, and a `project.jsonc` feature line.

## Commands Executed
- `go test ./...` → PASS (all packages).
- `go test -race ./cmd/thermal/...` → PASS.
- `go test ./cmd/thermal -run=FuzzParseArgsStrict -fuzz=FuzzParseArgsStrict -fuzztime=10s -count=1` → PASS; 276,710 executions, 200 interesting inputs, no panic.
- Strict CLI matrix (13 invalid invocations) → `STRICT_MATRIX_PASS`: every case exit `2`, empty stdout, ≤4 stderr lines, no `flag provided`, `parse error`, or `strconv.` text.
- `./scripts/output_gallery_gate.sh` → PASS (`validation exits are stable` at `2/2/2/0`).
- `./scripts/simulated_user_gate.sh` → PASS (77/77 checks).
- `./scripts/check.sh --pre-commit` → PASS (4s; gofmt, package tests, binary build).

## Acceptance Criteria

- AC-001: `thermal daily --last abc` → exit `2`; stdout empty; stderr `thermal: --last wants a whole number of days, got "abc"` + example; no `Thermal · daily` header. **PASS**
- AC-002: `--last 1.5`, `--last -1`, `--last 0`, `--weeks abc`, `--top xyz`, `--interval abc` → each exit `2` with a flag-specific message and no report. **PASS**
- AC-003: `thermal --weeks abc` → exit `2`; stderr 2 lines; no `parse error`; no global manual. **PASS**
- AC-004: `thermal opencode --tool codex daily` → exit `2`, `positional tool "opencode" conflicts with --tool codex`; `opencode --tool opencode daily` → exit `0`. **PASS**
- AC-005: `thermal --tool= daily` and `thermal --tool "" daily` → exit `2`, `--tool requires a non-empty value`. **PASS**
- AC-006: `thermal daily --since 2026-10-29 --until 2026-10-01` → exit `2`, both dates named, swap suggestion present. **PASS**
- AC-007: `thermal daily extra` → exit `2` with `unknown tool: extra`; `thermal daily extra third` → exit `2` with `unexpected argument "third"`. **PASS**
- AC-008: `thermal daily --metirc cost` → exit `2`, `unknown flag --metirc`, `did you mean --metric?`; `cost` is not blamed. **PASS**
- AC-009: `thermal daily --last --json` → exit `2`, `--last needs a value`; `--json` is not consumed as a value. **PASS**
- AC-010: Every sampled error begins `thermal: ` and contains none of `flag provided but not defined`, `parse error`, `strconv.`. **PASS**
- AC-011: `go test -race ./cmd/thermal/...` → PASS; fuzz → PASS with 276,710 executions. **PASS**
- AC-012: `./scripts/simulated_user_gate.sh` → PASS (77/77). **PASS**

## Regression Notes
- Valid invocations unchanged: `daily --last 7`, `opencode weekly --chart`, `--json`, `--weeks=12`, compact dates, `-v daily`, `--tool opencode daily`.
- The output gallery's validation exits were updated from `1` to `2` because usage errors are now consistently `2`.
- The retired-command paths (`replay`, `yield`, `dashboard`) still exit `1` with their specific retirement messages.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-argument-parser-strictness",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "flag package stops at first positional silent numeric fallback", "matches": 5}],
  "memos": [],
  "lessons": [{"id": "fact_4acd3aca", "symptom": "CLI silently ignored invalid values after the first positional"}],
  "cases": [],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
