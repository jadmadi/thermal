# Goal: Make version output and deprecation warnings factually true

## Goal Description
Priority: P1 (trust). The help page promises "SemVer, commit hash, and build
timestamp"; the source build prints `thermal dev (built from source)` and JSON
reports `"commit":"unknown","date":"unknown"` without saying the build is
unstamped. Separately, `--license` and the `nous` alias warn that they "will be
removed in v0.14.0" while the changelog's newest release is v0.21.0. The
warning's remediation is a repo-relative `docs/MIGRATION.md#...` path that does
not exist in a user's HOME.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`):

- `thermal version` -> exit `0`, stdout `thermal dev (built from source)`.
- `thermal version --json` -> exit `0`:
  `{"commit":"unknown","date":"unknown","version":"dev (built from source)"}`.
- `thermal --help` -> `version   Print SemVer, commit hash, and build
  timestamp`.
- `thermal -v` -> exit `0`, normal leaderboard, stderr empty.
- `thermal --license` -> exit `0`; stderr:
  `thermal: warning: --license is deprecated and will be removed in
  v0.14.0; use thermal license (see docs/MIGRATION.md#deprecated-license-flag)`.
- `thermal nous` -> exit `1`; the same v0.14.0 warning, then
  `database not found`.
- `thermal changelog --top 5` -> newest release `v0.21.0`.

The outcome of this goal is a **truthful metadata and warning contract**:

1. Version output states whether the build is stamped and never implies a
   SemVer/commit/date that is absent.
2. The help text matches the fallback.
3. `-v`/`--verbose` announces itself once on stderr, so it is observably
   different from a normal run.
4. Deprecation warnings never name an expired removal version; they name the
   replacement, a runnable command, and a resolvable URL.
5. A test rejects any warning whose removal version is <= the current version.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: cli-version-verbose-flag-contract, cli-staged-deprecation-and-upgrade
- **Sequence**: 261
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3
Related goals:
- `cli-version-verbose-flag-contract` owns the `-v`/`-V`/`--version` meanings;
  this goal owns the payload and the verbose signal.
- `cli-staged-deprecation-and-upgrade` owns the warning infrastructure; this
  goal owns the truthfulness of its text and deadlines.
- `version-migration-guide` owns the migration document; this goal ensures the
  warning points at a resolvable location.

## Files to Touch
- `internal/version/version.go`
- `cmd/thermal/version.go`
- `cmd/thermal/main.go`
- `cmd/thermal/args.go`
- `cmd/thermal/help.go`
- `cmd/thermal/version_test.go`
- `cmd/thermal/args_test.go`
- `cmd/thermal/upgrade_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `docs/PROMISES.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/cli-version-and-deprecation-truth/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/cli-version-and-deprecation-truth/plan.md`](plan.md)
- Prior contracts:
  [`goals/cli-version-verbose-flag-contract/goal.md`](../cli-version-verbose-flag-contract/goal.md),
  [`goals/cli-staged-deprecation-and-upgrade/goal.md`](../cli-staged-deprecation-and-upgrade/goal.md)

## Requirements

1. **Version payload.**
   - Text, unstamped source build:
     ```
     thermal dev (built from source; version not stamped)
       commit: unknown · built: unknown
     ```
   - Text, stamped build:
     ```
     thermal v0.21.0
       commit: 5f652d4
       built:  2026-10-29T10:19:25Z
     ```
   - Both lines are always present; do not silently omit `commit`/`built` when
     unknown. The word `unknown` is the honest value.
   - JSON:
     ```json
     {
       "type": "version",
       "version": "dev (built from source; version not stamped)",
       "stamped": false,
       "commit": "unknown",
       "date": "unknown",
       "semver": ""
     }
     ```
     Stamped builds set `stamped:true`, `semver:"v0.21.0"`, and the real
     commit/date.
   - `-V`, `--version`, and `version` produce the same payload (apart from a
     clock field, if any); prefer no `generatedAt` to keep it deterministic.
   - Version performs no data scan, network call, pricing load, or update
     check.

2. **Help truth.**
   - `thermal version --help` summary becomes:
     `Build metadata, including the source-build fallback`.
   - The description states:
     `When the binary is built without ldflags, version is "dev", commit and
     date are "unknown", and stamped is false.`
   - Remove the unconditional "SemVer, commit hash, and build timestamp"
     promise.
   - The global help line stays short: `version   Print build metadata`.
   - `-v, --verbose` remains distinct from `-V, --version`; the help note
     stays.

3. **Verbose signal.**
   - When `-v`/`--verbose` is active, write exactly one line to stderr before
     the command output:
     `thermal: verbose diagnostics enabled (stderr only)`.
   - If there are loader warnings, they follow on stderr; the announcement is
     not repeated.
   - `--json --verbose` keeps stdout valid; the announcement is stderr only.
   - `-v` with no data still prints the announcement, then the no-data
     contract.
   - Do not print the announcement when verbose is off.
   - `--verbose=false` produces no announcement.

4. **Deprecation table.**
   - One table is the source of truth:
     ```go
     type Deprecation struct {
         Item           string // "--license", "nous"
         Replacement    string // "thermal license", "hermes"
         RemovalVersion string // "" means "a future release"
         Anchor         string // URL fragment, not a repo path
     }

     var deprecations = []Deprecation{
         {"--license", "thermal license", "", "deprecated-license-flag"},
         {"nous", "hermes", "", "deprecated-tool-nous"},
     }
     ```
   - Warnings:
     - If `RemovalVersion` is empty:
       `thermal: warning: --license is deprecated; use thermal license`
       `  details: https://jadmadi.net/project/thermal/migration#deprecated-license-flag`
     - If set and still in the future:
       `... will be removed in vX.Y.Z; use ...`
     - If set and <= the current version: the test fails; the build must not
       ship. At runtime, fall back to the no-version form.
   - The URL must be absolute and resolvable; never `docs/MIGRATION.md`.
   - The replacement must be a runnable command (`thermal license`) or a
     `--tool` value (`hermes`), not prose.
   - The warning is one or two lines; no stack trace, no color on a non-TTY.

5. **Current-version comparison.**
   - Add `version.Current()` returning the SemVer or `dev`.
   - Add `version.IsPast(target string) bool` using the existing SemVer
     comparator, treating `dev` as "unknown, do not claim expired".
   - A table test iterates every deprecation and asserts:
     - `RemovalVersion == ""` or `!IsPast(RemovalVersion)`;
     - `Replacement` is non-empty and runnable;
     - `Anchor` starts with `https://` and contains `#`.
   - A second test asserts no source file contains the literal
     `v0.14.0` in a deprecation string.
   - A third test asserts the runtime fallback when `IsPast` is true.

6. **Warning channels.**
   - Warnings go to stderr; stdout is unchanged.
   - `--json` stdout stays valid; the warning is stderr.
   - `--no-color`/`NO_COLOR`/non-TTY produce no ANSI.
   - `--quiet` does not exist; do not add one for warnings.
   - A warning is emitted once per invocation, even if the flag appears twice.

7. **Tests.**
   - Version text/JSON matrix for stamped and unstamped builds, using
     ldflags-like injection in a test build or a package variable override.
   - Help string test: no unconditional SemVer promise.
   - `-v` announcement test: exactly one stderr line; `-v --json` stdout valid.
   - Deprecation table test: no expired version, absolute URL, runnable
     replacement.
   - Warning text test: no `docs/MIGRATION.md`, no raw version when unknown.
   - `--license` and `nous` end-to-end tests.
   - Gate: `simulated_user_gate.sh` covers `version`, `version --json`,
     `-v`, and `--license`.

8. **Documentation.**
   - README "Check your current version" shows both stamped and unstamped
     output.
   - MIGRATION entry for the version payload change and the warning text
     change; detection commands.
   - `docs/PROMISES.md` updates any version claim.
   - `project.jsonc` updated if the version feature text changes.

## Non-Goals
- Stamping source builds automatically (no `git describe` at runtime).
- Adding a `--check` to `version` (that is `upgrade --check`).
- Removing `--license` or `nous`; this goal only makes the warning true.
- Changing the `-v`/`-V` meanings (owned by the prior goal).
- Adding a network call to resolve the latest release.

## Acceptance Criteria
- AC-001: An unstamped `thermal version` exits `0` and prints
  `version not stamped`, `commit: unknown`, and `built: unknown`.
- AC-002: `thermal version --json | jq -e '.stamped==false and
  .commit=="unknown" and .date=="unknown"'` exits `0`.
- AC-003: A stamped test build prints `vX.Y.Z` and its real commit/date, with
  `stamped:true`.
- AC-004: `thermal version --help` contains the source-build fallback sentence
  and not the unconditional SemVer promise.
- AC-005: `thermal -v` exits `0` and writes exactly one
  `verbose diagnostics enabled` line to stderr.
- AC-006: `thermal -v --json | jq -e .` exits `0`; the announcement is stderr
  only.
- AC-007: `thermal --license` warning contains no `v0.14.0` and no
  `docs/MIGRATION.md`; it contains `thermal license` and an `https://` URL.
- AC-008: `thermal nous` warning contains no expired version and names
  `hermes`.
- AC-009: The deprecation table test rejects any `RemovalVersion` that is past.
- AC-010: A grep test finds no deprecation string with `v0.14.0`.
- AC-011: `go test -race ./cmd/thermal/... ./internal/version/...` exits `0`,
  including `TestVersionPayloadMatrix`, `TestVerboseAnnouncement`,
  `TestDeprecationTruth`, and `TestNoExpiredWarningVersion`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the
  version and license cases.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| Unstamped build | honest fallback, `stamped:false` |
| Stamped build | real version/commit/date |
| `-v` | one stderr announcement |
| `-v --json` | valid stdout, announcement stderr |
| `--verbose=false` | no announcement |
| `--license` | no expired version, runnable replacement |
| `nous` | no expired version, `hermes` |
| Past removal version | runtime no-version form; test fails |
| Repo-relative anchor | test fails |
| Non-TTY warning | zero ANSI |
| Repeated flag | one warning |

## Metrics & Regression Guards
- Version truth: every field is present; `stamped` matches the ldflags state.
- Help truth: no unconditional metadata promise; a grep test enforces it.
- Verbose signal: exactly one line per invocation.
- Deprecation truth: zero expired versions, zero repo-relative URLs, 100%
  runnable replacements.
- Gate: simulated user gate asserts AC-001, AC-005, and AC-007.

## Risks & Mitigations
- **Risk:** Printing `commit: unknown` looks worse than omitting it.
  **Mitigation:** honesty is the point; a stamped release shows real values.
- **Risk:** The `-v` announcement pollutes stderr in scripts.
  **Mitigation:** the caller asked for verbose; it is one line and stderr-only.
- **Risk:** Removing the version from warnings weakens the deprecation signal.
  **Mitigation:** keep the future version only when it is true; otherwise name
  the replacement and the migration URL.
- **Risk:** The migration URL may change.
  **Mitigation:** one constant and a test that it starts with `https://`; a
  link checker in CI if available.
- **Risk:** `dev` cannot be compared to a SemVer.
  **Mitigation:** treat `dev` as unknown and never claim a deadline has passed
  at runtime; the table test still rejects a past target for release builds.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "deprecated will be removed in v0.14.0"` and
  `sila recall "version dev built from source"`.
- **Writes**: record a `case` for the "deprecation truth table with current
  version guard" mechanism; record a `lesson` for the stale deadline.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "version and deprecation truth shipped"
  --open="release stamping verification"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/cli-version-and-deprecation-truth/goal.md)` under a `fix(cli):` or
   `feat(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the version matrix, verbose signal, and deprecation table evidence.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
