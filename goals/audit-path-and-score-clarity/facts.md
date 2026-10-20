# Facts: Make audit findings addressable and score understandable

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:1109-1118` lists `WARN CLAUDE.md 204 lines`, `WARN AGENTS.md 204 lines`, `WARN GEMINI.md 204 lines`, `PASS CLAUDE.md 1 lines`, `CRIT AGENTS.md 494 lines` with no paths.
- `docs/OUTPUT-GALLERY.md:1138-1140` `PASS Tool Databases` is spaced differently from `PASS  Instructions` and wraps without the same indent.
- `docs/OUTPUT-GALLERY.md:1094` `Score: [58/100 · Grade D]` with no component explanation; advice at `:1149-1157` names files but not paths.
- `internal/audit/audit.go` owns findings; `internal/render/audit.go` owns the table/advice; tests currently do not assert duplicate-basename paths or long-name width.

## Interpretation
The audit is measuring real context tax, but the output cannot be acted on when two files share a basename. Adding path/scope is a display/schema change, not a new check. The score needs a transparent explanation to be trustworthy.

## Ownership and Constraints
`internal/audit` owns finding data; `internal/render/audit.go` owns layout; tests own duplicate-name/long-name fixtures. Preserve read-only checks, stderr diagnostics, and no file-content leakage. No new dependencies.

## Verification Contract
```bash
go test -race ./internal/audit ./internal/render -run 'Audit|Path|Score'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

## Open Questions
- Should path be absolute, `~`-abbreviated, or repo-relative? The goal prefers `~`-abbreviated with root/scope labels for readability.
- Should the score formula change or only be explained? The goal defaults to explaining the existing formula; a formula change requires migration.
