# Goal: Make transcript verification independent of ambient tool-root overrides

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Session: `mcp-audit-goal` | Report: `receipt_test-env-isolation_1792088896`


## Goal Description
`TestScanSessionReceipts_Codex` fails whenever `CODEX_HOME` points at a real or empty tool store:

```bash
CODEX_HOME=/tmp/empty-codex-home go test ./internal/thermal -run TestScanSessionReceipts_Codex
# receipt_test.go: "expected 1 codex receipt, got 0" — FAIL
```

The production path (`ScanSessionReceipts` in `internal/thermal/receipt.go`) correctly honors `CODEX_HOME`, but the test builds its own temporary home and then loses to the ambient override. Verification must not depend on the machine's personal tool stores; every test that can read a tool-root override must isolate it.

Priority: **P2**. Follow-up to [hermetic-verification-gates](../hermetic-verification-gates/goal.md) (verification independent of personal stores).

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 108
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/receipt_test.go` (and any sibling test exercising env-reading scans)
- `internal/loaders/*_test.go` (any test that can hit `HERMES_HOME`, `DSH_HOME`, `GROK_HOME`, `AGY_HOME`, `ANTIGRAVITY_APP_DATA_DIR`, `OPENCODE_HOME`, `ZED_HOME`)
- `internal/audit/*_test.go` and `internal/render/*_test.go` if the audit sweep finds them affected
- `docs/PROMISES.md` (evidence note if wording changes)

## References
- [Evidence and constraints](facts.md)
- [Execution plan](plan.md)

## Done Condition
1. A poisoned-environment run of the full suite (`CODEX_HOME`, `GROK_HOME`, `DSH_HOME`, `HERMES_HOME`, `ANTIGRAVITY_APP_DATA_DIR`, `THERMAL_CONFIG` pointed at empty temp dirs) exits 0 for every affected package.
2. Tests that need an override set it explicitly with `t.Setenv`; tests that must not see one clear it (`t.Setenv("CODEX_HOME", "")`) so the test's own fixture wins.
3. No production behavior changes; no assertion is weakened to make the poisoned run pass.
4. Focused verification, full race suite, CLI build and simulated user gate pass; implementation-attributed commit before closure.

## Scope Boundary
Test isolation only. Do not change loader discovery semantics or the overrides themselves. Where a production path is genuinely untestable without an override, prefer `t.Setenv` over new globals or interfaces.
