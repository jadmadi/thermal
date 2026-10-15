# Facts: Resolve plan identifiers passed to replay --against

## Observed Evidence
- Manual reproduction (2026-10-15, isolated fixture home):

```bash
thermal replay --against claude-pro --no-color --offline
# │  claude-pro (API)            Payg        Unpriced                   —  PASS (0% throttled (unmetered API)) │
```

- `cmd/thermal/main.go:runReplayReport` passes `opts.Against` straight into `thermal.SimulateModel`, which looks the value up as a model rate card (`LookupPrice`). `claude-pro` is not in the models.dev catalog, so the row falls back to `pricingStatus: missing` + `Unpriced`.
- `internal/pricing/subscription.go:LookupPlan` already resolves plan ids and names, but it also prefix-matches (`strings.HasPrefix`), so `claude` would resolve to `claude-pro`. Any resolution used for `--against` must use exact id/name equality only.
- The simulated gate targets include `replay --against claude-pro` (label: "Replay against Claude Pro subscription"), so the intended user story is exactly this resolution; today the check only asserts exit 0 and a verdict token.

## Constraints
- Keep `--against <unknown-model>` behavior unchanged (model rate-card simulation).
- No fuzzy or prefix matching; two different targets must never collapse onto one row.
- The changed behavior is user-visible and requires a migration entry plus README wording.
- No new flags unless reviewed as necessary (prefer resolving the existing argument).

## File & Interface Contracts
- `cmd/thermal/main.go`: `runReplayReport` (`--against` and `--compare` handling).
- `internal/pricing/subscription.go`: add `LookupPlanExact(key string) (SubscriptionPlan, bool)` alongside `LookupPlan`.
- `internal/thermal/replay.go`: `SimulatePlan` (already used by `--compare`) is the plan path; `SimulateModel` stays the model path.

## Verification Contract
- Parity test: `--against <plan>` row equals `--compare <plan>` row for the same window.
- Negative test: `--against claude` stays a model lookup (no prefix capture) and does not silently resolve to `claude-pro`.
- Then the common final checks from `tasks/plan.md`.
