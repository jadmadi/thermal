# Goal: Contributor Onboarding Path, Preflight Gate & Retention Loop

## Goal Description
Establish Thermal's contributor onboarding path to pass all 4 blocking checks (B1–B4) and score at least 12 of 16 points on the `oss-contributor-onboarding` rubric. 
1. Build a unified local preflight script (`scripts/check.sh`) bundling commit sign-off checks, `gofmt -l .`, `go vet`, `go test -race ./...`, and `./scripts/simulated_user_gate.sh` to give contributors 100% local CI parity.
2. Seed the initial stock of 3 high-impact Good First Issues adhering to the 7-part anatomy (expected behavior, file location, skills required, approach, test verification, mentor) in `.github/GOOD_FIRST_ISSUES.md`.
3. Enhance `CONTRIBUTING.md` with:
   - Explicit Out-of-Scope boundaries harmonized with `GOVERNANCE.md`.
   - Published 72h review response target and Two-Clock communication expectations.
   - Issue claiming conventions (`/claim`, 7-day stale timeout).
   - Single-command local verification instructions (`./scripts/check.sh`).
4. Author `docs/CONTRIBUTOR_PLAYBOOK.md` establishing the maintainer's Two-Clock review procedure, fast/kind decline guidelines, release note recognition, and second-contribution retention loop.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: governance-foundations, repo-alignment-and-security, community-covenant-neutrality
- **Sequence**: 6
- **Shape**: ship

## Files to Touch
- scripts/check.sh
- CONTRIBUTING.md
- docs/GOOD_FIRST_ISSUES.md
- docs/CONTRIBUTOR_PLAYBOOK.md
- README.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/contributor-onboarding/facts.md`](facts.md)
- **Execution Plan**: [`goals/contributor-onboarding/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. `scripts/check.sh` runs cold and passes all local gates with zero errors.
3. `CONTRIBUTING.md` articulates claiming rules, response targets, local preflight, and out-of-scope boundaries without contradicting `GOVERNANCE.md` or `DUAL-LICENSE.md`.
4. At least 3 fully specified Good First Issues are documented and ready for triage.
5. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` and `./scripts/simulated_user_gate.sh`).
6. Zero architectural regressions; definitions of done satisfied.
