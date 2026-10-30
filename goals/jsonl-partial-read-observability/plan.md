# Plan: Make truncated and malformed JSONL records observable

## Execution Steps
- [x] **Phase 1 — ScanStats model**: Add `ScanStats` and the `Summary` fields;
  define the disjoint taxonomy and `SkippedRecords` derivation. Acceptance:
  `TestScanStatsDisjoint` and `TestCleanScanOmitsKeys` pass.
- [x] **Phase 2 — Shared scanner and loader adoption**: Add `scanJSONL` with
  line numbers and truncated-tail detection; replace the silent `continue`
  sites in Claude, commandcode, Droid, Grok, Agy, DSH, and the generic JSONL
  loader. Acceptance: `TestTruncatedTail`, `TestMalformedMiddle`, and
  `TestBlankAndNonAssistantNotSkipped` pass.
  - Malformed middle -> `MalformedLines`.
  - Unterminated final line -> `TruncatedTails`.
  - Ceiling -> `CeilingHits`.
  - Duplicate -> `Duplicates`.
- [x] **Phase 3 — Warnings and JSON**: Add `formatRecordWarning`, emit one
  stderr line per file under `--verbose`, aggregate per tool, and add the
  top-level `skippedRecords` sum. Acceptance: `TestPartialReadJSON`,
  `TestWarningsStderrOnly`, and `TestCleanScanByteIdentical` pass.
- [x] **Phase 4 — Cache round-trip**: Persist `ScanStats` in `dircache` and
  `codexcache`, bump the schema version, treat an old cache as a miss, and
  re-emit warnings on a hit. Acceptance: `TestScanStatsCacheRoundTrip`,
  `TestOldCacheIsMiss`, and `TestCachedWarningReEmitted`.
- [x] **Phase 5 — Docs, gate, evidence, closure**: Update
  `docs/JSON-CONTRACT.md`, README, MIGRATION, `project.jsonc`, the simulated
  gate; record evidence; commit, certify, reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | truncated fixture, `--json --verbose` | truncatedTails 1, warning |
| AC-002 | truncated fixture, no verbose | counters, empty stderr |
| AC-003 | malformed middle | malformedLines 1, valid records kept |
| AC-004 | blank/non-assistant | no skipped |
| AC-005 | ceiling | ceilingHits 1, warning |
| AC-006 | duplicate ID | duplicates 1, not skipped |
| AC-007 | clean scan | no new keys, byte-identical |
| AC-008 | cache round-trip | counters + warning survive |
| AC-009 | old cache | miss, rescan |
| AC-010 | `--json --verbose` | stdout parses, stderr warnings |
| AC-011 | `go test -race ...` | exit 0 |
| AC-012 | simulated user gate | exit 0 |

## Rollback
- Keep the counters additive and `omitempty`; if a consumer rejects them,
  document a schema version rather than removing the disclosure.
- If per-record warnings flood, keep the per-file count form and drop the
  line number, not the counter.
- If the cache schema bump is too disruptive, version the cache key and treat
  the old entry as a miss.
- Never return to an uncounted `continue`; at minimum increment a local
  counter.

## Phase 6 — Attribution & Registry Closure
- [ ] **Phase 6 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `feat(loaders): count and report skipped JSONL records
  (goals/jsonl-partial-read-observability/goal.md)`; (3) certify via
  `sila goals certify jsonl-partial-read-observability
  --evidence=goals/jsonl-partial-read-observability/evidence.md`; (4) run
  `sila goals` and commit the updated `goals/README.md` and certification
  receipts; verify `sila goals --json` reports `unverified_count: 0`.
