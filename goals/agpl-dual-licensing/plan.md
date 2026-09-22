# Plan: AGPL-3.0 Relicensing, Commercial Dual-Licensing & Contributor Gating

## Execution Steps

- [x] **Phase 1 — Core Relicensing & Attribution Artifacts**:
  - Update `LICENSE` from MIT to GNU Affero General Public License v3.0 (AGPL-3.0) with `Copyright (C) 2026 Jad Madi`.
  - Create `DUAL-LICENSE.md` documenting the Open Source Community (AGPL-3.0) and Commercial/Enterprise tracks.
  - Create `NOTICES.md` preserving all 28 third-party MIT, BSD-3, and Public Domain copyright notices.
- [x] **Phase 2 — Contributor License Agreement (CLA)**:
  - Create `CONTRIBUTING.md` defining contribution guidelines, DCO/CLA requirements, and commercial sublicensing grant to protect dual-licensing rights.
- [x] **Phase 3 — CLI `thermal license` Command**:
  - Add `license` command and `--license` flag to `cmd/thermal/main.go`.
  - Implement text display and `--json` format reporting license details, dual-license availability, and author contact.
  - Update CLI usage text in `cmd/thermal/main.go`.
  - Add unit tests in `cmd/thermal/simulated_user_test.go`.
- [x] **Phase 4 — Release Gate & Documentation**:
  - Add `thermal license` validation and license integrity checks to `scripts/simulated_user_gate.sh`.
  - Update `README.md` and `AGENTS.md` to document the AGPL-3.0 dual-license model.
  - Verify all unit tests with `go test -v -race ./...`.
  - Run `./scripts/simulated_user_gate.sh` to ensure exit code 0 across all checks.
- [x] **Phase 5 — Goal Attribution Commit & Auto-Reconciliation**:
  - Commit with `(goals/agpl-dual-licensing/goal.md)` and update `goals/README.md`.
