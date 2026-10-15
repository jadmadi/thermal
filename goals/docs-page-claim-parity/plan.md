# Plan: Align hosted documentation claims with the promises ledger

## Execution Steps
- [ ] **Step 1 — Full claim inventory**: Run the inventory greps over `docs/pages/` (`sub-10ms`, `10ms`, `mmap`, `pricing cache`, `incremental caching`, `SLSA`, `provenance`, `zero alloc`, `verified`) and list every claim with its line and the ledger outcome it must match.
  Acceptance and verification: inventory list attached to the goal evidence; each line marked keep / qualify / remove-claim.
- [ ] **Step 2 — Qualify the performance wording**: Scope the sub-10ms sentence on `index.html` to the measured baseline and link `docs/PERFORMANCE.md`; fix the `llms-full.txt` incremental-caching bullet so it does not promise a mechanism that does not exist yet (point at the measured evidence instead), and qualify the sub-second ingestion bullet.
  Acceptance and verification: wording matches `docs/PROMISES.md` section 7 (performance) and 10 (allocations); links resolve.
- [ ] **Step 3 — Align the audit and other capability wording**: Update the audit row to readability/writability diagnostics only (no SQLite integrity/mmap/pricing-cache claims).
  Acceptance and verification: wording matches `docs/PROMISES.md` section 8; no remaining `mmap` verification claim in the audit row.
- [ ] **Step 4 — Record and verify**: Add a note to `docs/PROMISES.md` that the hosted page is aligned (revision-bound); run the inventory grep again; run the gate checks that cover `docs/pages` (`install.sh` dry-run, llms presence).
  Acceptance and verification: second inventory review is clean; gate exits 0.
- [ ] **Step 5 — Common final checks and attribution commit**: `env -u NO_COLOR go test -race ./...`, `go vet ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, `./scripts/simulated_user_gate.sh`; commit with a `fix(docs):` or `docs:` subject ending in `(goals/docs-page-claim-parity/goal.md)`, then `sila goals`.
  Acceptance and verification: four commands exit 0; commit attributed.

## Verification
- `grep -n "sub-10ms\|mmap\|pricing cache\|incremental caching" docs/pages/index.html docs/pages/llms*.txt`
- `./scripts/simulated_user_gate.sh` (covers `docs/pages/install.sh` dry-run and llms presence)
- Common final checks from `tasks/plan.md`
