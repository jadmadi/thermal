# Goal: Workload Replay Simulation (`thermal replay`)

## Goal Description
Implement the `thermal replay` command to simulate historical developer workloads across alternative AI model price cards and subscription plan limits.

The command answers two foundational developer questions:
1. **"What would've handled it?"** Evaluates historical daily token volume against subscription rate limits and rolling window capacities (e.g. Claude Pro, Claude Max, ChatGPT Plus, Cursor Pro) to report pass/fail verdicts and throttling percentages.
2. **"What would it have cost?"** Replays exact historical disjoint tokens (`UncachedInput`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`) against target model rate cards from the `models.dev` catalog or predefined subscription packages.

### CLI Syntax
```bash
# Replay workload against a specific model rate card:
thermal replay --against claude-3-7-sonnet
thermal replay --against deepseek-v3 --since 2026-08-01

# Compare multiple subscriptions and API tiers side-by-side:
thermal replay --compare "claude-pro,chatgpt-plus,cursor-pro,deepseek-api"

# Export replay analysis to JSON:
thermal replay --compare all --json
```

## Dependencies & Execution Order
- **Mode**: Independent ⚡
- **Depends On**: none
- **Shape**: ship

## Files to Touch
- `cmd/thermal/main.go`: CLI command dispatch and flag registration (`--against`, `--compare`).
- `internal/thermal/replay.go`: Core simulation engine, disjoint token pricing replay, and throttle calculation.
- `internal/pricing/subscription.go`: Built-in subscription profile catalog (price, daily capacity, burst limits).
- `internal/render/replay.go`: Terminal presentation table, capacity verdicts, and recommendation highlights.
- `internal/thermal/replay_test.go` & `internal/render/replay_test.go`: Unit tests with race detection.

## References
- **Shared Understanding & Fact Sheet**: [`goals/workload-replay/facts.md`](facts.md)
- **Execution Plan**: [`goals/workload-replay/plan.md`](plan.md)

## Done Condition
1. `thermal replay` is recognized as a valid first-class command with `--against` and `--compare` flags.
2. Replay accurately prices disjoint tokens using prompt cache discounts.
3. Subscription capacity checks evaluate historical daily bursts and report throttle rates (`% of days throttled`).
4. Output displays an executive summary table comparing projected monthly cost, cost delta vs actual, and capacity verdict.
5. Automated unit tests verify replay math, throttle calculations, and golden terminal rendering.
6. Documentation updated (`README.md`, `AGENTS.md`, `cmd/thermal/main.go --help`, and HTML site).
