# Fact Sheet: Complete Sila Onboarding

## 1. Primary Purpose & Context
This package was scaffolded by `sila onboard` for thermal-streak. Its job is to finish connecting the repository to sila: map legacy task systems, record the knowledge they carried, and verify the wiring so every future session works through sila first.

## 2. Verification
- `sila onboard check --strict` is the acceptance test.
- Critical gaps: project not registered, duplicate registrations.
- Advisory gaps: missing handoff, memo, lesson, git hook, MCP registration, or an unmapped legacy surface.

## 3. Alternatives Considered
- Bulk-import the legacy tasks automatically. Rejected: old lists mix done, stale, and scratch items; a reviewed conversion avoids recreating the duplication this onboarding exists to end.
