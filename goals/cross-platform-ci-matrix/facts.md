# Facts: Cross-Platform Operating System Test Matrix in CI

## Architectural Invariants & Constraints
- Keep CI fast and deterministic across all matrix runners.
- Linux runners must continue running race detection (`go test -race`).
- macOS and Windows runners must execute full unit tests (`go test ./...`).
- Zero secret or token leaks across matrix runner logs.

## File & Interface Contracts
- Relevant files:
  - `.github/workflows/ci.yml`: defines GitHub Actions jobs and matrix runners.
  - `scripts/simulated_user_gate.sh`: verifies CLI behavior in hermetic fixture environments.

