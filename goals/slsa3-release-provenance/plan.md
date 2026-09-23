# Plan: SLSA Level 3 Build Provenance & Supply Chain Attestation via OpenSSF

## Execution Steps
- [x] Phase 1 — GoReleaser Artifact Configuration: Create or update .goreleaser.yaml with checksums.txt generation and multi-platform packaging.
- [x] Phase 2 — SLSA3 Reusable Workflow Chaining: Update .github/workflows/release.yml with id-token permissions and call slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml.
- [x] Phase 3 — Documentation & Verification Instructions: Document slsa-verifier verification instructions in docs/DISTRIBUTION.md and README.md.
