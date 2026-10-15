# Goal: Align hosted documentation claims with the promises ledger

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Session: `mcp-audit-goal` | Report: `receipt_docs-page-claim-parity_1792089651`


## Goal Description
The hosted documentation page still carries the blanket promise:

> …computing contribution heatmaps and token metrics across 15 AI coding tools in sub-10ms cached invocations. (`docs/pages/index.html`)

After the performance remediation, the README and [`docs/PROMISES.md`](../../docs/PROMISES.md) scope that claim: warm p50 7.8ms on the baseline fixture (PASS) and 42.4ms at 1,000-file directory scale (disclosed MISSED). The hosted page must match the ledger, link the evidence, and stop making a blanket claim the measurements do not support.

Priority: **P2**. Follow-up from the ingestion-performance-baseline remediation and the promise-claim-evidence ledger.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 110
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `docs/pages/index.html`
- `docs/pages/llms.txt` / `docs/pages/llms-full.txt` (if the same claim appears)
- `docs/PROMISES.md` (record the hosted-page evidence once aligned)
- `docs/PERFORMANCE.md` (link target already exists)

## References
- [Evidence and constraints](facts.md)
- [Execution plan](plan.md)

## Done Condition
1. A claim inventory over `docs/pages/` finds no blanket performance claim that contradicts `docs/PROMISES.md`; the sub-10ms wording is scoped to the measured baseline with a link to `docs/PERFORMANCE.md`.
2. Receipt/replay/audit/provenance wording on the hosted page matches the ledger outcomes (observed outcomes; scenario comparison; readability-only audit; workflow-defined provenance).
3. No product, CLI, or schema change; the page keeps its structure and links.
4. Verification: claim inventory command exits clean; existing gate checks (`docs/pages/install.sh` dry-run, llms.txt presence) still pass; implementation-attributed commit before closure.

## Scope Boundary
Documentation wording only. Do not redesign the page, add new claims, or touch generated assets beyond the claim text. Do not weaken the governance target statement — qualify the measurement, not the invariant.
