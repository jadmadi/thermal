# Evidence: cli-help-interception-and-scoped-help

Date: 2026-10-29
Project: thermal-streak
Goal: cli-help-interception-and-scoped-help

## Implementation Summary
- Added `cmd/thermal/help.go`: a command registry (`commandHelps`), tool/alias pages, `resolveHelpTarget`, `helpOutput`, and `printHelpFor`.
- Intercepted help at the top of `parseArgs` before `parseArgsFrom`, the version/license special cases, and every command dispatch.
- Added `internal/render.RenderCommandHelp` with a plain, width-safe page anatomy (Usage, Description, Flags, Examples, JSON, Note, See also).
- Added `help`, `help <command>`, and `<command> --help` forms; `thermal help <command>` is byte-identical to `thermal <command> --help`.
- Added tool pages and tool+verb pages (`thermal opencode --help`, `thermal opencode weekly --help`).
- Added alias-aware statusline pages (`prompt`, `status`) and retired-verb pages (`dashboard`, `serve`, `web`, `replay`, `yield`, `receipt`).
- Unknown help contexts exit `1` with a nearest-command suggestion; `today` → `daily`.
- Help never reads data, opens a TUI, writes a cache, calls the network, or mutates the binary. `upgrade --help` and `live --help` are safe.
- Updated the global help with `help`, usage, examples, and scoped-help wording.
- Added `cmd/thermal/help_test.go` and a simulated-gate help section.

## Commands Executed
- `go test ./...` → PASS (all packages).
- `go test -race ./...` → PASS (all packages).
- `./scripts/check.sh --pre-commit` → PASS (6s; gofmt, package tests, binary build).
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (86/86 checks).
- Verb-page loop (`daily weekly monthly projects models stats mix trend live statusline share audit changelog upgrade version license`) → all exit `0`, all contain `Usage:`.
- `cmp <(thermal daily --help) <(thermal help daily)` → identical; same for `weekly`, `live`, and `upgrade`.
- PTY `live --help` → exit `0`, no `\x1b[?1049h`, no `THERMAL LIVE` frame, `Usage:` present.
- Disposable `upgrade --help` → md5 unchanged (`69437d7e…` before and after), no `.old`/`.new`.
- Empty HOME `thermal --help` and `thermal upgrade --help` → exit `0`, stderr empty, zero new files.
- `NO_COLOR=1 thermal daily --help` → 0 ANSI escapes; 0 lines over 80 columns.

## Acceptance Criteria

- AC-001: 16 verb pages (`daily` through `license`, including `help`) exit `0`, write to stdout, and contain `Usage:`. **PASS**
- AC-002: `thermal daily --help` contains `--since`, `--until`, `--last`, `--breakdown`, `--chart`, `--json`; it contains no `Daily Usage` table and no date row. **PASS**
- AC-003: `daily`, `weekly`, `live`, and `upgrade` `--help` output is byte-identical to `thermal help <verb>`. **PASS**
- AC-004: `thermal live --help` in a 100x30 PTY exits without alternate-screen sequences and without a `THERMAL LIVE` frame. **PASS**
- AC-005: A disposable binary's md5 is unchanged after `upgrade --help`; no `.old`/`.new` artifacts. **PASS**
- AC-006: Empty HOME `thermal --help` and `thermal upgrade --help` exit `0`, produce no data error, and create zero files. **PASS**
- AC-007: `thermal help nope` exits `1` with `unknown command "nope"` and the command-list pointer; `thermal today --help` suggests `thermal daily`. **PASS**
- AC-008: `thermal help help` exits `0` and documents both `thermal help [command]` and `thermal <command> --help`. **PASS**
- AC-009: `thermal --json daily --help` and `thermal daily --help --json` print help, exit `0`, and stdout is not JSON. **PASS**
- AC-010: `NO_COLOR=1` help has 0 ANSI escapes; no line exceeds 80 columns. **PASS**
- AC-011: `go test -race ./...` passes, including `TestResolveHelpTarget`, `TestHelpPagesFitEightyColumns`, `TestHelpPageSpecifics`, `TestHelpAliasesAndTools`, `TestHelpUnknownAndRetired`, and `TestHelpHelpPage`. **PASS**
- AC-012: `./scripts/simulated_user_gate.sh` passes 86/86 and includes the scoped-help section. **PASS**

## Regression Notes
- `thermal --help` remains the global manual; it now includes `help` and scoped-help examples.
- `thermal -v` and `thermal version` still print version metadata (the version/verbose contract is a separate goal).
- `thermal upgrade --help` no longer upgrades; the destructive path requires the upgrade goal's explicit flags.
- Data commands are unchanged when no help token is present.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-help-interception-and-scoped-help",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Intercept help before dispatch and render command-scoped pages", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_71905b60", "intent": "Intercept help before dispatch and render command-scoped pages"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
