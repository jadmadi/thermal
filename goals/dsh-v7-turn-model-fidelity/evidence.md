# Evidence: dsh-v7-turn-model-fidelity

Date: 2026-11-05
Project: thermal-streak
Branch: feat/dsh-v7-turn-model
Goal: dsh-v7-turn-model-fidelity

Verdict: PASS

## Implementation Summary

- `internal/loaders/dsh.go` adds the version 7 `turnBoundary` row and takes
  the larger of `sessionStats.val.turns` and `turnBoundary.val.lastTurn`.
- A session whose `lastTurn` is greater than zero stays active even when it
  recorded no tokens, matching the DSH session list.
- `fillDshModels` indexes
  `~/.dsh/sessions/<flattened-cwd>/<id>/session*.jsonl.zstd` and
  head-decodes the first `request/context` (then `request/header`) event for
  active sessions with tokens and no projection model.
- Decoding uses `newJSONLScanner` and the existing
  `github.com/klauspost/compress/zstd` dependency. It stops at the first
  assistant message or 2,000 lines.
- A missing log or a log without a model is silent. A read or decode failure
  emits one `formatScanWarning`. A session with `modelSelection` already set
  never opens a log.
- Model lookups run through the existing bounded pattern (semaphore 16) and
  pass through `modelName()` before they become map keys.
- `internal/loaders/dsh_test.go` adds 11 tests covering turns, both model
  event shapes, and every rejection path.
- README, AGENTS section 2.D, and `project.jsonc` `lastEdit` are updated.

## Baseline vs Fixed (same frozen snapshot)

Frozen with `cp -a ~/.dsh/storages /tmp/dshprobe/` and
`cp -a ~/.dsh/sessions /tmp/dshprobe/`.

| Metric             | HEAD before | This branch | Delta    |
| ------------------ | ----------- | ----------- | -------- |
| Active sessions    | 16          | 16          | 0        |
| Total tokens       | 101,774,684 | 101,774,684 | 0        |
| Turns              | 21          | 39          | +18      |
| Unattributed tokens| 83,827,115  | 0           | -83.8M   |
| Estimated cost     | $0.2324     | $3.53       | +$3.30   |

The token total is identical, which proves the change only fills turns and
models. The model breakdown after the fix:

| Model                | Tokens     | Priced |
| -------------------- | ---------- | ------ |
| deepseek-v4.1-flash  | 83,559,957 | yes    |
| atria-dawn-preview   | 13,081,677 | no     |
| deepseek-v4-flash    | 4,865,892  | yes    |
| deepseek-flash       | 182,951    | yes    |
| deepseek-v4-pro      | 84,207     | yes    |

`atria-dawn-preview` has no models.dev entry. It is attributed and reported
under `No pricing for`, not counted as unattributed.

## Commands Executed

- Red phase, `go test ./internal/loaders/ -run 'TestLoadDshData_(TurnBoundary|Model|ZeroToken|TurnFallback)'`:
  five failures (`Turns = 1, want 8`; `Turns = 2, want 5`; two empty model
  maps; no decode warning).
- `go test -race ./internal/loaders/... -count=1` → PASS
  (`ok ... 2.4s`).
- `go vet ./...` → exit 0.
- `./scripts/check.sh --pre-commit` → PASS in 12 s (CLA, Diff Sentry,
  gofmt, package tests, binary compilation).
- Frozen snapshot, new binary:
  `DSH_HOME=/tmp/dshprobe /tmp/thermal-test daily --tool=dsh --json` →
  39 turns, 0 unattributed, five models.
- Frozen snapshot, HEAD binary built in a throwaway worktree → 21 turns,
  83,827,115 unattributed, $0.2324.
- Live store, `/tmp/thermal-test daily --tool=dsh --verbose` → exit 0,
  `~$3.66` total, stderr carries only the verbose banner.
- Timings on the static snapshot: cold 21 ms, warm 8 ms.

## Acceptance Criteria

| AC | Result | Receipt |
| --- | --- | --- |
| AC-001 | PASS | `go test -race ./internal/loaders/...` exits 0; 11 new tests |
| AC-002 | PASS | snapshot reports 39 turns (baseline 21) |
| AC-003 | PASS | unattributed 0; five models in the breakdown |
| AC-004 | PASS | $3.53 estimated (baseline $0.2324) |
| AC-005 | PASS | live verbose run exits 0 with no warnings |
| AC-006 | PASS | `gofmt -l` empty; `go vet ./...` exit 0 |
| AC-007 | PASS | `./scripts/check.sh --pre-commit` exit 0 |
| AC-008 | PASS | see `certification.json` and the closure receipt |

## Negative Cases Verified

| Case | Result |
| --- | --- |
| `lastTurn` 5, `sessionStats` 2 | 5 turns |
| `lastTurn` 2, `sessionStats` 5 | 5 turns |
| both absent, tokens > 0 | 1 turn |
| modelless, no log | empty model, no warning |
| modelless, log without request | empty model, no warning |
| modelless, request without model | empty model, no warning |
| modelless, corrupt zstd | tokens kept, one warning naming the log |
| `modelSelection` set, corrupt log | no read, no warning |
| zero-token session, corrupt log | no read, no warning |
| turn-only session, zero tokens | active with turns, zero tokens |

## Risks Left Open

- A model switch inside a live session is not re-read; the live source
  signature covers `storages` only. Documented in AGENTS.md.
- The static cache root remains `~/.dsh`, so the signature still stats about
  28,612 files, mostly `profiles/node_modules`. Measured at about 100 ms per
  invocation. A separate performance goal owns that change.
- `atria-dawn-preview` stays unpriced until models.dev lists it.

## Files Changed

- `internal/loaders/dsh.go`
- `internal/loaders/dsh_test.go`
- `README.md`
- `AGENTS.md`
- `project.jsonc`
- `goals/dsh-v7-turn-model-fidelity/goal.md`
- `goals/dsh-v7-turn-model-fidelity/facts.md`
- `goals/dsh-v7-turn-model-fidelity/plan.md`
- `goals/README.md` (generated index)

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "dsh-v7-turn-model-fidelity",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "prime_context"},
  "recall": [{"query": "DSH projcache loader turns model zstd session log", "matches": 107}],
  "memos": [],
  "lessons": [{"id": "fact_bf5872ec", "symptom": "DSH v7 showed zero models and one turn per session"}],
  "cases": [{"id": "fact_b695e8c6", "intent": "Attribute DSH v7 sessions whose projection cache dropped the model"}],
  "facts_verify": {"exit_code": 0, "note": "pre-existing historical hash mismatches reported, none from this goal"},
  "handoff": {"id": "hoff_2136e2807bf5"}
}
```
