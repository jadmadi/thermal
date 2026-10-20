# Facts: Make statusline and share-card scope explicit and URLs usable

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:1162` statusline: `🔥 91d · 220.0M tok · CMD`; leaderboard top CMD is `12d` at `:159`; no scope labels.
- `docs/OUTPUT-GALLERY.md:1073-1078` share: `Scope: All`, `Active Days: 145 days`, `Total Volume: 38.0B tokens`, `Spend Profile: ~$20596.53`; daily active days are 96 at `:261`.
- `share --json | jq 'keys'` returns `snapshot`, `token`, `url`; `snapshot.cost` and `snapshot.scope` are null while the text card prints values.
- `docs/OUTPUT-GALLERY.md:1087` prints the full encoded URL as one very long line; the README calls the card compact.

## Interpretation
The numbers may each be correct, but the presentation hides their scope. A global streak, today's tokens, and a top-tool label can coexist if labeled; a share card cannot mix all-time active days with a daily-window spend without saying so.

## Ownership and Constraints
`internal/render/statusline.go` owns the prompt; `internal/share` owns the card/encoding; `cmd` wires flags/JSON. Preserve stateless/client-only behavior, zero network, and no file paths/prompts in the encoded payload. No new dependencies.

## Verification Contract
```bash
go test -race ./internal/render ./internal/share ./cmd/thermal -run 'Statusline|Share|Scope|URL'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

## Open Questions
- Is the share card intended to be all-time or the daily window? The goal requires one scope; choose and test it.
- Should the default share output print a shortened URL or only the card with `--url`/`--json` for the full URL? The goal requires a usable default, not a 2KB line.
