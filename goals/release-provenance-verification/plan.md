# Plan: Independently verify a released artifact's SLSA provenance

## Execution Steps
- [ ] **Step 1 — Enumerate the published release**: Identify the latest published GitHub release and list its assets (`gh release view --json tagName,assets` or the releases API/plain URLs). Pick one archive for the running platform plus `checksums.txt` and `checksums.txt.intoto.jsonl`.
  Acceptance and verification: asset list recorded with sizes; tag and asset names captured in the evidence log.
- [ ] **Step 2 — Obtain `slsa-verifier`**: `go install github.com/slsa-framework/slsa-verifier/v2/cli/slsa-verifier@latest` (or a release binary); record `slsa-verifier version`.
  Acceptance and verification: `command -v slsa-verifier` succeeds; version string logged. If installation is impossible (offline), stop and record a blocked result.
- [ ] **Step 3 — Verify**: In a temporary directory, run the documented recipe: `slsa-verifier verify-artifact checksums.txt --provenance-path checksums.txt.intoto.jsonl --source-uri github.com/jadmadi/thermal --source-tag <tag>`, then `sha256sum --check --ignore-missing checksums.txt` for the downloaded archive. Capture stdout, stderr, and exit codes verbatim.
  Acceptance and verification: exit codes recorded; on success the verifier prints the verified source/commit; on failure the exact mismatch is recorded.
- [ ] **Step 4 — Record the outcome**: Write the raw log to `docs/evaluations/2026-10-15-release-provenance/` and update `docs/PROMISES.md` section 9: **verified** (artifact + version + commands) or **unverified with disclosed failure**. Correct `docs/DISTRIBUTION.md` only if its commands drift from the real output.
  Acceptance and verification: ledger wording matches the recorded evidence; no overclaiming.
- [ ] **Step 5 — Common final checks and attribution commit**: `env -u NO_COLOR go test -race ./...`, `go vet ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, `./scripts/simulated_user_gate.sh`; commit with a `docs:` or `fix(docs):` subject ending in `(goals/release-provenance-verification/goal.md)`, then `sila goals`. (A docs commit is ignored for closure; include a `chore:`-typed commit only if the ledger outcome is a code-adjacent change.)
  Acceptance and verification: commands exit 0; commit attributed; no push.

## Verification
- Recorded verifier output under `docs/evaluations/2026-10-15-release-provenance/`
- `grep -A 12 "Distribution and build provenance" docs/PROMISES.md`
- Common final checks from `tasks/plan.md`

## Risks
- Network unavailable in this environment → record `blocked`, leave the ledger unverified, and state the missing step.
- The provenance asset may not exist for older releases; pick a release that publishes it, or record which releases lack it as a finding.
