# Goal: Make truncated and malformed JSONL records observable

## Goal Description
Priority: P1 (data trust). A half-written final line is the most common
real-world corruption: an agent is killed mid-write. Today every JSONL loader
silently `continue`s on a JSON parse error, so the record disappears and the
totals shrink. `--verbose` reports an unreadable file but not a malformed one,
and the JSON summary has no counter a pipeline can alert on. The number looks
authoritative while the data is partial.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`) on a
fixture whose only Claude file had one valid assistant record and one
truncated final line:

- `thermal daily --json --verbose` -> exit `0`, stderr 0 B, stdout shows one
  active day and the valid record's tokens; the truncated record is invisible.
- A control file with only the valid line produces the same summary, so a
  consumer cannot distinguish "one valid record" from "one valid plus one
  dropped".
- An unreadable file does produce a warning, proving the channel exists:
  `thermal: warning: Claude: <path>: open ...: permission denied`.
- A line over the 32 MiB ceiling produces
  `<path>: line exceeds N byte ceiling, rest of file skipped`.

The outcome of this goal is a **partial-read observability contract**:

1. Every JSONL scan counts records read, skipped, malformed, truncated, and
   over-ceiling.
2. Malformed and truncated records produce one stderr warning under
   `--verbose`; stdout and `--json` stay clean.
3. JSON exposes `skippedRecords`, `malformedLines`, and `truncatedTails` per
   tool and in the aggregate.
4. Dircache and codex-cache round-trip the counters; a cache hit does not hide
   a previous partial read.
5. Counters are disjoint: a blank line, an intentional non-assistant record,
   or a duplicate is not "skipped" unless the scan actually dropped data.

## Dependencies & Execution Order
- **Mode**: Dependent (prerequisite goals are implemented)
- **Depends On**: `jsonl-scan-diagnostics`, `jsonl-line-ceilings`
- **Sequence**: 263
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3

Related goals:
- `jsonl-scan-diagnostics` owns `scanner.Err()` and the warning channel; this
  goal owns per-record parse failures and counters.
- `jsonl-line-ceilings` owns the 32 MiB scanner ceiling; this goal counts a
  ceiling hit separately from a malformed line.
- `directory-scan-incremental-cache` and `codex-incremental-rollout-cache` own
  the caches; this goal requires them to persist the new counters.

## Files to Touch
- `internal/thermal/types.go`
- `internal/loaders/jsonl.go`
- `internal/loaders/claude.go`
- `internal/loaders/commandcode.go`
- `internal/loaders/droid.go`
- `internal/loaders/grok.go`
- `internal/loaders/agy.go`
- `internal/loaders/dsh.go`
- `internal/loaders/jsonl_test.go`
- `internal/loaders/claude_droid_test.go`
- `internal/loaders/commandcode_test.go`
- `internal/loaders/dircache.go`
- `internal/loaders/dircache_test.go`
- `internal/loaders/codexcache.go`
- `cmd/thermal/json.go`
- `docs/JSON-CONTRACT.md`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/jsonl-partial-read-observability/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/jsonl-partial-read-observability/plan.md`](plan.md)
- Prior contracts:
  [`goals/jsonl-scan-diagnostics/goal.md`](../jsonl-scan-diagnostics/goal.md),
  [`goals/jsonl-line-ceilings/goal.md`](../jsonl-line-ceilings/goal.md)

## Requirements

1. **Scan counters.**
   ```go
   type ScanStats struct {
       RecordsRead     int64 `json:"recordsRead,omitempty"`
       RecordsSkipped  int64 `json:"recordsSkipped,omitempty"`
       MalformedLines  int64 `json:"malformedLines,omitempty"`
       TruncatedTails  int64 `json:"truncatedTails,omitempty"`
       CeilingHits     int64 `json:"ceilingHits,omitempty"`
       Duplicates      int64 `json:"duplicates,omitempty"`
   }
   ```
   - `RecordsRead` counts non-empty physical lines the scanner yielded.
   - `MalformedLines` counts `json.Unmarshal` failures that are not the final
     unterminated line.
   - `TruncatedTails` counts a final line that is unterminated and fails to
     parse. Detect with `bufio.Scanner` plus a check that the file's last byte
     is not `\n`; or read the last byte separately.
   - `CeilingHits` counts `bufio.ErrTooLong` from `scanner.Err()`.
   - `Duplicates` counts the existing dedupe removals.
   - Blank lines, comments (if any), and records whose `type` is intentionally
     ignored are **not** skipped; they are `RecordsRead` but not errors.
   - A skipped record is one that contained usage the loader could not
     interpret. Do not count a valid non-assistant record as skipped.

2. **Summary surface.**
   - Add to `thermal.Summary`:
     ```go
     SkippedRecords int64 `json:"skippedRecords,omitempty"`
     MalformedLines int64 `json:"malformedLines,omitempty"`
     TruncatedTails int64 `json:"truncatedTails,omitempty"`
     CeilingHits    int64 `json:"ceilingHits,omitempty"`
     ScanStats      *ScanStats `json:"scanStats,omitempty"`
     ```
   - Keep `Warnings []string json:"-"`; warnings never enter JSON.
   - `SkippedRecords = MalformedLines + TruncatedTails + CeilingHits`.
   - Aggregate across tools in the report envelope: a top-level
     `"skippedRecords"` sum and a per-tool `"scanStats"` object.
   - Omit zero values; a clean scan adds no keys and no bytes.
   - `--json` with a partial read still exits `0`; the data is valid but
     partial. A `--strict` flag is not part of this goal.

3. **Warnings.**
   - Under `--verbose`, emit one line per file, not per record:
     ```
     thermal: warning: Claude: skipped 1 malformed record in <path> (line 2)
     thermal: warning: Claude: truncated final record in <path> (line 2)
     thermal: warning: Claude: line exceeds 33554432 byte ceiling; rest of file skipped
     thermal: warning: Claude: 2 records skipped in <path>
     ```
   - The path may be `~`-shortened; the line number is 1-based.
   - Sort warnings deterministically; do not duplicate a file-level summary and
     a per-record line for the same count (choose the summary form when
     `>1`, the singular form when `==1`).
   - Without `--verbose`, no warning; the JSON counters remain.
   - Never write a warning to stdout, even when `--json` is set.

4. **Loader coverage.**
   - Apply to every JSONL scanner:
     `claude.go`, `commandcode.go`, `droid.go`, `grok.go`, `agy.go`,
     `dsh.go` (JSON files use `json.Unmarshal` on the whole file; count a
     malformed file as one skipped record), and `jsonl.go`.
   - A loader that uses `json.Decoder` on a stream counts a `Decode` error as a
     malformed record and stops at that record.
   - A loader that intentionally ignores a record type does not increment
     `RecordsSkipped`.
   - Per-file results carry `ScanStats`; the worker pool aggregates under the
     existing thread-safe path.

5. **Cache round-trip.**
   - `dircache.go` persists `ScanStats` alongside `Warnings`; a cache hit
     restores both.
   - `codexcache.go` persists the counters for rollout files.
   - A cache hit must not lose a previous partial-read warning; the
     `--verbose` warning is re-emitted from the cached stats.
   - Cache schema version bumps; an older cache without the counters is
     treated as a miss, not as a clean scan.
   - A test writes a partial file, scans, writes the cache, scans again, and
     asserts the counters and warning survive.

6. **Tests.**
   - Truncated final line: one `TruncatedTails`, one warning under `--verbose`,
     one valid record counted.
   - Malformed middle line: one `MalformedLines`, valid records before and
     after counted.
   - Blank lines and non-assistant records: `RecordsRead` increments, no
     skipped counter.
   - Line ceiling: one `CeilingHits`, existing warning preserved.
   - Unreadable file: existing warning, no `MalformedLines`.
   - Duplicate IDs: `Duplicates` increments, not `SkippedRecords`.
   - Cache round-trip for `dircache` and `codexcache`.
   - JSON schema: partial read includes `skippedRecords`, per-tool `scanStats`,
     and no `warnings` key; a clean read omits them.
   - stdout purity: `--json --verbose` stdout parses, warnings on stderr.
   - Gate: `simulated_user_gate.sh` includes a truncated fixture.

7. **Documentation.**
   - `docs/JSON-CONTRACT.md` documents the new additive keys and the
     "partial read is valid data" rule.
   - README mentions that `--verbose` reports skipped records.
   - MIGRATION entry for the additive JSON keys; no breaking change.
   - `project.jsonc` updated.

## Non-Goals
- Repairing or rewriting malformed lines.
- Reading a partial line as valid JSON.
- Adding a `--strict` exit code for partial reads.
- Changing the 32 MiB ceiling.
- Counting prompt text, tool calls, or non-usage records as skipped.
- Changing loader token math.

## Acceptance Criteria
- AC-001: A fixture with one valid and one truncated final line, scanned with
  `--json --verbose`, exits `0`; stdout has one active day, `truncatedTails:1`,
  and `skippedRecords:1`; stderr has one truncated-record warning.
- AC-002: The same fixture without `--verbose` has the same JSON counters and
  empty stderr.
- AC-003: A malformed middle line yields `malformedLines:1` and both valid
  records are counted.
- AC-004: Blank lines and non-assistant records do not increment
  `skippedRecords`.
- AC-005: A line over the ceiling yields `ceilingHits:1` and the existing
  warning.
- AC-006: A duplicate message ID increments `duplicates`, not
  `skippedRecords`.
- AC-007: A clean scan omits all new keys and is byte-identical to today.
- AC-008: `dircache` and `codexcache` round-trip the counters and re-emit the
  warning under `--verbose`.
- AC-009: An old cache without counters is a miss, not a clean scan.
- AC-010: `--json --verbose` stdout parses with `jq -e .`; warnings are stderr
  only.
- AC-011: `go test -race ./internal/loaders/... ./cmd/thermal/...` exits `0`,
  including `TestTruncatedTail`, `TestMalformedMiddle`,
  `TestScanStatsCacheRoundTrip`, and `TestPartialReadJSON`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the
  truncated fixture.

## Negative Cases & Failure Modes
| Fixture | Required result |
| --- | --- |
| Valid file | no new keys, no warning |
| Truncated final line | `truncatedTails:1`, warning |
| Malformed middle line | `malformedLines:1`, warning |
| Blank lines | read, not skipped |
| Non-assistant record | read, not skipped |
| Duplicate ID | `duplicates:1`, not skipped |
| Over-ceiling line | `ceilingHits:1`, warning |
| Unreadable file | existing warning |
| Cache hit with counters | counters restored |
| Old cache | miss, rescan |
| `--json` + partial | valid JSON, exit 0 |
| `--verbose` | stderr only |

## Metrics & Regression Guards
- Counter disjointness: `SkippedRecords == MalformedLines + TruncatedTails +
  CeilingHits`; a test enforces it.
- No double counting: a record is counted once across the categories.
- Clean-scan bytes: a clean JSON payload is unchanged; a golden test enforces
  it.
- Cache fidelity: counters and warnings survive two scans.
- Channel purity: zero warnings on stdout; every JSON line parses.
- Gate: simulated user gate asserts AC-001, AC-003, and AC-010.

## Risks & Mitigations
- **Risk:** Counting every ignored line inflates the number.
  **Mitigation:** only parse failures and ceiling/truncation count; document
  the taxonomy.
- **Risk:** Detecting a truncated tail requires a second read.
  **Mitigation:** `os.Stat` + read the last byte once per file; no full reread.
- **Risk:** JSON consumers treat new keys as a schema break.
  **Mitigation:** additive, `omitempty`, documented; clean scans are
  byte-identical.
- **Risk:** Cache schema drift hides counters.
  **Mitigation:** version bump; old cache is a miss.
- **Risk:** Warnings duplicate across worker results.
  **Mitigation:** aggregate per file before rendering; sort and dedupe.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "truncated jsonl silently skipped"` and
  `sila recall "scanner err warning"`.
- **Writes**: record a `case` for the "disjoint scan counters + cache
  round-trip" mechanism; record a `lesson` for silent parse continues.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "partial-read observability shipped"
  --open="strict mode decision"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/jsonl-partial-read-observability/goal.md)` under a `feat(loaders):`
   or `fix(loaders):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the truncated, malformed, ceiling, and cache evidence.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
