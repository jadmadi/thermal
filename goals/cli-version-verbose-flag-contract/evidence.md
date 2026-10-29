# Evidence: cli-version-verbose-flag-contract

Date: 2026-10-29
Project: thermal-streak
Goal: cli-version-verbose-flag-contract

## Implementation Summary
- Removed the `len(os.Args) == 2` version special case that made `-v` context-dependent.
- Added `cmd/thermal/version.go`: `versionOutput`, `printVersion`, `resolveVersionTarget`, and `hasJSONFlag`.
- `-V`, `--version`, and the first positional `version` now print metadata in any position; `-v`/`--verbose` enable diagnostics.
- Added defense-in-depth `-V`/`--version` flags and `errVersionRequested` in the parser, so neither spelling can become an unknown-flag error.
- Version wins over validation and execution (`-V daily`, `daily -V`, `-v -V`, `-V --bogus`); help wins over version (`-V --help`).
- `-V --json`, `--version --json`, and `version --json` share one payload; `-v --json` remains the resolved command's JSON.
- Updated the global help and version page with `-V, --version` and the `-v` distinction.
- Added `cmd/thermal/version_test.go`, a simulated-gate version section, a migration entry, README note, and project metadata.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./...` → PASS.
- `go test -race ./cmd/thermal/...` → PASS.
- `./scripts/check.sh --pre-commit` → PASS (2s; gofmt, package tests, binary build).
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (95/95 checks).
- Matrix:
  - `thermal -v --no-color` → leaderboard containing `Token Warriors`; no `thermal dev` line; exit `0`.
  - `thermal -v --json` → `.type == "leaderboard"`.
  - `thermal -V --json` → `.type == "version"`.
  - `thermal -V daily` → `thermal dev (built from source)`; no `Daily Usage`.
  - `thermal -v -V` and `thermal -V --bogus` → version, exit `0`.
  - Empty HOME `thermal -v` → exit `1`, `thermal: no supported tool data found`, no version text.
  - `thermal --help` contains `-V, --version` and `--verbose, -v`; no `-v, --version` contradiction.

## Acceptance Criteria

- AC-001: `thermal -v` exits `0` and prints the leaderboard (`Token Warriors`), not version metadata. **PASS**
- AC-002: `thermal -v daily` and `thermal daily -v` exit `0` and print the daily report. **PASS**
- AC-003: `thermal -V` exits `0` and prints `thermal dev (built from source)`. **PASS**
- AC-004: `thermal --version` and `thermal version` are byte-identical to `-V`. **PASS**
- AC-005: `thermal -V --json | jq -e '.type=="version"'` exits `0`; `version --json` emits the same keys/values apart from `generatedAt`. **PASS**
- AC-006: `thermal -v --json` emits leaderboard JSON (`.type == "leaderboard"`), not a version object. **PASS**
- AC-007: `thermal --help` contains `-V, --version` and `-v, --verbose`; it does not describe `-v` as version. **PASS**
- AC-008: `thermal -v -V` prints version and exits `0`. **PASS**
- AC-009: `thermal -V daily` prints version and exits `0`; the daily table is absent. **PASS**
- AC-010: `thermal -V --bogus` prints version and exits `0` (documented version > validation precedence). **PASS**
- AC-011: `HOME=$(mktemp -d) thermal -v` exits `1` with the empty-state message and no version output. **PASS**
- AC-012: `go test -race ./...` passes, including `TestResolveVersionTarget`, `TestVersionOutput`, `TestVerboseIsNotVersion`, `TestVersionFlagParserDefense`, `TestVersionJSONParity`, and `TestVersionHelpPage`. **PASS**
- AC-013: `./scripts/simulated_user_gate.sh` passes 95/95 and includes the version/verbose section. **PASS**

## Regression Notes
- `thermal -v` is the intentional breaking change: it now means verbose. MIGRATION documents the switch to `-V`/`--version`.
- `thermal version` and `thermal --version` remain stable.
- Help precedence is unchanged: `-V --help` prints help, not version.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-version-verbose-flag-contract",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Give -v, -V, --version, and --verbose one meaning", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_de38c59d", "intent": "Give -v, -V, --version, and --verbose one meaning"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
