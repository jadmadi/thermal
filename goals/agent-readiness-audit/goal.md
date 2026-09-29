# Goal: Repository Agent Readiness and AGENTS.md Audit

## Goal Description
Extend thermal audit with an agent-readiness check verifying local repository instruction hygiene (presence and validity of AGENTS.md, gitignore rules for agent databases, and memory-mapped file access permissions).

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 77
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- internal/audit/audit.go
- internal/audit/audit_test.go
- internal/render/audit.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/agent-readiness-audit/facts.md`](facts.md)
- **Execution Plan**: [`goals/agent-readiness-audit/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
