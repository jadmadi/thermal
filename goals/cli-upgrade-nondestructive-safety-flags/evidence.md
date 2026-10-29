# Evidence: cli-upgrade-nondestructive-safety-flags

Date: 2026-10-29
Project: thermal-streak
Goal: cli-upgrade-nondestructive-safety-flags

## Implementation Summary
- Replaced the argument-free `runUpgrade()` with `upgradeOptions` + `runUpgradeWithOptions`, injected with an API URL, HTTP clients, an executable resolver, stdin, and a TTY flag.
- Added `parseUpgradeArgs` with `--check`, `--dry-run`, `--json`, `--yes`, `--force`, and `-h/--help`; unknown flags and stray positionals exit `2`.
- `--check` fetches release metadata and emits current/latest/asset/target; no download or mutation.
- `--dry-run` resolves the asset and target; no download or mutation.
- A mutation requires a TTY confirmation or `--yes`; non-interactive mutation without `--yes` exits `3` before any network call.
- `--json` writes exactly one buffered object to stdout; progress stays off stdout.
- `performUpgrade` probes the target directory, downloads/extracts/chmods first, then performs the atomic `.old`/`.new` swap and restores `.old` on rename failure.
- `--force` allows same-version reinstall or downgrade; without it, same/older releases produce `up-to-date`.
- Updated the upgrade help page, migration guide, JSON contract, README update section, project metadata, and simulated gate.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./...` → PASS.
- `go test -race ./cmd/thermal/...` → PASS.
- `./scripts/check.sh --pre-commit` → PASS (1s; gofmt, package tests, binary build).
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (98/98 checks).
- Disposable `upgrade --help` → exit `0`; md5 unchanged.
- `thermal upgrade --quiet` → exit `2`, `unknown flag for upgrade: --quiet`; no network.
- `thermal upgrade </dev/null` → exit `3`, `refusing to replace the binary without a terminal`; no network; md5 unchanged.
- `thermal upgrade --json </dev/null` → exit `3`, valid JSON `action:"declined"`, `mutated:false`; no network.
- `thermal upgrade --check --json` → exit `0`, `action:"check"`, `mutated:false`; binary md5 unchanged.
- `thermal upgrade --dry-run --json` → exit `0`, `action:"dry-run"`, `mutated:false`; binary md5 unchanged.
- PTY `thermal upgrade` with `n` → prompt shown, exit `3`, `cancelled`, md5 unchanged.

## Acceptance Criteria

- AC-001: `thermal upgrade --help` exits `0`, contains `--check`, `--dry-run`, `--json`, `--yes`, and `--force`, and leaves a disposable binary's md5 unchanged. **PASS**
- AC-002: `--check --json` via the httptest harness exits `0` with `action:"check"`, non-empty current/latest, and zero download requests. **PASS**
- AC-003: `--dry-run --json` exits `0` with `action:"dry-run"`, `mutated:false`, and no `.old`/`.new` files. **PASS**
- AC-004: The httptest harness with the current version returns `action:"up-to-date"` and no mutation. **PASS**
- AC-005: The httptest harness with a newer version returns `action:"upgraded"`, `mutated:true`, and replaces the target file (content verified). **PASS**
- AC-006: `thermal upgrade --quiet` exits `2`, names `--quiet`, and makes no HTTP request. **PASS**
- AC-007: `thermal upgrade </dev/null` exits `3`, prints the refusal, and leaves the binary unchanged. **PASS**
- AC-008: In a PTY, `thermal upgrade` prompts; `n` exits `3`, prints `upgrade cancelled`, and leaves the binary unchanged. **PASS**
- AC-009: The httptest 404 path exits `1`, prints an actionable error, and leaves the target byte-identical. **PASS**
- AC-010: A truncated `.tar.gz` exits `1`, leaves the target byte-identical, and removes `.new`/`.old`. **PASS**
- AC-011: `go test -race ./...` passes, including `TestUpgradeCheckDoesNotMutate`, `TestUpgradeDryRunDoesNotMutate`, `TestUpgradeJSONContract`, `TestUpgradeNonTTYRefusal`, `TestUpgradeJSONDeclined`, `TestUpgradeDeclinePrompt`, `TestUpgradeYesReplacesBinary`, `TestUpgradeSameVersionNoMutation`, `TestUpgradeForceSameVersion`, `TestUpgradeReleaseAPIFailure`, `TestUpgradeTruncatedArchiveLeavesTarget`, and `TestParseUpgradeArgs`. **PASS**
- AC-012: `./scripts/simulated_user_gate.sh` passes 98/98 and includes the upgrade-safety section. **PASS**

## Regression Notes
- `thermal upgrade` still performs a real upgrade on a TTY after confirmation, and with `--yes` in automation.
- `--help` is intercepted before upgrade parsing and never mutates.
- The background update check and cache behavior are unchanged.
- The old global `httpClient`/`downloadClient` are retained as production defaults; tests inject their own.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-upgrade-nondestructive-safety-flags",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Make self-upgrade previewable and non-destructive", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_1fd437c3", "intent": "Make self-upgrade previewable and non-destructive"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
