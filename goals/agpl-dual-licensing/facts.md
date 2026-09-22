# Facts: AGPL-3.0 Relicensing, Commercial Dual-Licensing & Contributor Gating

## Architectural Invariants & Constraints

1. **Inbound Compatibility**:
   - All 28 direct and indirect dependencies in `go.mod` are licensed under permissive terms: MIT (17), BSD-3-Clause (11), and Public Domain (SQLite core).
   - Inbound permissive code combined into an outbound AGPL-3.0 binary is 100% valid under open-source copyright law, provided original copyright notices and disclaimers are preserved in `NOTICES.md` or license outputs.
2. **Author Attribution**:
   - Jad Madi is the sole human author of all commits in the repository.
   - All occurrences of author name must strictly be **Jad Madi** (correcting legacy typo "Jafar").
3. **Dual-Licensing Mechanics**:
   - Open source community track: GNU Affero General Public License v3.0 (AGPL-3.0). Requires source code disclosure upon network interaction or binary distribution of modified works.
   - Commercial track: Paid commercial license for proprietary embedding, closed-source SaaS/cloud leaderboards, and enterprises with internal AGPL bans.
   - Past releases (prior to this transition commit/tag) permanently remain under the MIT License for those who obtained them.
4. **Contributor Gating (CLA)**:
   - To preserve the author's legal ability to sell commercial licenses containing community contributions, all future pull requests require a Contributor License Agreement (CLA) in `CONTRIBUTING.md`.
   - Contributors retain their copyright, contribute under AGPL-3.0, and grant Jad Madi a worldwide, perpetual, non-exclusive license to sublicense and sell under commercial terms.
5. **CLI Integration**:
   - `thermal license` and `--license` must be first-class commands in `cmd/thermal/main.go`.
   - Must support plain terminal formatting and `--json` structured output.
   - Must be verified in `scripts/simulated_user_gate.sh`.

## File & Interface Contracts

- `LICENSE`: GNU Affero General Public License v3.0 text with `Copyright (C) 2026 Jad Madi`.
- `DUAL-LICENSE.md`: Plain-language explanation of Community AGPL vs Commercial Enterprise tracks.
- `NOTICES.md`: Preserved third-party copyright notices and disclaimers for all 28 dependencies.
- `CONTRIBUTING.md`: Contribution guide and Contributor License Agreement (CLA) terms.
- `cmd/thermal/main.go`:
  - `license` positional word and `--license` flag.
  - `runLicense(asJSON bool)` function rendering dual-license summary and contact info.
- `scripts/simulated_user_gate.sh`: Automated release gate tests for `thermal license` and license integrity.
