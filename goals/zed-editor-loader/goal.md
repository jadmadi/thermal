# Goal: Zed Editor Data Ingestion & Heatmap Loader

## Goal Description
Ingest Zed AI assistant threads and tokens from threads.db into Thermal

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## References
- **Shared Understanding & Fact Sheet**: [`goals/zed-editor-loader/facts.md`](facts.md)
- **Execution Plan**: [`goals/zed-editor-loader/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
