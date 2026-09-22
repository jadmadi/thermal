# Goal: Coverage-First Distribution: Homebrew Tap, Standalone Installer, Agent Docs \u0026 Curated Showcases

## Goal Description
Implement Thermal's Coverage-First distribution mix: 1. Configure automated Homebrew tap formula generation in .goreleaser.yml (jadmadi/homebrew-tap). 2. Author and deploy standalone zero-dependency POSIX shell installer (docs/pages/install.sh) with SHA-256 verification. 3. Author and deploy machine-readable agent documentation (docs/pages/llms.txt and llms-full.txt). 4. Align repository discovery metadata and topics (claude, grok, deepseek, hermes, llm-metrics). 5. Author docs/DISTRIBUTION.md with curated list PR packages (awesome-go, awesome-ai-tools), referral tracking, and quarterly review sweep calendar.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: release-communication-and-changelog, contributor-onboarding
- **Sequence**: 10
- **Shape**: ship

## Files to Touch
- .goreleaser.yml
- docs/pages/install.sh
- docs/pages/llms.txt
- docs/pages/llms-full.txt
- docs/DISTRIBUTION.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/distribution-coverage-foundation/facts.md`](facts.md)
- **Execution Plan**: [`goals/distribution-coverage-foundation/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
