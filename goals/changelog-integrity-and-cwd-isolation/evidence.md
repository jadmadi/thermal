# Evidence: changelog-integrity-and-cwd-isolation

Date: 2026-10-29
Project: thermal-streak
Goal: changelog-integrity-and-cwd-isolation

## Verification Summary
- Removed the duplicate `## [0.21.0]` full-history dump and orphaned `## [Unreleased]` blocks from `CHANGELOG.md`, stripped all `(goals/<slug>/goal.md)` and `[goal: <slug>]` tokens across all historical entries, and synchronized `internal/changelog/CHANGELOG.md` byte-for-byte with `CHANGELOG.md`.
- Updated `LoadContent()` in `internal/changelog/changelog.go` to return `embeddedChangelog` directly, eliminating working-directory `CHANGELOG.md` hijacking.
- Hardened `Parse()` in `internal/changelog/changelog.go` to strip `(goals/<slug>/goal.md)` and `[goal: <slug>]` tokens via `goalTokenRegex` and skip duplicate version headers via `seenVersions`.
- Added unit tests `TestParseStripsGoalTokensAndDeduplicatesReleases`, `TestLoadContentIgnoresWorkingDirectoryChangelog`, and `TestEmbeddedChangelogCleanliness` in `internal/changelog/changelog_test.go`.
- Added automated checks in `scripts/simulated_user_gate.sh` (Section 13) verifying `cmp -s CHANGELOG.md internal/changelog/CHANGELOG.md`, zero `(goals/` or `[goal:` tokens, and zero duplicate version headers.

## Commands Executed
- `CGO_ENABLED=0 go test -v ./internal/changelog/...` -> PASS (6/6 tests)
- `go test -race -v ./internal/changelog/...` -> PASS (6/6 tests)
- `./scripts/simulated_user_gate.sh` -> PASS (77/77 checks)

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "changelog-integrity-and-cwd-isolation",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "changelog release-please goal token embedded", "matches": 5}],
  "memos": [{"id": "fact_2968052a", "text": "LoadContent() in internal/changelog/changelog.go must return embeddedChangelog directly without probing CWD paths, and CHANGELOG.md must remain byte-identical to internal/changelog/CHANGELOG.md with zero Sila goal tokens"}],
  "lessons": [{"id": "fact_2f166772", "symptom": "thermal changelog renders duplicate release cards and internal Sila goal tokens when run inside a repo directory"}],
  "cases": [{"id": "fact_043232ff", "intent": "Isolate embedded CLI changelog from working-directory files and sanitize internal goal attribution tokens"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_e9d82fb8ef26"}
}
```
