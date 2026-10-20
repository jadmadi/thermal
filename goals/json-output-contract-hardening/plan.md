# Plan: Standardize JSON envelopes, row schemas, precision, and alias fields

## Execution Steps
- [x] **Task 1 — Define the envelope/rounding contract**: Write docs/JSON-CONTRACT.md with envelope keys, timestamp rule, rounding, row schemas, and share fields. Acceptance: the document covers every `--json` surface and names the removal/deprecation policy for aliases.
- [x] **Task 2 — Remove aliases and add missing fields**: Delete duplicate leaderboard keys; add `type`/totals to leaderboard, `unattributedTokens` to models/stats, and cost/scope/window/activeDays to share. Acceptance: `jq '.results[0]|keys'` has no uppercase aliases; every displayed share field is non-null.
- [x] **Task 3 — Apply rounding at serialization**: Add JSON rounding helpers for costs, percentages, shares, and means; wire all payloads. Acceptance: `TestJSONPrecision` rejects >4 decimal places and the gallery examples are stable.
- [x] **Task 4 — Add contract tests and docs**: Add envelope/alias/precision/share tests and update docs/MIGRATION/README/project.jsonc. Acceptance: `go test -race ./cmd/thermal -run TestJSON` passes; migration has six parts.
- [x] **Task 5 — Full gate and attribution commit**: Run focused/full tests, build, gate; commit with the goal token. Acceptance: gate clean; commit subject ends in `(goals/json-output-contract-hardening/goal.md)`.

## Dependencies and Handoff
The share/statusline goal consumes the share JSON contract; the gallery goal consumes the row-level schemas. The reconciliation goal owns the numeric values; this goal only makes them stable and machine-readable.

## Verification
```bash
go test -race ./cmd/thermal ./internal/thermal ./internal/share -run 'JSON|Envelope|Precision|Alias'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
