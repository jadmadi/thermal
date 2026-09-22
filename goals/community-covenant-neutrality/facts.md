# Facts: Community Reciprocity Covenant, Scope Demarcation & Fork Guidelines

## Architectural Invariants & Constraints

1. **The Contributor Trust Problem in Dual-Licensed Projects**:
   - As documented in open-source neutrality audits (e.g. Redis, MongoDB, Elastic, Terraform), projects operating under dual-licensing with an asymmetric CLA risk contributor alienation if the maintainer's commercial rights are perceived as unbounded.
   - Contributors need legal and moral certainty that their AGPL-3.0 contributions will not be appropriated into closed-source or source-available products (like BUSL or SSPL) while community access is deprecated.

2. **Thermal Community Reciprocity Covenant**:
   - Add a formal **Community Covenant** to `DUAL-LICENSE.md` and `CONTRIBUTING.md`:
     - **Irrevocable AGPL Guarantee**: Thermal commits that the core CLI binary, all agent ingestion loaders, the terminal TUI, and local analytics engines will remain available under an OSI-approved copyleft license (GNU AGPL-3.0) in perpetuity.
     - **No Proprietary Wrapper of Community Code**: Jad Madi will never retroactively relicense existing community-contributed code to non-OSI terms without contributor consent.
     - **Permitted Commercial Scope**: Explicitly defines what enterprise commercial licensing covers: closed enterprise SaaS fleet hosting, proprietary closed-source IDE plugin embeddings, enterprise multi-seat compliance exemptions, and dedicated SLAs.

3. **Fair-Use Trademark & Forking Guidelines**:
   - In accordance with Karl Fogel's *Producing Open Source Software*, the credibility of open-source governance rests on the genuine possibility of forking.
   - Guidelines must explicitly state:
     - Anyone may fork the AGPL-3.0 codebase, modify it, and distribute it under AGPL-3.0.
     - To protect users from confusion, forks must be renamed and must not use the **Thermal™** name, flame logos, or "Don't break the streak" branding.
     - Attribution requirement: Forks may state "Based on Thermal by Jad Madi" or cite upstream commit history without infringing trademarks.

4. **CLI Visibility & Machine-Readable Output**:
   - `thermal license` must output a concise summary of the Community Reciprocity Covenant.
   - `thermal license --json` must include a `"covenant"` field specifying the AGPL protection guarantee.

## File & Interface Contracts

- `DUAL-LICENSE.md`: Add Section on "Community Reciprocity Covenant & Future Scope Boundaries".
- `CONTRIBUTING.md`: Add summary of the Covenant and contributor IP protections.
- `cmd/thermal/main.go`: Update `runLicense()` to include covenant details in plain text and JSON.
- `cmd/thermal/args_test.go`: Test suite verifying `thermal license --json` contains covenant fields.
- `scripts/simulated_user_gate.sh`: Verify license covenant output formatting.

