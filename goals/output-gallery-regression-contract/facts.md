# Facts: Turn the output gallery into a reproducible consistency gate

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:2647-2763` labels `thermal daily --since 2026-01-01` as “report flag without report” but it exits 0 and prints a full daily report; the actual no-report command is `thermal --since 2026-01-01`.
- `docs/OUTPUT-GALLERY.md:261` daily total 38.0B vs `:427` monthly 38.1B (capture drift); the header records only one date, not a per-section timestamp/fixture.
- `docs/OUTPUT-GALLERY.md:1354-1366` JSON table shows only top-level keys; it hides the leaderboard aliases and raw float noise documented in the JSON goal.
- `docs/OUTPUT-GALLERY.md` is ~146KB/2,798 lines, yet the TOC only links one variant; width/color/hybrid live frames are absent.
- The generator used to create the gallery lived in `/tmp` and was not committed; there is no reproducible script or gate.

## Interpretation
The gallery is a strong review artifact but not yet a test. Its most damaging flaw is that a reader cannot tell data drift from a product bug. A fixed snapshot plus a consistency matrix turns it into a regression gate; the prerequisite product fixes make the assertions stable.

## Ownership and Constraints
`docs/OUTPUT-GALLERY.md` and the new generator/gate scripts own the artifact; product code ownership remains with the prerequisite goals. Preserve privacy (no new file paths/prompt contents), no network, and no new dependencies beyond existing shell/Go tooling.

## Verification Contract
```bash
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
./scripts/output_gallery_gate.sh
```

## Open Questions
- Which fixture should be canonical: the simulated-gate fixture or a smaller dedicated gallery fixture? The goal prefers the gate fixture to avoid a second data model.
- Should the gallery be checked in as full markdown, a normalized summary plus appendices, or generated on demand? The goal requires a checked-in normalized baseline and a generated full artifact.
