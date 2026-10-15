# Facts: Independently verify a released artifact's SLSA provenance

## Observed Evidence
- `docs/PROMISES.md` section 9 (Distribution and build provenance): outcome **unverified** — "a workflow definition is not a verified artifact. In this audit no released artifact was downloaded and verified with `slsa-verifier`".
- `README.md` publishes the verification recipe (`slsa-verifier verify-artifact checksums.txt --provenance-path checksums.txt.intoto.jsonl --source-uri github.com/jadmadi/thermal --source-tag vX.Y.Z` plus `sha256sum --check --ignore-missing checksums.txt`) and tags yet-unchecked claims to the ledger.
- `.github/workflows/release.yml` runs the OpenSSF generic SLSA3 generator on `v*` tags (workflow-reviewed only).
- `docs/DISTRIBUTION.md` contains the SLSA3 verification runbook; the last released tag in the repository is `v0.20.0` with `.release-please-manifest.json` at `0.20.1` — the exact published asset set must be enumerated from GitHub Releases at run time.
- `slsa-verifier` is not currently installed on this machine (`command -v slsa-verifier` is expected to fail); installation via `go install github.com/slsa-framework/slsa-verifier/v2/cli/slsa-verifier@latest` or the published binary is part of the run.

## Constraints
- Read-only verification against published releases; no publishing, tagging, or workflow edits.
- Record raw stdout/stderr and exit codes; a verifier failure must be recorded as a disclosed failure, never retried into a pass.
- Network dependency is disclosed; an offline environment produces a `blocked` result with evidence, not a fabricated receipt.
- Keep repository hygiene: downloads live under a temporary directory (or `rm/`), never as loose artifacts in the repo root.

## File & Interface Contracts
- `docs/PROMISES.md` section 9 outcome row.
- `docs/evaluations/2026-10-15-release-provenance/` raw log directory (new).
- `docs/DISTRIBUTION.md` runbook (only if the documented commands need correction).

## Verification Contract
- Recorded commands: `slsa-verifier version`; `slsa-verifier verify-artifact checksums.txt --provenance-path checksums.txt.intoto.jsonl --source-uri github.com/jadmadi/thermal --source-tag <tag>`; `sha256sum --check --ignore-missing checksums.txt`.
- Then the common final checks from `tasks/plan.md` (documentation-only change, so the suite should stay green unchanged).
