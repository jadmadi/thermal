# Facts: Make truncated and malformed JSONL records observable

## Observed Evidence (2026-10-30, `/tmp/thermal-cli-review`, `thermal dev`)

Fixture: a Claude project directory with one valid assistant record and one
half-written final line, no trailing newline.

```
$ thermal daily --json --verbose
exit 0
stderr: 0 bytes
stdout:
{
  "costBasis": "total",
  "data": [{
    "activeDays": 1,
    "inputTokens": 2000,
    "outputTokens": 1000,
    "totalTokens": 419400,
    "missingPricing": ["claude-3.5-haiku"],
    ...
  }]
}
```
The valid record is counted; the truncated record is silently dropped. There
is no `skippedRecords` field and no warning even with `--verbose`.

Control: the same file with only the valid record produces the same summary,
so the two states are indistinguishable.

For contrast, an unreadable file does warn:
```
thermal: warning: Claude: <path>: open <path>: permission denied
```
And an over-ceiling line does warn:
```
<path>: line exceeds 33554432 byte ceiling, rest of file skipped
```

## Root Cause
- `internal/loaders/claude.go:95-97`:
  ```go
  if err := json.Unmarshal([]byte(line), &rec); err != nil {
      continue
  }
  ```
  No counter, no warning, no distinction between a malformed middle line and a
  truncated final line.
- The same pattern exists in:
  `commandcode.go:110-111`, `droid.go:70-71`, `grok.go:124-125`,
  `agy.go:211-212` and `:315-316`, and `jsonl.go:72-73`.
- `internal/thermal/types.go:67-89` (`Summary`) has `Warnings []string
  json:"-"` but no skipped/malformed/truncated counters.
- `internal/loaders/jsonl.go:43-51` (`formatScanWarning`) only formats
  `bufio.ErrTooLong` and generic scanner errors; it has no per-record form.
- `internal/loaders/dircache.go:35` persists `Warnings` but not parse counters;
  `:120` notes that `Summary` hides warnings from JSON.
- `internal/loaders/codexcache.go:257`/`:278` handles scanner warnings but not
  per-record parse failures.
- The worker results (`fileResult`) carry `warning` and duplicate counts, not a
  `ScanStats` aggregate.

## Architectural Invariants & Constraints
- A malformed record is a skipped record; it must be counted, not silently
  dropped.
- Counters are disjoint: malformed, truncated tail, and ceiling hit are
  separate; `SkippedRecords` is their sum.
- Blank lines and intentionally ignored record types are read, not skipped.
- Duplicates are their own counter; they are not malformed.
- Warnings go to stderr under `--verbose`; stdout and JSON stay clean.
- A partial read is valid data with a disclosure, not an error; exit `0`.
- JSON additions are `omitempty`; a clean scan is byte-identical.
- Cache hits must restore the counters and re-emit the warning.
- No prompt text, session title, or file content enters a counter or warning.
- The 32 MiB line ceiling stays; a ceiling hit is counted, not retried.
- Aggregation is thread-safe and deterministic (sorted by file path).

## File & Interface Contracts
- `internal/thermal/types.go`:
  ```go
  type ScanStats struct {
      RecordsRead    int64 `json:"recordsRead,omitempty"`
      RecordsSkipped int64 `json:"recordsSkipped,omitempty"`
      MalformedLines int64 `json:"malformedLines,omitempty"`
      TruncatedTails int64 `json:"truncatedTails,omitempty"`
      CeilingHits    int64 `json:"ceilingHits,omitempty"`
      Duplicates     int64 `json:"duplicates,omitempty"`
  }

  type Summary struct {
      // existing fields...
      SkippedRecords int64      `json:"skippedRecords,omitempty"`
      MalformedLines int64      `json:"malformedLines,omitempty"`
      TruncatedTails int64      `json:"truncatedTails,omitempty"`
      CeilingHits    int64      `json:"ceilingHits,omitempty"`
      ScanStats      *ScanStats `json:"scanStats,omitempty"`
      Warnings       []string   `json:"-"`
  }
  ```
- `internal/loaders/jsonl.go`:
  ```go
  type jsonlScan struct {
      Stats   thermal.ScanStats
      Warning string
  }

  func scanJSONL(path string, handle func(line []byte) error) jsonlScan
  func isTruncatedTail(path string, line []byte, err error) bool
  func formatRecordWarning(tool, path string, stats thermal.ScanStats) string
  ```
  `scanJSONL` reads the file once, tracks line numbers, and checks the last
  byte with a one-byte read to distinguish a truncated tail.
- `internal/loaders/claude.go` (and siblings):
  ```go
  type fileResult struct {
      // existing fields...
      stats thermal.ScanStats
  }
  ```
  On `json.Unmarshal` error: increment `MalformedLines` or `TruncatedTails`;
  do not `continue` without counting.
- `internal/loaders/dircache.go`: persist `Stats` and `Warnings`; bump the
  cache schema version.
- `internal/loaders/codexcache.go`: same for rollout files.
- `cmd/thermal/json.go`: aggregate `ScanStats` per tool and a top-level
  `skippedRecords` sum.

## Counter Taxonomy
| Event | RecordsRead | Malformed | Truncated | Ceiling | Duplicates | Skipped |
| --- | --- | --- | --- | --- | --- | --- |
| Valid record | +1 | 0 | 0 | 0 | 0 | 0 |
| Blank line | 0 | 0 | 0 | 0 | 0 | 0 |
| Non-assistant record | +1 | 0 | 0 | 0 | 0 | 0 |
| Malformed middle | +1 | +1 | 0 | 0 | 0 | +1 |
| Truncated tail | +1 | 0 | +1 | 0 | 0 | +1 |
| Over-ceiling line | +1 | 0 | 0 | +1 | 0 | +1 |
| Duplicate ID | +1 | 0 | 0 | 0 | +1 | 0 |
| Unreadable file | 0 | 0 | 0 | 0 | 0 | 0 (warning) |

## Verification Contract
```bash
B=/tmp/thermal-cli-review
H=<fixture-with-one-valid-and-one-truncated-record>

# Truncated tail is counted and warned
$B daily --json --verbose >/tmp/o 2>/tmp/e; test $? -eq 0
jq -e '.skippedRecords==1 and .truncatedTails==1' /tmp/o
grep -q 'truncated final record' /tmp/e

# Without verbose: counters, no stderr
$B daily --json >/tmp/o 2>/tmp/e; test $? -eq 0
test ! -s /tmp/e
jq -e '.skippedRecords==1' /tmp/o

# Malformed middle line
$B daily --json --verbose >/tmp/o 2>/tmp/e
jq -e '.malformedLines==1' /tmp/o

# Clean scan omits keys
$B daily --json >/tmp/o 2>/dev/null
jq -e 'has("skippedRecords")|not' /tmp/o

# Cache round-trip
$B daily --json --verbose >/tmp/o1 2>/tmp/e1
$B daily --json --verbose >/tmp/o2 2>/tmp/e2
jq -e '.skippedRecords==1' /tmp/o2
diff <(jq 'del(.generatedAt)' /tmp/o1) <(jq 'del(.generatedAt)' /tmp/o2)

# stdout purity
$B daily --json --verbose 2>/tmp/e | jq -e . >/dev/null

CGO_ENABLED=0 go test -race ./internal/loaders/... ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- One valid + one truncated final line (no trailing newline).
- One malformed middle line between two valid records.
- Blank lines and a valid non-assistant record.
- A duplicate message ID.
- A line over the lowered test ceiling.
- An unreadable file.
- A `dircache` round-trip and a `codexcache` round-trip.
- An old cache without `Stats`.

## Open Questions
- Should a partial read exit non-zero? Recommendation: no; it is valid data
  with a disclosure. A future `--strict` can opt in.
- Should `RecordsRead` count blank lines? Recommendation: no; they contain no
  record.
- Should a valid non-assistant record count as read? Recommendation: yes; it
  was parsed, but it is not skipped.
- Should the warning be one line per file or per record?
  Recommendation: one per file, with a count; per-record would flood.
- Should the cache key include the counter schema version?
  Recommendation: yes; an old cache is a miss, not a clean scan.
