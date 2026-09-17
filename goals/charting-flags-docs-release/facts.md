# Facts: Chart flags, docs, and 0.7.0 release

## Architectural Invariants & Constraints
- `--chart` only adds rows to stdout. JSON output is untouched, so scripts keep working.
- Release follows the repository pipeline: pull request into main, release-please opens the version pull request, merging it tags and runs GoReleaser. UPX stays linux and windows only.
- The branch starts from the latest `origin/main` and gets `main` merged in before review, per AGENTS.md.
- Documentation must state that cost is estimated for several tools and that day granularity limits switching analysis.

## File & Interface Contracts
- `--chart` lands on `projects`, `models`, `daily`, `weekly`, `monthly`.
- AGENTS.md gains: the TUI dependency boundary rule, chart conventions (no color-only encoding, color degradation, width awareness), and the explicit `thermal dashboard` entry rule.
- README gains a charts section with a screencast and the key list, plus the new verbs in the usage block.
