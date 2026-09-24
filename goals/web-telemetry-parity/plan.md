# Plan: Make web totals match canonical reports

## Execution Steps
- [ ] **Task 1 — Define a canonical snapshot adapter**: Use existing thermal aggregators over the same loaded rows; avoid introducing a second analytics engine. Preserve activity for streaks while excluding it from tokens. Carry recorded, estimated and unpriced coverage correctly.

  Acceptance and verification: Mixed token/activity fixtures match AggregateStats and AggregateProjects; a recorded-cost day plus estimated day is accounted per day.

- [ ] **Task 2 — Apply or reject collection options explicitly**: Thread no-estimate and supported tool/window options from CLI to JSON and HTTP collection; reject intentionally unsupported combinations with explicit errors and migration entries.

  Acceptance and verification: serve --json --no-estimate performs no pricing fetch; a tool filter cannot return other tools; accepted windows affect totals or fail clearly.

- [ ] **Task 3 — Test REST and SSE snapshot parity**: Inject fixed loaded rows/clock/pricer to isolate collection from local databases; assert totals and metadata across JSON, REST and SSE.

  Acceptance and verification: Run `go test -race ./internal/server ./internal/thermal`; reuse an identical fixture across static and web totals.

- [ ] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [ ] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/web-telemetry-parity/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
No prerequisites. Coordinate shared test/document files if another goal is active.

## Verification
`go test -race ./internal/server ./internal/thermal`

No visual redesign, new telemetry store or cloud endpoint. Preserve established JSON keys unless a documented migration is necessary. Follow ui-craft if frontend rendering changes become necessary.
