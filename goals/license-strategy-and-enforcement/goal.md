# Goal: OSS License Strategy Evaluation, CI Enforcement & Licensing Hardening

## Goal Description
Evaluate Thermal's licensing strategy using the oss-license-strategy framework. Enforce contributor CLA/DCO sign-off in CI (.github/workflows/ci.yml) with automated recovery instructions in CONTRIBUTING.md. Add canonical SPDX-License-Identifier headers across all 107 Go files. Establish mixed-asset scope boundaries in DUAL-LICENSE.md (Code: AGPLv3, Docs: CC-BY-4.0, Branding: All Rights Reserved). Elevate licensing architecture rationale into public repository documentation. Sharpen positioning to 'Thermal™ by Jad Madi — AI Agent Telemetry & Streak Engine' and anchor composite common-law trademark definitions.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- .github/workflows/ci.yml
- CONTRIBUTING.md
- DUAL-LICENSE.md
- cmd/thermal/main.go
- cmd
- internal

## References
- **Shared Understanding & Fact Sheet**: [`goals/license-strategy-and-enforcement/facts.md`](facts.md)
- **Execution Plan**: [`goals/license-strategy-and-enforcement/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
