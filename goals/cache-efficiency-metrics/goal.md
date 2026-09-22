# Goal: Cache Efficiency & Disjoint Token Breakdown

## Goal Description
Expose disjoint canonical token metrics (uncached input, cache read, cache write, output, reasoning) and cache efficiency ratios in `thermal stats`, `thermal <tool>`, and reports.

In modern agentic workflows (Claude Code, Codex, OpenCode), multi-turn agents re-transmit large repository context repeatedly. Across real-world developer workloads (as highlighted by Tyler Southwick's analysis of 28.7B tokens in StackReplay), 98%+ of total tokens are prompt cache reads (`CacheRead`), while actual generated output is typically <1% and uncached input is ~1%.

While Thermal internally enforces disjoint token buckets in `ModelTokens` (Rule A.6), the terminal interface rolls everything into a single `Tokens` number. Developers seeing billions of tokens currently cannot see *why* their volume is so high or evaluate their cache hit rate.

This goal adds:
1. Disjoint token composition analysis (`Cache Read`, `Cache Write`, `Uncached Input`, `Output`, `Reasoning`) with Unicode bar visuals in `thermal stats`.
2. Cache hit rate percentage and cache efficiency metrics.
3. Summary disclosure in single-tool views and period reports under `--breakdown`.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/cache-efficiency-metrics/facts.md`](facts.md)
- **Execution Plan**: [`goals/cache-efficiency-metrics/plan.md`](plan.md)

## Done Condition
1. `thermal stats` displays a dedicated "Token composition & cache efficiency" section with counts, percentages, and Unicode visual bars.
2. Cache hit rate (`CacheRead / (UncachedInput + CacheRead + CacheWrite)`) is calculated safely with zero-division protection.
3. Single-tool dashboards or `--breakdown` report rows surface cache read shares.
4. Unit tests in `stats_test.go` and `analytics_test.go` verify accuracy and formatting.
5. All tests pass with `go test -v -race ./...` and `golangci-lint` passes with zero issues.
