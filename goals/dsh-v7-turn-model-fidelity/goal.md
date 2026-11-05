# Goal: Read DSH v7 turn counts and session models

## Goal Description
Priority: P1 (data fidelity). DeepSeek Harness (dsh) version 7 stopped
writing `sessionStats` and `modelSelection` into
`~/.dsh/storages/session_projcache/sessions/*.json`. Thermal still parses the
version 5 shape, so on 2026-11-05 it read token totals exactly but reported
the wrong turn count and lost model attribution:

- 79,381,433 tokens were read correctly from a frozen snapshot.
- Thermal reported 21 turns; the same sessions carry 30 in
  `turnBoundary.val.lastTurn` and `sessionStats.val.turns`.
- 61,433,864 of those tokens (77.4%) had no model, so no price applied.
- The DSH row showed about $0.23 instead of about $2.69 at the catalog
  alias rates used for `deepseek-v4.1-flash` and `deepseek-flash`.

The outcome of this goal is fidelity for current DSH stores, with no CLI or
JSON contract change:

1. Turn counts prefer `turnBoundary.val.lastTurn` when `sessionStats` is
   absent, and take the larger value when both exist.
2. A session model falls back to the first `request/context` or
   `request/header` event in
   `~/.dsh/sessions/<flattened-cwd>/<session-id>/session*.jsonl.zstd`
   when `modelSelection.val.lastUsed.model` is empty.
3. The existing bounded scan, warning channel, canonical model names,
   project attribution, and token math stay unchanged.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required

Related goals:
- `dsh-loader` owns the original version 5 loader and its tests; this goal
  owns the version 7 drift in that same file.
- Any future cache-root narrowing for `~/.dsh` is a separate performance
  goal and is a non-goal here.

## Files to Touch
- `internal/loaders/dsh.go`
- `internal/loaders/dsh_test.go`
- `README.md`
- `AGENTS.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/dsh-v7-turn-model-fidelity/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/dsh-v7-turn-model-fidelity/plan.md`](plan.md)
- Original loader goal:
  [`goals/dsh-loader/goal.md`](../dsh-loader/goal.md)
- Real store inspected on 2026-11-05 at `~/.dsh`.

## Requirements

1. **Turn counts.**
   - Add `turnBoundary` to the `dshRows` shape:
     `{"val":{"lastTurn":N}}`.
   - In `extractDshSession`, read `sessionStats.val.turns` (existing) and
     `turnBoundary.val.lastTurn` (new), then use the larger value.
   - Keep the existing fallback: when the result is `0` and tokens are
     greater than `0`, count one turn.
   - A session whose `lastTurn` is greater than `0` stays active even when
     it recorded no tokens. It contributes turns and zero tokens, matching
     the DSH session list. Probe sessions such as "Reply with exactly: ok"
     therefore appear as one turn with no spend.
   - Do not change `durationMs` (it still comes from `sessionStats` only).

2. **Model fallback.**
   - Run the fallback only for sessions that are active (`total > 0`) and
     have an empty model after `extractDshSession`.
   - Build an index of session logs under
     `filepath.Join(dataDir, "sessions")`: two directory levels, mapping
     the session directory name to its log path. Match
     `session.v4.jsonl.zstd` first, then any `session*.jsonl.zstd`.
   - Decode with `github.com/klauspost/compress/zstd`, which is already a
     direct dependency (`internal/loaders/zed.go:17`).
   - Scan decompressed lines with `newJSONLScanner` (32 MiB ceiling).
   - Read only lines containing `"request/context"` or `"request/header"`;
     take `data.model` first, then `data.header.config.model`.
   - Stop at the first non-empty model. Stop the scan at the first
     `assistant/message` event or after 2,000 lines, whichever comes first,
     so a log without request metadata cannot stream the whole file.
   - Pass the result through `modelName()` before it becomes a map key.
   - A missing log file or a log without a model is not an error and emits
     no warning. A read or zstd decode failure emits one warning through
     `formatScanWarning` in `Summary.Warnings`.
   - Process model lookups with the existing bounded worker pattern
     (semaphore capped at 16); a session with `modelSelection` already set
     never opens a log.

3. **Cache and live behavior.**
   - The static `DirScanCache` root stays `~/.dsh`, so session log writes
     already invalidate a cached snapshot. Do not narrow it.
   - `GetToolSourceSig` keeps scanning `storages` only. A model switch in a
     live session is not re-read; record this in `AGENTS.md` as a known
     limitation of the live view.

4. **Documentation.**
   - `README.md:793` states the DSH source shape. Update it to name the
     version 7 fallback and the turn source.
   - `AGENTS.md` section 2.D updates the DeepSeek (DSH) bullet with the
     version 7 fields and the zstd fallback.
   - `project.jsonc` `lastEdit` is updated to the commit timestamp; no
     capability or version claim changes.

## Non-Goals
- Per-turn model attribution. Thermal attributes per session, so a session
  that switches models keeps the first model found.
- Changing the pricing catalog or adding an `atria-dawn-preview` entry.
- Narrowing the `~/.dsh` cache root to `~/.dsh/storages` (separate
  performance goal).
- Reading message counts or activity-only DSH data.
- Changing the `--json` schema, CLI flags, or report rendering.
- A new loader file; the change stays inside `internal/loaders/dsh.go` so
  the loader keeps one home.

## Acceptance Criteria
- AC-001: `go test -race ./internal/loaders/...` exits `0` and includes
  `TestLoadDshData_TurnBoundaryFallback`,
  `TestLoadDshData_TurnBoundaryWinsOverSessionStats`,
  `TestLoadDshData_ModelFromSessionLog`,
  `TestLoadDshData_ModelSelectionSkipsLog`, and
  `TestLoadDshData_ModelLogCorruptWarns`.
- AC-002: On the frozen snapshot used below, `thermal daily --tool=dsh
  --json` reports 39 turns, matching the projection ground truth (the
  previous loader reported 19 on the same snapshot).
- AC-003: The same snapshot reports `deepseek-v4.1-flash`,
  `deepseek-flash`, `deepseek-v4-flash`, `deepseek-v4-pro`, and
  `atria-dawn-preview` in the model breakdown; unattributed tokens drop
  from 61,433,864 to 0. `atria-dawn-preview` stays unpriced, which is a
  catalog gap, not an attribution gap.
- AC-004: The estimated DSH cost rises from about $0.23 to about $3.53
  under the existing alias rates.
- AC-005: `thermal daily --tool=dsh --verbose` against the live store exits
  `0` and prints no new warnings.
- AC-006: `go vet ./...` and `gofmt -l` on the touched Go files are clean.
- AC-007: `./scripts/check.sh --pre-commit` exits `0`.
- AC-008: `sila goals --json` reports `status: "implemented"` and
  `verification_status: "verified"` for this slug with a clean tree.

## Negative Cases & Failure Modes
| Input | Required result |
| --- | --- |
| `sessionStats.turns` 2 and `lastTurn` 5 | 5 turns |
| `sessionStats` absent, `lastTurn` 8 | 8 turns |
| both absent, tokens > 0 | 1 turn |
| empty model, no session log | model `""`, no warning |
| empty model, log without request events | model `""`, no warning |
| empty model, corrupt zstd | tokens kept, one warning |
| model set, corrupt log present | no log read, no warning |
| zero-token session, corrupt log | no log read, no warning |
| `request/context` absent, `request/header` present | header model used |
| two session dirs for one id | first index hit wins |

## Metrics & Regression Guards
- Turn parity: the frozen snapshot must move from 21 to 30 turns.
- Attribution: at least 95% of snapshot tokens carry a model after the fix.
- Noise: zero new warnings on the live store.
- Isolation: a loader test proves a set `modelSelection` never opens the
  log.
- Cache: the static snapshot remains valid across an unchanged tree.

## Risks & Mitigations
- **Risk:** zstd head decoding costs more than the token scan.
  **Mitigation:** decode only model-less active sessions, stop at the first
  request event, cap at 2,000 lines.
- **Risk:** a log is mid-write and decodes to a partial frame.
  **Mitigation:** treat decode failure as a warning and keep the token
  totals, which come from the projcache.
- **Risk:** a model-switching session is misattributed.
  **Mitigation:** document the limitation in `AGENTS.md`; do not guess.
- **Risk:** a stale `lastTurn` inflates turns for an open turn.
  **Mitigation:** the projcache value is the same one DSH shows; take it as
  recorded.

## Memory Plan
- **Prime**: `prime_context` at task start (retrieval recorded in the
  session).
- **Recall**: `sila recall "DSH projcache loader turns model zstd session
  log"` returned 107 resolutions; none covered version 7.
- **Writes**: one `case` for the "projection cache plus transcript
  fallback" mechanism and one `lesson` for silent schema drift.
- **Facts**: `sila facts verify` before certification.
- **Handoff**: `sila handoff "dsh v7 turn and model fidelity shipped"
  --open="cache-root narrowing goal"`.
- **Evidence**: `## Memory Writes` block with the `sila/memory-writes/v1`
  fenced JSON.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/dsh-v7-turn-model-fidelity/goal.md)` under a `fix(loaders):`
   subject.
3. `evidence.md` contains a standalone `Verdict: PASS` with the snapshot
   numbers, test command, and live run.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0` for this
   slug, with a clean `git status`.
