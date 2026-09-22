# Facts: Version Migration Guide Architecture & docs/MIGRATION.md

## Architectural Invariants & Constraints

1. **Alignment with `version-migration-guide` Specification**:
   - A migration guide is a goal-oriented how-to in the Diátaxis taxonomy. It assumes competence and focuses exclusively on what the reader and automated scripts must now do, in the order they must do it.
   - Under the Churn Rule, Thermal's consumers are external developers and CI/agent scripts that cannot be updated in-place via repository codemods. Therefore, an authoritative, publicly accessible migration guide is mandatory.

2. **Mandatory 6-Part Entry Anatomy**:
   Every breaking change in `docs/MIGRATION.md` MUST conform strictly to this order and structure:
   1. **Verb-prefixed heading with stable anchor**: `### Removed: ...`, `### Changed: ...`, `### Renamed: ...`, `### Deprecated: ...`
   2. **The affected-if line (detection signal)**: A search pattern finding impacted scripts/commands AND the exact runtime error string.
   3. **What changed and why**: Capped at three sentences maximum.
   4. **Before/After snippets**: Explicit, runnable blocks labeled `// Before` and `// After`.
   5. **The fix**: Explicit replacement command or numbered steps.
   6. **The escape hatch**: The compatibility flag, temporary workaround, or an explicit statement: `"None"`.

3. **Retroactive Inventory of Historical Breaking Changes**:
   - **v0.7.0 — Removed `cc` alias**: Dropped ambiguous `cc` alias in favor of `claude` (and `command-code`).
   - **v0.7.0 — Strict Report Option Validation**: Flags (`--since`, `--until`, `--last`, `--order`, `--breakdown`) that previously failed silently now exit with error code 1 when invoked without a report verb (`daily`, `weekly`, `monthly`, etc.).
   - **v0.11.0 — Disjoint Token Mechanics**: Enforced strict token disjointness where nested tokens (cache inside input, reasoning inside output) are subtracted at loader boundary.
   - **v0.12.0 — Workload Replay JSON Schema**: Introduced `verdict`, `headroom_pct`, and `monthly_delta_usd` payload structures into `thermal replay --json`.

4. **Machine-Readable & Agent-Consumable Requirements**:
   - Literal error messages and search regex patterns must be quoted exactly.
   - Code snippets must be complete, copy-pasteable bash invocations.
   - Stable URL anchors per entry so support, agent prompts, and CI checks can link to individual changes.

5. **Governance Gate in `AGENTS.md`**:
   - Prohibit merging breaking CLI arguments, flag semantics, or `--json` schema modifications without a matching entry drafted in `docs/MIGRATION.md`.

## File & Interface Contracts

- `docs/MIGRATION.md`: Canonical user and agent upgrade guide organized by release versions and ranked by blast radius.
- `docs/pages/index.html`: Navigation, footer, and documentation links pointing to the migration guide.
- `README.md`: Upgrade and migration guide reference for users and automation workflows.
- `AGENTS.md`: Governance gate rules enforcing migration documentation before breaking code lands.
