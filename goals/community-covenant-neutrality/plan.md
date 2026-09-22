# Plan: Community Reciprocity Covenant, Scope Demarcation & Fork Guidelines

## Execution Tasks

1. [x] **Draft the Thermal Community Reciprocity Covenant in `DUAL-LICENSE.md`**:
   - Guarantee irrevocable AGPL-3.0 availability for the complete CLI, loaders, and telemetry engine.
   - Prohibit unilateral relicensing of community contributions to non-OSI terms.
   - Demarcate future commercial offerings (e.g. enterprise multi-tenant fleet hosting, proprietary IDE embeddings) without restricting community usage.
2. [x] **Clarify Forking & Trademark Guidelines**:
   - Explicitly detail requirements for independent forks (distinct naming, removal of Thermal marks, attribution preservation).
3. [x] **Update `CONTRIBUTING.md`**:
   - Summarize the Community Covenant alongside the CLA terms to build contributor confidence.
4. [x] **Update CLI License Command (`cmd/thermal/main.go`)**:
   - Enhance `runLicense()` to include covenant guarantee in stdout.
   - Expose `"covenant"` metadata object in `thermal license --json`.
5. [x] **Update Tests & Release Gate**:
   - Add unit test in `cmd/thermal/args_test.go` verifying the JSON covenant field.
   - Update `scripts/simulated_user_gate.sh` to check license covenant compliance.
6. [x] **Verification & Validation**:
   - Run `go test -race ./...`.
   - Run `./scripts/simulated_user_gate.sh`.
- [x] **Phase 4 — Attribution Commit**: Commit with `(goals/community-covenant-neutrality/goal.md)` using a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` subject (docs/test commits are ignored), then run `sila goals`.

