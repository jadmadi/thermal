# Goal: Repository Alignment, Code Ownership, PR Lifecycle & Security Policy

## Goal Description
Establish .github/CODEOWNERS, MAINTAINERS.md, comprehensive SECURITY.md with 48h response SLA, PR/issue templates, and automated CI vulnerability scanning.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 2
- **Shape**: ship

## Files to Touch
- .github/CODEOWNERS
- SECURITY.md
- MAINTAINERS.md
- .github/PULL_REQUEST_TEMPLATE.md
- .github/ISSUE_TEMPLATE/bug_report.yml
- .github/ISSUE_TEMPLATE/loader_proposal.yml
- .github/workflows/ci.yml

## References
- **Shared Understanding & Fact Sheet**: [`goals/repo-alignment-and-security/facts.md`](facts.md)
- **Execution Plan**: [`goals/repo-alignment-and-security/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
