# Goal: Resolve plan identifiers passed to replay --against

## Goal Description
`thermal replay --against claude-pro` silently produces a pay-as-you-go row named `claude-pro (API)` with `pricingStatus: missing` and `Unpriced`, because `--against` only resolves model rate cards while `claude-pro` is a built-in subscription plan. Users reasonably expect the plan they named; today they get an unpriced pseudo-model with no hint that `--compare claude-pro` was meant.

Observed:

```
  │  claude-pro (API)            Payg        Unpriced                   —  PASS (0% throttled (unmetered API)) │
```

Detect known plan identifiers passed to `--against` and resolve them as plans (equivalent to `--compare <plan>`), while keeping model rate-card behavior for anything that is not a known plan.

Priority: **P2**. Identified during the workload-replay remediation audit.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 109
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `cmd/thermal/main.go` (`runReplayReport` `--against` wiring)
- `internal/pricing/subscription.go` (add an exact-match lookup; do not reuse prefix matching for this)
- `internal/thermal/replay.go` / `internal/thermal/replay_test.go` (seam for plan vs model targeting)
- `cmd/thermal/simulated_user_test.go` (targets test expectations)
- `docs/MIGRATION.md` (behavior change entry)
- `README.md` (replay `--against` note)

## References
- [Evidence and constraints](facts.md)
- [Execution plan](plan.md)

## Done Condition
1. `thermal replay --against <plan-id-or-name>` renders the subscription plan row (monthly fee, pricingStatus complete, capacity verdict) exactly as `--compare <plan>` does for that plan.
2. Unknown identifiers keep today's model rate-card behavior; a model whose name happens to share a plan's prefix (`claude`, `chatgpt`) is **not** captured by fuzzy prefix matching — resolution is exact id or exact name, case-insensitive.
3. `--against` with a plan id and `--compare` with the same plan id produce identical rows (parity test).
4. A six-part `docs/MIGRATION.md` entry documents the changed `--against` semantics; README wording matches.
5. Focused verification, full race suite, CLI build and simulated user gate pass; implementation-attributed commit before closure.

## Scope Boundary
No new plans, prices, or simulation math. No change to `--compare`, defaults, or JSON schema keys. This is argument resolution and documentation only.
