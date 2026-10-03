# Facts: Multi-Target Cross-Compilation Verification Gate in CI

## Architectural Invariants & Constraints
- Validate that all 5 official release targets compile cleanly in CI:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`
- All builds must enforce `CGO_ENABLED=0` and `-trimpath`.
- Cross-compilation gate must execute quickly (under 60 seconds total).

## File & Interface Contracts
- Relevant files:
  - `.github/workflows/ci.yml`: CI build and verification workflow.
  - `build.sh`: local cross-compilation reference script.

