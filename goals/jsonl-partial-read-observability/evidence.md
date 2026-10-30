# Evidence: Make truncated and malformed JSONL records observable

Date: 2026-10-30
Binary: `/tmp/thermal-g9` (source build)
Goal: `jsonl-partial-read-observability`

## Acceptance Matrix

| AC | Fixture / command | Result |
| --- | --- | --- |
| AC-001 | valid + truncated Claude line, `daily --json --verbose` | `skippedRecords:1`, `truncatedTails:1`, one stderr warning |
| AC-002 | same without `--verbose` | counters present; stderr empty |
| AC-003 | malformed middle line | `malformedLines:1`; both valid records counted |
| AC-004 | blank and non-assistant records | read, not skipped |
| AC-005 | over-ceiling line | `ceilingHits` counted; existing scanner warning |
| AC-006 | duplicate message id | `duplicates`, not `skippedRecords` |
| AC-007 | clean scan | new keys omitted; payload unchanged |
| AC-008 | directory cache round-trip | counters and warning survive the second scan |
| AC-009 | old cache schema | `dirScanCacheVersion`/`codexRolloutCacheVersion` bumped; old entries miss |
| AC-010 | `--json --verbose` | stdout parses; warnings only on stderr |
| AC-011 | `go test ./internal/loaders/... ./cmd/thermal/...` | ok |
| AC-012 | `simulated_user_gate.sh` | 171 passed, 0 failed |

## Observed Output

```text
$ thermal daily --json --verbose
exit 0
{
  "type": "daily",
  "skippedRecords": 1,
  "truncatedTails": 1
}
stderr:
thermal: warning: Claude: skipped 1 truncated tail record(s) in <path>
```

```text
$ thermal daily --json --verbose   # malformed middle line
exit 0
{"skippedRecords":1,"malformedLines":1,"activeDays":1}
thermal: warning: Claude: skipped 1 malformed record(s) in <path>
```

```text
$ thermal daily --json   # clean file
exit 0
{"type":"daily", ...}   # no skippedRecords key
```

## Implementation Notes

- Added `thermal.ScanStats` and the `Summary` fields `SkippedRecords`,
  `MalformedLines`, `TruncatedTails`, `CeilingHits`, and `ScanStats`.
- Claude and generic history scans distinguish a malformed middle line from an
  unterminated final line using `fileEndsWithNewline`.
- command-code, Droid, Grok, Agy, and Codex rollout scans count malformed
  records; Codex rollout cache persists the counters.
- `formatRecordWarning` emits one per-file warning under `--verbose`; a
  scanner ceiling keeps its existing warning and is not duplicated.
- Report envelopes expose `skippedRecords`, plus malformed/truncated/ceiling
  counts on period reports; per-tool `scanStats` rides the summary.
- `dirScanCacheVersion` bumped to 4 and `codexRolloutCacheVersion` to 2 so old
  cache entries miss instead of decoding into zero counters.
- Clean payloads omit all new keys; warnings never enter stdout.

## Verification Commands

```bash
go test ./internal/loaders/... ./cmd/thermal/...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All passed. The gate reported `171 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "jsonl-partial-read-observability",
  "prime": {"ran": true},
  "recall": [
    {"query": "truncated jsonl silently skipped", "matches": 5},
    {"query": "scanner err warning", "matches": 5}
  ],
  "memos": [{"id": "fact_f1ac9b92"}],
  "lessons": [{"id": "fact_beeee424"}],
  "cases": [{"id": "fact_a14171cf"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6258ac4945e3"}
}
```

Verdict: PASS
