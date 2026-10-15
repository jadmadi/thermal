# Facts: Align hosted documentation claims with the promises ledger

## Observed Evidence
Claim inventory (2026-10-15, revision `911144c`, `grep -n "sub-10ms\|10ms\|mmap\|SLSA\|provenance\|zero alloc" docs/pages/index.html docs/pages/llms.txt docs/pages/llms-full.txt`):

- `docs/pages/index.html:809` — "…compiles contribution heatmaps and token metrics across 15 AI coding tools in **sub-10ms cached invocations**." Blanket claim; contradicts the disclosed 42.4ms warm p50 at 1,000-file scale in `docs/PERFORMANCE.md`.
- `docs/pages/index.html:1705` — "`thermal audit` — Local setup & context health: **verify SQLite permissions, mmap, logs, and pricing cache**." Contradicts the AUD-2026 remediation: the audit measures filesystem readability only and explicitly does not check SQLite integrity, mmap, transcript ceilings, or pricing-cache integrity.
- `docs/pages/llms-full.txt:19` — "**Sub-10ms Cached Invocations**: Multi-file and large log directories use incremental caching (`~/.cache/thermal/<tool>.json`) tracking file modification timestamps and sizes." Descries caching directory scanners that do not yet have incremental caches (see the directory-scan-incremental-cache goal) and repeats the blanket latency claim.
- `docs/pages/llms-full.txt:13` — "Multi-gigabyte SQLite databases are scanned … for **sub-second ingestion**." Plausible but unmeasured; should point at `docs/PERFORMANCE.md` or be qualified.
- `docs/pages/index.html:2117` — FAQ answer about read-only ingestion and the only files Thermal writes (its own caches): consistent with the ledger; keep.
- `docs/PROMISES.md` is the canonical outcome table; `docs/PERFORMANCE.md` is the evidence artifact for performance wording.

## Constraints
- Wording only: no product, CLI, or schema changes; page structure preserved.
- Qualify measurements, do not weaken the governance target (sub-10ms cached invocation remains the invariant).
- Keep claim wording consistent with ledger outcomes: receipts = observed command outcomes; replay = scenario comparison with coverage labels; audit = readability-only; provenance = workflow-defined and unverified until an artifact is verified.
- `docs/pages` is a deployable static site (see the gate's install.sh dry-run and llms presence checks); avoid breaking those checks.

## File & Interface Contracts
- `docs/pages/index.html`: performance paragraph (~line 809) and the audit command row (~line 1705).
- `docs/pages/llms-full.txt`: performance bullet(s) (~lines 13, 19).
- `docs/pages/llms.txt`: check for the same claims during the inventory.
- `docs/PROMISES.md`: record the hosted-page alignment in the evidence base once done.

## Verification Contract
- Inventory command: `grep -n "sub-10ms\|mmap\|pricing cache\|incremental caching" docs/pages/index.html docs/pages/llms*.txt` reviewed line by line against `docs/PROMISES.md`.
- Existing gate checks (`docs/pages/install.sh` dry-run; llms.txt/llms-full.txt presence) still pass.
- Common final checks from `tasks/plan.md`.
