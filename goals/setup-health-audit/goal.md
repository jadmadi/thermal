# Goal: Read-Only Local Setup & Context Health Audit ('thermal audit')

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Session: `mcp-audit-goal` | Report: `receipt_setup-health-audit_1792064736`


## Goal Description
Implement non-destructive setup and context health diagnostic ('thermal audit'). Scans local agent configurations to identify unused MCP servers (consuming system prompt context without tool calls), bloated rule files (CLAUDE.md, AGENTS.md, .cursorrules > 200 lines), and repetitive file re-read loops, outputting actionable advice without mutating user files.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 103
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- cmd/thermal/main.go
- internal/audit/audit.go
- internal/audit/audit_test.go
- internal/render/audit.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/setup-health-audit/facts.md`](facts.md)
- **Execution Plan**: [`goals/setup-health-audit/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.

## Reopened acceptance, 2026-10-14

Historical implementation and completed tasks remain preserved. New acceptance is pending.
The [current remediation contract](remediation-2026-10-14.md) adds binding acceptance criteria;
complete it as well as preserved prior criteria before claiming this goal delivered.
See the [evaluation](../../docs/evaluations/2026-10-14-promise-vs-delivery.md).
