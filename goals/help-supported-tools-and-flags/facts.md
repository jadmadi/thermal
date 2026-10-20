# Facts: Make help's tool registry and flag reference exhaustive and accurate

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:125-128` lists `opencode · devin · codex · claude · zcode · grok · mimo · codewhale · dsh · hermes · cmd` and `agy · muse · droid`; Zed is missing.
- `docs/OUTPUT-GALLERY.md:159` shows a `Zed` leaderboard row; project metadata claims 15 tools.
- `docs/OUTPUT-GALLERY.md:99` shows `--start-of-week` without a value placeholder; the flag is registered as a string in `cmd/thermal/main.go`.
- `docs/OUTPUT-GALLERY.md:102` shows `--sort` only as `streak|tokens|cost`, while `validateReportFlags` accepts `tokens|cost|days|recent` for projects and `tokens|cost` for models.
- `internal/render/help.go:252` is a hard-coded tool list; `internal/render/help_test.go` only checks a subset of commands and does not assert the tool count.

## Interpretation
The help is a hand-maintained copy of product state. The missing Zed is not cosmetic: a user cannot discover a supported tool from help. The flag gaps make valid commands look invalid or hide compatibility aliases.

## Ownership and Constraints
`internal/render/help.go` owns prose/layout; the loader registry owns tool truth; `help_test.go` owns the contract. Preserve the existing visual sections and `WrapText` behavior. No new dependencies.

## Verification Contract
```bash
go test -race ./internal/render -run 'Help|Tool|Flag'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
/tmp/thermal-test --help | grep -E 'zed|--start-of-week|--sort'
./scripts/simulated_user_gate.sh
```

## Open Questions
- Should the help list display names or CLI aliases first? The goal requires both to be traceable; pick one ordering and test it.
- Is `--format` a supported compatibility alias or a retired flag? The goal accepts either, but requires one documented outcome.
