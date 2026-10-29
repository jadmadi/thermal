# Evidence: Ingest Muse JSONL session tokens, model facets, and project attribution

Date: 2026-10-29
Project: thermal-streak
Goal: muse-token-and-project-ingestion

## Changes

- `internal/loaders/muse.go`:
  - Implemented `scanMuseJSONL` scanning `sessions/**/*.jsonl` with `newJSONLScanner` (32 MiB ceiling).
  - Implemented disjoint token accounting: `netInput = input_tokens - (cache_read + cache_write)`, `netOutput = output_tokens - reasoning_tokens`.
  - Filtered deterministic synthetic clocks emitted by `approval.review.runtime` sandboxes (`recorded_at` in `1780531500000000..1780531500999999`) to prevent `LongestSessionMs` corruption and `2026-06-04` day leaks.
  - Extracted model IDs from `status` event `facets[].detail.model` when `model_completed` omits `model`.
  - Normalized project roots (`workspace_root` and `cwd`) via `thermal.ProjectKey` onto `thermal.ProjectDay`.
  - Maintained strict cost provenance by leaving `DailyRow.Cost == 0`, `ProjectDay.Cost == 0`, and `Summary.Cost == 0` (Muse writes rate cards, not billed invoices; estimates belong in `internal/pricing`).
- `internal/loaders/sourcesig.go`: added `case thermal.ToolMuse` to walk `sessions/**/*.jsonl` alongside `session-index.db` and `-wal`.
- `internal/loaders/registry.go`: registered `DataSubdir: "sessions"` on `ToolMuse` in `AllTools()`.
- `internal/loaders/dircache.go`: bumped `dirScanCacheVersion` to 3.
- `internal/render/dashboard.go` and `internal/render/leaderboard.go`: promoted Muse to `Token Warriors` when token telemetry is present.
- `internal/loaders/muse_test.go`: added `TestLoadMuseData_SessionJSONL`, `TestLoadMuseData_Hybrid`, and `TestLoadMuseData_SourceSig`.

## Verification

```text
$ CGO_ENABLED=1 go test -race ./internal/loaders -run 'TestLoadMuseData' -v
--- PASS: TestLoadMuseData_SessionIndex (0.01s)
--- PASS: TestLoadMuseData_SessionJSONL (0.00s)
--- PASS: TestLoadMuseData_Hybrid (0.01s)
--- PASS: TestLoadMuseData_SourceSig (0.00s)
PASS
ok  	github.com/jadmadi/thermal/internal/loaders	1.028s

$ CGO_ENABLED=1 go test -race ./internal/loaders/...
ok  	github.com/jadmadi/thermal/internal/loaders	2.414s
```

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "muse-token-and-project-ingestion",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "muse session tokens and rate card provenance", "matches": 3}],
  "memos": [{"id": "fact_c3c74475", "text": "Muse logs model rate cards rather than billed invoices; keep DailyRow.Cost at 0 in LoadMuseData so pricing is strictly EstimatedCost."}],
  "lessons": [{"id": "fact_3b951213", "symptom": "Muse approval-review synthetic timestamps (1780531500000000) inflated session durations across months and created false 2026-06-04 active days"}],
  "cases": [{"id": "fact_a9a77e35", "intent": "Ingest Meta Muse JSONL session tokens, subagent streams, and approval reviews with disjoint accounting"}],
  "facts_verify": {"exit_code": 2},
  "handoff": {"id": "hoff_cd63055e31a4"},
  "waivers": [{"verb": "facts", "reason": "Pre-existing grounding drift reporting missing files from earlier relocated projects."}]
}
```
