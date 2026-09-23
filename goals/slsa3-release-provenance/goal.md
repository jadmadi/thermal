# Goal: SLSA Level 3 Build Provenance & Supply Chain Attestation via OpenSSF

## Goal Description
Implement SLSA Level 3 build provenance and cryptographic attestation for Thermal releases using the OpenSSF generic SLSA3 generator with GoReleaser. Generates non-falsifiable in-toto provenance (checksums.txt.intoto.jsonl) via Sigstore keyless OIDC, satisfying enterprise OSPO compliance, NIST SSDF, and Executive Order 14028.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- .github/workflows/release.yml
- .goreleaser.yaml
- docs/DISTRIBUTION.md
- README.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/slsa3-release-provenance/facts.md`](facts.md)
- **Execution Plan**: [`goals/slsa3-release-provenance/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
