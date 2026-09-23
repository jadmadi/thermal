# Facts: SLSA Level 3 Build Provenance & Supply Chain Attestation via OpenSSF

## Architectural Invariants & Constraints
- Path B Hybrid Architecture: Keep GoReleaser for multi-platform archive and Homebrew packaging, chaining OpenSSF generator_generic_slsa3.yml to sign checksums.txt.
- Zero Falsification: Provenance generated inside isolated, ephemeral GitHub Actions runner using OIDC id-token federation with Sigstore Fulcio and Rekor.
- Binary Verifiability: Any released binary must be independently verifiable using slsa-verifier without requiring private keys or proprietary tools.

## File & Interface Contracts
- Relevant files:
  - `.github/workflows/release.yml`
  - `.goreleaser.yaml`
  - `docs/DISTRIBUTION.md`
  - `README.md`
