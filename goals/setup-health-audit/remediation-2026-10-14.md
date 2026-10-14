# Audit claim remediation AUD-2026

## Context and files

P2. Ready, no prerequisites. Cwd `/mnt/Jad/github/lab/thermal-streak`.
Why: a readable file is not a validated database or proof that mmap works.
Read `internal/audit/audit.go:checkToolDatabaseAccess`, `auditAgentReadiness`, the audit probe,
`README.md` audit paragraph and `tasks/plan.md`.
Edit `internal/audit/audit.go`, `internal/audit/audit_test.go`, `README.md`, and `project.jsonc`.
Migration documentation is a permitted companion if status semantics change.

## Fixed scope

Design choice: make the existing lightweight checks honest rather than add costly integrity scans.
Keep `RunAuditWithOptions(workDir, homeDir string) AuditReport` and JSON field names.
The Tool Databases finding measures filesystem readability only. Its message must say exactly
that and explicitly say SQLite integrity and mmap were not checked. Readability can PASS only
under that narrow label; no `verified database` or `mmap-ready` wording. Similarly replace
`sub-10ms ready` for writable cache with a factual writable-cache statement.
If MkdirAll fails, return a visible WARN instead of omitting the cache finding. Retain cleanup of
probe files and clarify documentation: audit writes a temporary probe only to Thermal's cache,
never to agent stores. Do not advertise pricing-cache integrity or transcript-ceiling checks
unless there is actual code/evidence for them; remove those unsupported public claims.
Preserve rule/MCP heuristic behavior and label estimates as estimates, not observed usage.

- AUD-201: add `TestAuditDatabaseReadabilityClaim`: plain text at state_5.sqlite can be readable,
  but output never claims SQLite/mmap validation; missing and unreadable paths disclose limits.
  Update the historical audit probe to the new narrow expectation; preserve the original audit log.
- AUD-202: add `TestAuditCacheFailureVisible`: a file at the expected cache-directory location
  causes visible WARN; success says writable only. Source fixture bytes unchanged after audit.
- AUD-203: `go test -race ./internal/audit ./internal/render -run Audit` exits 0; common checks pass.
  Inspect README and project metadata for matching capability descriptions and updated lastEdit.

No new SQLite scans, registry architecture changes, dependencies, or claims of comprehensive health.
Stop/return/review contract: `tasks/plan.md`.
