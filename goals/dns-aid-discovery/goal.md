# Goal: Publish DNS for AI Discovery (DNS-AID) records for DNS-based agent discovery

## Goal Description
Publish DNS for AI Discovery (DNS-AID) records under domain (_index._agents.thermal.jadmadi.net or _a2a._agents) using ServiceMode SVCB/HTTPS records with alpn and endpoint parameters, signed with DNSSEC.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/dns-aid-discovery/facts.md`](facts.md)
- **Execution Plan**: [`goals/dns-aid-discovery/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
