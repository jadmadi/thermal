# Goal: Make statusline and share-card scope explicit and URLs usable

## Goal Description
Priority: P2 (scope/clarity). The statusline prints `🔥 91d · 220.0M tok · CMD` without saying whether 91d is global, 220.0M is today, or CMD is the top tool; the leaderboard shows CMD at 12d. The share card says `Scope: All`, `Active Days: 145`, `Total Volume: 38.0B`, and `Spend: ~$20596.53`; daily says 96 active days. The share JSON has `cost:null`/`scope:null`, and the share URL is roughly 2KB, despite being described as compact.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: output-formatting-primitives, output-report-totals-reconciliation, json-output-contract-hardening
- **Sequence**: 208
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/render/statusline.go`
- `internal/render/statusline_test.go`
- `cmd/thermal/main.go`
- `internal/share/share.go`
- `internal/share/share_test.go`
- `internal/thermal/types.go`
- `docs/MIGRATION.md`
- `README.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Statusline labels**: Make the scope explicit: e.g. `🔥 91d all tools · 220.0M tok today · CMD` or `🔥 91d · 220.0M tok today · top: CMD`. Preserve compactness and `--plain`/`--nerd` variants. Add tests for global vs per-tool and today vs all-time.
2. **Share scope**: Add `scope`/`window`/`windowLabel` to the share snapshot. If `Scope: All` means all-time, activeDays/volume/spend must all be all-time; if it means the daily report window, activeDays must match daily (96) and the label must say the window. Mixed-scope fields are forbidden. The text and JSON must derive from the same snapshot.
3. **Share JSON**: Include `scope`, `window`, `activeDays`, `recordedCost`, `estimatedCost`, `totalCost`, `totalTokens`, and `url`/`token`; no displayed field may be null. Add a test that decodes `share --json` and asserts every text card value is present with the same scope.
4. **URL usability**: Define a compact display policy. Options: show a shortened URL plus the full URL only in `--json`/a separate `--url-only`/`--full` flag, or truncate the middle with an explicit `…` and a note. Never print a 2KB line as the default “compact” card without a warning. Document the URL length/limits and keep the full token recoverable from JSON.
5. **Statusline/share parity**: The share streak (`91d`) and statusline streak must use the same global/per-tool definition, or each must label its scope. The share active days and total volume/spend must be a single snapshot; no field may be sourced from a different window without a label.
6. **Responsive/plain**: Test statusline at typical widths and with `--plain`/`--nerd`; share text at 80/120; no wrapping/truncation of the values. Preserve no-color output.
7. **Docs/migration**: Add a six-part migration entry for the share JSON fields/URL display and any statusline wording change; update README/project.jsonc.

## Done Condition
1. Statusline names the streak scope and token window; the share card names its scope/window; a reader can tell why 91d vs 12d and 145 vs 96 differ (or the values are reconciled).
2. `share --json` exposes scope/window/activeDays/totalTokens/recorded/estimated/totalCost/url; no displayed field is null; text and JSON agree.
3. The default share output no longer prints an unqualified 2KB URL as “compact”; the full URL remains available via JSON or an explicit flag.
4. Statusline/share tests pass at 80/120 and plain/nerd; full race suite, build, and simulated gate pass; docs/MIGRATION/README/project.jsonc updated; the attribution commit ends in `(goals/statusline-and-share-scope-contract/goal.md)`.

## Scope Boundary
Statusline wording and share-card scope/JSON/URL display only. Do not change streak calculation, daily totals, cost estimation, or the share encoding algorithm; consume the reconciliation/JSON contracts. Preserve client-only/stateless privacy guarantees.
