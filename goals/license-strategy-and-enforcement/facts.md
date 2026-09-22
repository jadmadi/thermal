# Facts: OSS License Strategy Evaluation, CI Enforcement & Licensing Hardening

## Architectural Invariants & Constraints

1. **Inbound Compatibility & Verification**:
   - All 28 direct and indirect dependencies in `go.mod` are 100% permissive (MIT, BSD-3-Clause, Public Domain). Directional combination into an outbound AGPL-3.0 binary is fully compliant with open-source copyright law.
   - Third-party notices and disclaimers are preserved in `NOTICES.md` and exposed via `thermal license --json`.

2. **Automated Contributor Gating (CLA / DCO)**:
   - Thermal is dual-licensed (AGPLv3 Community + Commercial Enterprise). External contributions require a Contributor License Agreement (CLA) granting Jad Madi commercial sublicensing rights to prevent copyright deadlock and maintain the legal ability to offer commercial licenses.
   - Mandating `Signed-off-by` in `CONTRIBUTING.md` alone is insufficient ("a policy enforced by human memory is folklore"). CI must strictly enforce the trailer on every pull request commit.
   - Verification implementation: Native bash step in `.github/workflows/ci.yml` checking `git rev-list --no-merges origin/${base}..HEAD` with regex `^Signed-off-by: [^<]+ <[^@]+@[^>]+>$`.
   - Recovery commands must be documented in `CONTRIBUTING.md` and printed directly on CI failure: `git commit --amend -s` (single commit) and `git rebase --signoff origin/main` (multi-commit).

3. **Source File License Hygiene (SPDX Headers)**:
   - All 107 Go source files across `cmd/` and `internal/` must carry the canonical SPDX identifier:
     `// Copyright (C) 2026 Jad Madi. All rights reserved.`
     `// SPDX-License-Identifier: AGPL-3.0-only`
   - Build tags (`//go:build race`, `//go:build !race`) must be respected at the top of files, followed by blank line and SPDX header.
   - Preserves `gofmt` compliance and godoc package comments.

4. **Mixed Asset Scope Boundary**:
   - A single root `LICENSE` silently claims all repository assets unless explicit boundaries are defined.
   - Documented in `DUAL-LICENSE.md`:
     1. Software & CLI Source Code (`cmd/`, `internal/`, `build.sh`): GNU AGPLv3 (Community) / Commercial Enterprise License.
     2. Documentation & Markdown Guides (`docs/`, `*.md`): Creative Commons Attribution 4.0 International (CC-BY-4.0).
     3. Visual Branding, Screenshots & Logos (`docs/pages/*.webp`, `og.png`, root `*.webp`): All Rights Reserved by Jad Madi.

5. **Public Architectural Record Outside `goals/`**:
   - Internal task management folders (`goals/`) must not be the source of truth for public architectural or licensing rationale.
   - `DUAL-LICENSE.md` permanently records:
     - Why AGPL-3.0 was chosen (closing the SaaS/cloud network loophole under Section 13).
     - Why dual-licensing was adopted (enterprise OSPO compliance bypass, commercial embedding).
     - Why alternatives were eliminated (Permissive allows closed SaaS exploitation; BUSL/SSPL breaks OSI standards; GPLv3 leaves cloud loophole).
     - Past version boundary (v0.1.0 remains MIT; v0.1.1+ AGPLv3).

6. **Trademark Protection & Positioning**:
   - "Thermal" is a generic dictionary word. Solely claiming "Thermal" across all computer software is weak against prior art.
   - Protected as a composite mark: **`jadmadi/thermal™`** and **"Thermal™ by Jad Madi — AI Agent Telemetry & Streak Engine"**, including the flame/heatmap branding and tagline **"Don't break the streak"**.
   - Asserted through open-source use in commerce under *Planetary Motion v. Techsplosion* (261 F.3d 1188).
   - Unofficial forks are permitted under AGPLv3 but must rebrand and remove all Thermal trademarks.

## File & Interface Contracts

- `.github/workflows/ci.yml`: `Verify Contributor Sign-off (CLA / DCO)` step on `pull_request` with `fetch-depth: 0`.
- `CONTRIBUTING.md`: Sign-off instructions, CLA justification, and PR recovery commands.
- `DUAL-LICENSE.md`:
  - Sharpened title: `Thermal™ by Jad Madi — AI Agent Telemetry & Streak Engine`.
  - Section 4: Refined Trademark Policy.
  - Section 5: Asset Scope & Licensing Boundaries.
  - Section 6: Public Licensing Architecture & Decision Record.
- `cmd/thermal/main.go`: Updated `runLicense()` banner output.
- `cmd/**/*.go` and `internal/**/*.go`: Canonical `SPDX-License-Identifier: AGPL-3.0-only` headers.
- `scripts/simulated_user_gate.sh`: Release gate verification of license integrity.
