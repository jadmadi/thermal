# Goal: A scan that stops early is reported, never silent

## Goal Description
No loader checks scanner.Err(), so a line over the ceiling or a read error ends the scan and the rest of the file disappears with no signal; the only diagnostics channel today is a fatal LoadToolData error printed under --verbose. Add the check to every JSONL scan and a non-fatal warnings channel that reaches --verbose stderr without touching stdout or JSON.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: jsonl-line-ceilings
- **Sequence**: 13
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/jsonl-scan-diagnostics/facts.md`](facts.md)
- **Execution Plan**: [`goals/jsonl-scan-diagnostics/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
