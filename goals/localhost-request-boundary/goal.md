# Goal: Reject untrusted hosts at the local telemetry API

## Goal Description
Validate request authority before loading private local telemetry while preserving explicit host binding behavior.

Priority: **P1**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Dependent (sequenced)
- **Depends On**: web-telemetry-parity
- **Sequence**: 66
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/server/server.go`
- `internal/server/server_test.go`
- `docs/MIGRATION.md`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. Untrusted request authorities cannot read telemetry or trigger its collection.
2. Explicitly configured local access and IPv4/IPv6 work with a narrowly specified policy.
3. Header tests are supplemented with behavioral rejection tests and migration guidance.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
No account system, external identity provider or remote telemetry service. Coordinate with web-telemetry-parity to avoid concurrent edits to server.go.
