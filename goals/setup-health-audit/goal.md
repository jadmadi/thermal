# Goal: Read-Only Local Setup & Context Health Audit ('thermal audit')

## Goal Description
Implement non-destructive setup and context health diagnostic ('thermal audit'). Scans local agent configurations to identify unused MCP servers (consuming system prompt context without tool calls), bloated rule files (CLAUDE.md, AGENTS.md, .cursorrules > 200 lines), and repetitive file re-read loops, outputting actionable advice without mutating user files.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
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
