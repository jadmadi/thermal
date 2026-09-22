# Plan: OSS License Strategy Evaluation, CI Enforcement & Licensing Hardening

## Execution Steps

- [x] **Phase 1 — Evaluation & Architecture Strategy**:
  - Evaluate Thermal using the `oss-license-strategy` framework.
  - Audit inbound dependencies (28 modules: 17 MIT, 11 BSD-3, 1 Public Domain -> 100% directional compatibility).
  - Verify Jad Madi as sole author in git log.
  - Establish strategic posture (Commercial Optionality) and eliminate alternatives with written justification.

- [x] **Phase 2 — CI Enforcement & Contributor Guide**:
  - Implement zero-dependency native bash commit trailer check in `.github/workflows/ci.yml`.
  - Configure `actions/checkout` with `fetch-depth: 0`.
  - Document CLA justification and PR recovery commands in `CONTRIBUTING.md`.

- [x] **Phase 3 — Source File License Hygiene (SPDX Headers)**:
  - Add canonical `// Copyright (C) 2026 Jad Madi. All rights reserved.` and `// SPDX-License-Identifier: AGPL-3.0-only` headers across all 107 Go source files.
  - Ensure build tags (`//go:build race` / `//go:build !race`) are preserved cleanly.
  - Verify formatting via `gofmt -l .` (0 unformatted files).

- [x] **Phase 4 — Asset Boundaries & Public Architectural Record**:
  - Update `DUAL-LICENSE.md` with:
    - Sharpened title: `Thermal™ by Jad Madi — AI Agent Telemetry & Streak Engine`.
    - Section 4: Composite trademark policy (`jadmadi/thermal™`, "Thermal by Jad Madi", "Don't break the streak").
    - Section 5: Repository asset scope boundaries (Code: AGPLv3, Docs: CC-BY-4.0, Branding: All Rights Reserved).
    - Section 6: Public licensing architecture & decision record (removing reliance on `goals/`).
  - Update `cmd/thermal/main.go` license banner to match sharpened copy.

- [x] **Phase 5 — Quality Gates & Verification**:
  - Run `go vet ./...` (passes clean).
  - Run `go test -v -race ./...` (all 6 packages pass cleanly).
  - Run `./scripts/simulated_user_gate.sh` (46/46 checks pass).
  - Verify `thermal license` and `thermal license --json` output.

- [ ] **Phase 6 — Attribution Commit**:
  - Commit changes with subject: `feat(license): harden licensing architecture, CI enforcement & SPDX headers (goals/license-strategy-and-enforcement/goal.md)`
