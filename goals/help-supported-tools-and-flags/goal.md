# Goal: Make help's tool registry and flag reference exhaustive and accurate

## Goal Description
Priority: P1 (discoverability). `thermal --help` lists 11 Token Warriors and 3 Activity Hunters but omits Zed, which appears in the leaderboard; the project claims 15 tools. The same help says `--sort streak|tokens|cost`, omitting projects' `days|recent`; hides `--format`; and shows `--start-of-week` without the required `<day>` value. It also needs to state which commands honor `--chart` after the chart/no-op fix.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 203
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/render/help.go`
- `internal/render/help_test.go`
- `internal/loaders/registry.go`
- `internal/thermal/types.go`
- `cmd/thermal/main.go`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Registry-driven tools**: Replace the hard-coded Token Warriors/Activity Hunters lists with a single registry-derived list (or add a test that asserts the static list equals the registry). Include all 15 supported tools: Devin, OpenCode, MiMoCode, Codex, Claude, Antigravity/Agy, command-code/CMD, codewhale, Droid, DeepSeek (DSH), Grok, Muse, Nous Hermes, ZCode, Zed. The help must show Zed.
2. **Display names/aliases**: Render each tool as a stable display name plus CLI alias where they differ (`MiMoCode (mimo)`, `CMD (command-code)`, `DeepSeek DSH (dsh)`, `Nous Hermes (hermes)`), or add a mapping line. The leaderboard's display names and help's CLI names must be traceable without guessing.
3. **Sort reference**: Replace the single `--sort <field>` line with per-command applicability: leaderboard `streak|tokens|cost`; projects `tokens|cost|days|recent`; models `tokens|cost`; and any analytics command that accepts `--sort` with a documented value set. Add a test that every allowed value in `validateReportFlags` appears in help and no retired value appears.
4. **Format reference**: Either document `--format text|json` (including that `json` is an alias for `--json`) or remove it with a migration entry. Do not leave a supported flag hidden. If `--format md/markdown` is retired, keep the retirement error and mention it in the migration.
5. **Value placeholders**: Fix `--start-of-week` to `--start-of-week <day>`; audit all value-taking flags for a missing `<...>` placeholder (`--tool`, `--db`, `--weeks`, `--since`, `--until`, `--last`, `--order`, `--sort`, `--top`, `--metric`, `--by`, `--grain`, `--interval`, `--flame-ceiling`).
6. **Chart applicability**: After the chart goal decides whether mix/stats/trend honor `--chart`, reflect the final truth in help. If `--chart` is rejected there, show the allowed commands; if implemented, document the chart type. No help text may promise a chart that does not appear.
7. **Golden help test**: Add a test that asserts the exact supported-tool count, all 15 names/aliases, every registered flag name, required value placeholders, and the absence of retired words (`dashboard`, `web`, `serve`, `replay`, `receipt`, `yield`, `9-box`). Test at width 80 and 120 for wrapping.
8. **Docs**: Update README's command/flag reference and `project.jsonc` feature/tool list. Add a migration entry if `--format` or any flag semantics change.

## Done Condition
1. `thermal --help` lists Zed and all 15 tools; the count and names match the loader registry/project metadata; a test fails if a new tool is added without help.
2. `--sort` shows the correct values for leaderboard, projects, and models; `--format` is documented or removed with a migration; `--start-of-week <day>` and every value-taking flag have a placeholder.
3. Help reflects the final `--chart` applicability and contains no retired command words.
4. Help golden tests pass at 80/120 columns; full race suite, build, and simulated gate pass; README/project.jsonc are synchronized; the attribution commit ends in `(goals/help-supported-tools-and-flags/goal.md)`.

## Scope Boundary
Help/README/metadata truth only. Do not change command behavior except where docs reveal a value placeholder; do not implement chart behavior (chart goal) or JSON changes. Preserve help width/wrapping and `--no-color` output.
