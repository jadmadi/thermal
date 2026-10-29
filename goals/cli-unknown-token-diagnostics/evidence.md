# Evidence: cli-unknown-token-diagnostics

Date: 2026-10-29
Project: thermal-streak
Goal: cli-unknown-token-diagnostics

## Implementation Summary
- Added a slot-aware classifier in `cmd/thermal/diagnostics.go`: retired verbs, known commands, known tools, path-like tokens, empty tokens, synonyms, and fuzzy matches.
- Replaced the remaining `unknown tool` sites with `printUnknownToken`, which exits `2` for unknown/wrong-slot/path/empty tokens and `1` for retired verbs.
- Command typos are labeled as commands and suggest the nearest command (`today`/`daly` → daily, `weak` → weekly, `cost` → stats).
- Report words in the tool slot are named (`"weekly" is a report command; run thermal weekly`).
- Path-like tokens are explained and pointed at `--db`.
- Unknown tools show a nearest tool and the valid tool list.
- Retired verbs keep their specific message and now include `See also: thermal live` for `dashboard`.
- Added a strict empty-positional rejection in the argument parser.
- Added `cmd/thermal/diagnostics_test.go` for classification, rendering, line budget, and vocabulary coverage.
- Updated README, migration guide, project metadata, and the simulated gate.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./...` → PASS.
- `./scripts/check.sh --pre-commit` → PASS (1s; gofmt, package tests, binary build).
- `./scripts/simulated_user_gate.sh` → PASS (111/111 checks).
- `thermal today` → exit `2`, `unknown command "today"`, `did you mean: thermal daily?`, no `unknown tool`; 5 lines.
- `thermal daly` and `thermal weak` → exit `2`, suggestions `daily` and `weekly`.
- `thermal stats weekly` → exit `2`, `"weekly" is a report command`, `run thermal weekly`.
- `thermal projects /tmp` → exit `2`, `looks like a path`, `use --db <path>`.
- `thermal --tool nope` → exit `2`, `unknown tool "nope"`, `did you mean: "nous"?`, valid tool list.
- `thermal cost` → exit `2`, `unknown command "cost"`, `did you mean: thermal stats?`.
- `thermal opnecode` → exit `2`, `unknown tool "opnecode"`, `did you mean: "opencode"?`.
- `thermal a` → exit `2`, no confident suggestion, valid tool list.
- `thermal dashboard` → exit `1`, retirement message, `See also: thermal live`.
- `thermal opencode`, `thermal --tool all`, and `thermal daily` → exit `0` (valid paths).
- `thermal ""` → exit `2`, `empty argument is not a command or tool`.

## Acceptance Criteria

- AC-001: `thermal today` exits `2`, stdout empty, contains `unknown command "today"` and `Did you mean: thermal daily?`; it does not contain `unknown tool`. **PASS**
- AC-002: `thermal daly` suggests `daily`; `thermal weak` suggests `weekly`. **PASS**
- AC-003: `thermal stats weekly` exits `2`, contains `"weekly" is a report command`, suggests `thermal weekly`, and does not contain `unknown tool`. **PASS**
- AC-004: `thermal projects /tmp` exits `2`, contains `looks like a path`, mentions `--db`, and does not suggest a tool. **PASS**
- AC-005: `thermal --tool nope` exits `2`, contains `unknown tool "nope"`, lists at least five tools, and suggests only a tool (`nous`). **PASS**
- AC-006: `thermal --tool opencode`, `thermal opencode`, `thermal daily`, and `thermal --tool all` exit `0`; the gate uses the fixture-backed `agy` instead of OpenCode. **PASS**
- AC-007: `thermal dashboard` keeps the retirement message, exits `1`, and includes `See also: thermal live`. **PASS**
- AC-008: `thermal a` exits `2` with the tool list and no confident suggestion (the fuzzy threshold requires at least three characters). **PASS**
- AC-009: Every sampled diagnostic is ≤5 lines, starts with `thermal: `, and writes nothing to stdout. **PASS**
- AC-010: Diagnostics are plain text with no ANSI; `NO_COLOR` is moot but respected. **PASS**
- AC-011: `go test -race ./...` passes, including `TestClassifyToken`, `TestRenderTokenDiagnostic`, and `TestDiagnosticCandidateCoverage`. **PASS**
- AC-012: `./scripts/simulated_user_gate.sh` passes 111/111 and includes the diagnostics section. **PASS**

## Regression Notes
- Valid commands and tools are untouched; the classifier runs only after exact resolution fails.
- The command/tool vocabularies are covered by a test that fails when a new command or tool is added without a candidate entry.
- The diagnostics remain stderr-only and never contaminate JSON.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-unknown-token-diagnostics",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Classify unknown CLI tokens by slot and suggest commands or tools", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_a51f79ab", "intent": "Classify unknown CLI tokens by slot and suggest commands or tools"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
