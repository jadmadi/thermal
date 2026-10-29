# Goal: Ingest Muse JSONL session tokens, model facets, and project attribution

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Report: `receipt_muse-token-and-project-ingestion_1793262214`


## Goal Description
Upgrade the Meta Muse loader (`internal/loaders/muse.go`), live source-signature scanner (`internal/loaders/sourcesig.go`), directory cache registry (`internal/loaders/registry.go`, `internal/loaders/dircache.go`), and leaderboard/dashboard renderers (`internal/render/leaderboard.go`, `internal/render/dashboard.go`) to ingest real-time token telemetry (`input_tokens`, `output_tokens`, `reasoning_tokens`, `cache_read_tokens`, `cache_write_tokens`), model facets, and project attribution from `~/.local/share/muse/sessions/**/*.jsonl` while keeping `DailyRow.Cost` strictly zero so catalog rate-card math is never mislabeled as recorded billed spend.

## Dependencies & Execution Order
- **Mode**: Independent (can run in parallel)
- **Depends On**: none
- **Sequence**: 220
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required

## Files to Touch
- Edit: `internal/loaders/muse.go`
- Edit: `internal/loaders/muse_test.go`
- Edit: `internal/loaders/sourcesig.go`
- Edit: `internal/loaders/registry.go`
- Edit: `internal/loaders/dircache.go`
- Edit: `internal/render/dashboard.go`
- Edit: `internal/render/leaderboard.go`

## Requirements
1. **Muse Session JSONL Scanner (`internal/loaders/muse.go`)**:
   - Walk `<dataDir>/sessions/**/*.jsonl` using a bounded worker pool (capped at 16 workers) and `newJSONLScanner` (32 MiB ceiling), checking `scanner.Err()` and appending diagnostics via `formatScanWarning`.
   - Extract disjoint token counts from `model_completed` events: `netInput = input_tokens - (cache_read_tokens + cache_write_tokens)`, `netOutput = output_tokens - reasoning_tokens`, `reasoning = reasoning_tokens`, `cacheRead = cache_read_tokens`, `cacheWrite = cache_write_tokens`.
   - Resolve active model IDs from `event.model`, `record.model` (`run.model.configured`), and `status` event `details.facets[].detail.model`, normalized through `modelName()`.
   - Detect deterministic synthetic timestamps (`recorded_at` in `1780531500000000..1780531500999999`) emitted by `approval.review.runtime` sandboxes so they do not corrupt `LongestSessionMs` or bucket turns into `2026-06-04`; attribute their calendar day from the parent session timestamp or `YYYY/MM/DD` directory path.
   - Attribute project paths (`workspace_root` and `cwd`) through `thermal.ProjectKey` onto `thermal.ProjectDay` rows, and fold in unindexed legacy rows from `session-index.db`.
   - Leave `DailyRow.Cost`, `ProjectDay.Cost`, and `Summary.Cost` at `0` in `LoadMuseData` because Muse session logs do not record billed USD receipts; Muse pricing belongs in `internal/pricing` (`EstimatedCost`).
2. **Live Source Signature & Delta Cache (`internal/loaders/sourcesig.go`, `internal/loaders/registry.go`, `internal/loaders/dircache.go`)**:
   - Add `thermal.ToolMuse` to `GetToolSourceSig` so `thermal live` polls `sessions/**/*.jsonl` alongside `session-index.db` and `-wal`.
   - Set `DataSubdir: "sessions"` on `ToolMuse` in `AllTools()` and enable `loadDirScanCache` / `saveDirScanCache` with bumped `dirScanCacheVersion`.
3. **Token Warriors Promotion (`internal/render/leaderboard.go`, `internal/render/dashboard.go`)**:
   - Promote Muse from `Activity Hunters` to `Token Warriors` when token telemetry is present, falling back to `prompts` only for legacy prompt-only `session-index.db` stores.

## Acceptance Criteria
- AC-001: `CGO_ENABLED=0 go test -race ./internal/loaders -run 'TestLoadMuseData'` exits 0.
- AC-002: `LoadMuseData` produces disjoint token totals, attributes `approval-review` turns to the parent session's day without inflating `LongestSessionMs`, and leaves `DailyRow.Cost == 0`.
- AC-003: `GetToolSourceSig` detects appended bytes in `sessions/**/*.jsonl` for `thermal.ToolMuse`.

## References
- **Shared Understanding & Fact Sheet**: [`goals/muse-token-and-project-ingestion/facts.md`](facts.md)
- **Execution Plan**: [`goals/muse-token-and-project-ingestion/plan.md`](plan.md)

## Memory Plan
- **Prime**: run `sila prime` (CLI) or `prime_context` (MCP) at task start and record the briefing timestamp.
- **Recall**: run `sila recall "<symptom or design question>"` before implementation; record each query and match count.
- **Writes**: record at least one `memo`, `lesson`, or `case` grounded to a file in this goal; use a `case` when the goal introduces a reusable mechanism.
- **Facts**: run `sila facts verify` before certification and record the exit code and violation delta.
- **Handoff**: run `sila handoff "<summary>" --open="..."` at session end and record the handoff ID.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`, per `memory-protocol-goal-contract`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`); unchecked tasks block the `implemented` transition.
2. Implementation is committed with `(goals/muse-token-and-project-ingestion/goal.md)` or `[goal: muse-token-and-project-ingestion]` attribution under a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` prefix.
3. Goal is certified (`goals/muse-token-and-project-ingestion/certification.json` with `verified: true` or a signed `audit-receipt.json`) and any markdown evidence contains an explicit standalone `Verdict: PASS` declaration.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0` for delivered goals, with a clean `git status`.
