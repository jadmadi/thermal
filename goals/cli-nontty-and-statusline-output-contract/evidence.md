# Evidence: cli-nontty-and-statusline-output-contract

Date: 2026-10-30
Project: thermal-streak
Goal: cli-nontty-and-statusline-output-contract

## Implementation Summary
- Refactored `loadUsage` to return `(usageSet, error)`. Foreground commands use `loadUsageOrExit`; the statusline calls `loadUsage` directly and degrades.
- `thermal statusline`/`prompt`/`status` with no data exits `0` with no stdout or stderr. `--json` emits `state:"no-data"` with zero values; with data it emits `state:"ok"`.
- `--verbose` on a no-data statusline prints one `statusline: no data (15 sources checked)` line to stderr.
- `thermal live` in a non-TTY writes guidance to stderr and exits `2`; stdout stays empty. `live --json` and `live --json --stream` remain machine-clean.
- `live --json --stream` emits periodic snapshots, ignores SIGPIPE on Unix, and returns on write errors, so `| head -1` exits cleanly.
- `share` with no data remains a deliberate stderr failure with exit `1`; `daily` and other reports keep the inventory on stderr and exit `1`.
- Added `statusline` state tests and a channel-contract simulated gate; updated the gallery live-non-TTY capture and the gate's Go cache/binary reuse.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./...` → PASS.
- `./scripts/check.sh --pre-commit` → PASS (1s).
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (131/131 checks).
- `thermal live </dev/null` → exit `2`, stdout 0 bytes, stderr `thermal: live needs a terminal` + three machine alternatives.
- `thermal live --json </dev/null | jq -r .type` → `live`, exit `0`.
- `thermal live --json --stream </dev/null | head -1 | jq -e '.type=="snapshot"'` → exit `0`, stderr empty.
- Empty HOME `thermal statusline` → exit `0`, stdout 0 bytes, stderr 0 bytes.
- Empty HOME `thermal statusline --plain` → exit `0`, stdout/stderr 0 bytes.
- Empty HOME `thermal statusline --json` → exit `0`, `state:"no-data"`, zero tokens/streak.
- Empty HOME `thermal statusline --verbose` → exit `0`, stdout empty, one stderr diagnostic.
- `thermal statusline --json` with data → `state:"ok"`.
- Empty HOME `thermal share` → exit `1`, stdout 0 bytes, stderr inventory.
- Empty HOME `thermal daily` → exit `1`, stdout 0 bytes, stderr inventory.

## Acceptance Criteria

- AC-001: `thermal live` in a non-TTY exits `2`, stdout is empty, and stderr contains `live needs a terminal` plus `--json`/`--json --stream` guidance. **PASS**
- AC-002: `thermal live --json` exits `0` with `type:"live"`; `live --json --stream` emits valid NDJSON snapshots. **PASS**
- AC-003: Empty-environment `thermal statusline` and `--plain` exit `0` with no stdout/stderr. **PASS**
- AC-004: Empty-environment `thermal statusline --json` exits `0` with `state:"no-data"`, zero `streak`/`todayTokens`, and a valid `day`/`updatedAt`. **PASS**
- AC-005: Statusline JSON with data includes `state:"ok"`; text output remains non-empty. **PASS**
- AC-006: Empty-environment `thermal share` exits `1`, stdout empty, stderr non-empty (inventory). **PASS**
- AC-007: Empty-environment `thermal daily` exits `1`, stdout empty, stderr inventory. **PASS**
- AC-008: `thermal statusline --verbose` with no data exits `0`, stdout empty, and stderr has exactly one diagnostic line. **PASS**
- AC-009: `thermal live --json --stream | head -1` exits cleanly with a valid snapshot. **PASS**
- AC-010: `go test -race ./...` passes, including `TestStatuslineNoDataIsSilent`, `TestStatuslineNoDataJSON`, `TestStatuslineNoDataVerbose`, `TestStatuslineNoDataPlain`, `TestRenderStatuslineNoData`, `TestRenderStatuslineOKState`, and the updated `TestSimulatedUser_LiveCommand`. **PASS**
- AC-011: `./scripts/simulated_user_gate.sh` passes 131/131 and includes the non-TTY/statusline section. **PASS**

## Regression Notes
- Foreground reports still use the inventory on empty data and exit `1`.
- Live TTY behavior and the TUI are unchanged.
- `statusline` no-data is not cached; the next prompt retries and becomes `ok` after data appears.
- The gallery gate now reuses the gate binary and honors the original Go caches, removing a hermetic-environment rebuild failure.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-nontty-and-statusline-output-contract",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Channel-clean non-TTY live and silent statusline degradation", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_14f1c2d9", "intent": "Channel-clean non-TTY live and silent statusline degradation"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
