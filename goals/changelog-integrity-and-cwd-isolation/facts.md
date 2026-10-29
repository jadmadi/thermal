# Facts: Sanitize changelog history, isolate embedded changelog from CWD hijacking, and strip Sila goal tokens

## Observed Evidence
- `CHANGELOG.md:8-173` contains two `## [0.21.0]` sections: lines 8-159 were prepended by `release-please` (commit `03df5ca`) scanning from the root commit because `.release-please-manifest.json` referenced `"0.20.1"` which had no git tag, while lines 161-173 contain the curated `0.21.0` notes.
- `CHANGELOG.md` contains raw `(goals/<slug>/goal.md)` and `[goal: <slug>]` strings across multiple historical releases.
- `internal/changelog/changelog.go:140-153` (`LoadContent`) reads `CHANGELOG.md`, `../CHANGELOG.md`, and `../../CHANGELOG.md` from the current working directory before falling back to `embeddedChangelog`, causing `thermal changelog` to read arbitrary third-party changelogs when invoked in other repositories.
- `internal/changelog/changelog.go:39` (`Parse`) strips markdown commit links via `commitRegex`, but does not strip `(goals/<slug>/goal.md)` or `[goal: <slug>]` tokens.

## Architectural Invariants & Constraints
- `thermal changelog` must always render Thermal's embedded `CHANGELOG.md` regardless of the user's current working directory.
- `CHANGELOG.md` and `internal/changelog/CHANGELOG.md` must stay byte-for-byte identical and free of internal Sila goal scaffolding tokens.
- `Parse()` must strip Sila goal tokens and ignore duplicate release version sections.

## File & Interface Contracts
- `internal/changelog/changelog.go`: `LoadContent() string`, `Parse(content string) ([]Release, error)`
- `CHANGELOG.md` and `internal/changelog/CHANGELOG.md`
- `scripts/simulated_user_gate.sh`

## Verification Contract
```bash
CGO_ENABLED=0 go test -race ./internal/changelog/...
./scripts/simulated_user_gate.sh
```
