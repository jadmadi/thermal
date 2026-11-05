# Facts: Read DSH v7 turn counts and session models

Evidence collected on 2026-11-05 from the live store at `~/.dsh` and from a
frozen copy at `/tmp/dshprobe` (produced with `cp -a ~/.dsh/storages`).

## Store Layout (verified)

- Session projections:
  `~/.dsh/storages/session_projcache/sessions/<session-id>.json`
  (48 files on 2026-11-05, 10 with tokens).
- Monolithic fallback: `~/.dsh/storages/session_projcache.json`
  (`tables.sessions`, 2 legacy entries, 84,207 tokens total).
- Event logs: `~/.dsh/sessions/<flattened-cwd>/<session-id>/` with
  `session.v4.jsonl.zstd` (current) or `session.jsonl.zstd` (August
  format). 23 session directories; 294 MB total; one directory
  (`session-a3956fa1-...`) has no projection entry and no usage events.
- Workspace index: `~/.dsh/storages/workspace.json` maps workspace paths to
  session ids. Not read by the loader today.

## Schema Drift (verified with `jq`)

- Projection `version`: 39 files at 7 (formatVersion 4), 8 at 7
  (formatVersion 3), 1 at 5 (legacy).
- Version 7 `record.rows` keys: `agentPreset`, `contextBreakdown`,
  `contextPressure`, `goal`, `inbox`, `llmRetry`, `permissions`, `plan`,
  `sandboxMode`, `sessionListMetadata`, `subagent`, `subagentCatalog`,
  `subagentModelSelectionPolicy`, `subagentTiming`, `timeContext`, `title`,
  `titleInput`, `todos`, `tokenUsage`, `turnBoundary`, `userQuestions`,
  `workingActivity`, `workingDirectory`.
- Version 7 has no `sessionStats` and no `modelSelection`. Version 5 had
  both.
- Version 5 `sessionStats.val.turns` equalled `lastTurn` in every old
  record checked (5/5, 2/2, 4/4, 2/2).

## Turn Source (verified)

- `turnBoundary.val.lastTurn` is present in 46 of 47 projected sessions.
- Active sessions on 2026-11-05: `lastTurn` 8, 4 and 1. Thermal reported
  one turn per session, three in total.
- Frozen snapshot: thermal reported 21 turns; `lastTurn`/`sessionStats`
  carry 30.

## Model Source (verified)

- `modelSelection.val.lastUsed.model` is set only in 11 of 47 sessions, all
  of which predate the version 7 default-preset flow; the token-bearing
  version 7 sessions have no `modelSelection` row at all.
- `~/.dsh/sessions/<cwd>/<id>/session.v4.jsonl.zstd` records the model in
  two events:
  - `{"type":"request/header","data":{"header":{"config":{"provider":
    "opencode-go","model":"deepseek-v4.1-flash"}}}}` at sequence 21.
  - `{"type":"request/context","data":{"provider":"opencode-go",
    "model":"deepseek-v4.1-flash","contextWindow":1000000}}` at sequence 22.
- Sampled logs resolve `deepseek-v4.1-flash` and `deepseek-flash`.
- `internal/loaders/zed.go:17` already imports
  `github.com/klauspost/compress/zstd`, so no new dependency is needed.
- `internal/loaders/jsonl.go:33` provides `newJSONLScanner` with a 32 MiB
  line ceiling and `formatScanWarning` for `Summary.Warnings`.

## Token Parity (verified)

- Frozen snapshot ground truth from the projections: 79,381,433 tokens.
- `thermal daily --tool=dsh --json` against the same frozen copy:
  79,381,433 tokens. Token math is correct and must not change.
- Unattributed tokens: 61,433,864 (77.4%).
- Model-less input/output/cacheRead: 1,704,402 / 226,320 / 59,503,142.
- At the catalog alias rates for `deepseek-v3` (input 0.287, output 1.147,
  cache read 0.0287, cache write 0.287 USD per million) the missing model
  accounts for about $2.46, so the DSH estimate moves from about $0.23 to
  about $2.69.
- `atria-dawn-preview` has no catalog price and stays unpriced. That is a
  catalog gap, not a loader gap.

## Re-Frozen Snapshot at Implementation (verified)

The store kept growing during the work, so the verification copy was
re-frozen with `cp -a ~/.dsh/storages` and `cp -a ~/.dsh/sessions`:

- 16 active projections, 101,774,684 tokens, 39 turns.
- 10 projections carry tokens and 6 are turn-only probes with
  `lastTurn` 1 or 2 and zero tokens.
- The previous loader produced 19 turns on this snapshot: 16 token-bearing
  sessions at one turn each plus 3 sessions with `sessionStats`.
- The fixed loader produces 39 turns, exactly the sum of
  `max(lastTurn, sessionStats)` across active projections.
- Unattributed tokens: 0.
- Model breakdown: `deepseek-v4.1-flash` 83,559,957,
  `atria-dawn-preview` 13,081,677, `deepseek-v4-flash` 4,865,892,
  `deepseek-flash` 182,951, `deepseek-v4-pro` 84,207.
- Estimated cost: $3.53, with only `atria-dawn-preview` unpriced.

## Current Loader State (verified)

- `internal/loaders/dsh.go` version 5 structs: `dshSessionStatsRow`,
  `dshModelSelectionRow`, `dshTokenUsageRow`.
- `extractDshSession` reads turns only from `sessionStats` and the model
  only from `modelSelection`.
- `LoadDshData` scans the projection directory with 16 workers, then
  merges the monolith while skipping seen ids.
- The live source signature (`GetToolSourceSig`) walks `storages` only;
  the static `DirScanCache` root is all of `~/.dsh`.

## Constraints

- Read-only ingestion: never write to `~/.dsh`.
- Every JSONL scan checks `scanner.Err()` and reports through
  `formatScanWarning`, stderr/verbose only.
- Canonical model ids pass through `modelName()` before map keys.
- Project attribution stays `thermal.ProjectKey(identity.cwd)`.
- Files stay under 1,000 lines; `dsh.go` is 518 lines today.

## Pre-Deploy Checklist

1. Table-driven tests for both sources and every negative case.
2. `go test -race ./internal/loaders/...`.
3. `gofmt -l` on touched files and `go vet ./...`.
4. Frozen snapshot comparison for turns, models, and cost.
5. Live `thermal daily --tool=dsh --verbose` run with no new warnings.
6. `./scripts/check.sh --pre-commit`.
7. README, AGENTS, and `project.jsonc` updated in the same commit.
