# Goal: JSONL line ceilings that fit real logs

## Goal Description
A 256KB scanner ceiling is smaller than a real command-code transcript line (409,473 bytes), so one session stops at line 283 of 320 and the 38 messages after it never enter the counts; the two days they belong to currently show no activity at all. Raise the ceiling for every JSONL scanner to a size real data fits, prove with tests that a line over the old cap is read in full, and show the affected days recovering their counts.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 12
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/jsonl-line-ceilings/facts.md`](facts.md)
- **Execution Plan**: [`goals/jsonl-line-ceilings/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
