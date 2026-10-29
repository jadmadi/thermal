# Goal: Give -v, -V, --version, and --verbose one unambiguous contract

## Goal Description
Priority: P1 (discoverability). `-v` is context-dependent: it prints the version when it is the only argument, and means verbose everywhere else. The help page documents `-v` as verbose and does not mention `--version` or `-V` at all, so the one place a user checks contradicts the behavior they observe.

Observed on the current source build (2026-10-29):

- `thermal -v` → `thermal dev (built from source)`, exit `0`.
- `thermal --version` → the same, exit `0`.
- `thermal version` → the same, exit `0`.
- `thermal -v daily` → the daily report, no version, exit `0`.
- `thermal -v --json` → leaderboard JSON, no version, exit `0`.
- `thermal -V` → `flag provided but not defined: -V`, then the full 84-line manual, exit `2`.
- `thermal --help` → `--verbose, -v   Output diagnostic warnings to stderr`; no `--version`, no `-V`.
- `thermal --help` also lists the `version` command, so the intended version spelling exists — it is just not a flag.

The outcome of this goal is **one flag contract**: `-V`/`--version` prints version metadata, `-v`/`--verbose` enables diagnostics, `thermal version` remains a command, the help documents all four spellings, and no invocation changes meaning based on argument count.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 254
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 2 (coordinate precedence with `cli-help-interception-and-scoped-help`; that goal owns help > version > validation > execution)
- **Related goals**: `cli-help-interception-and-scoped-help` owns the global precedence table and the help pages that must list the final flags. `cli-argument-parser-strictness` owns unknown-flag rejection; `-V` must become a known flag rather than an error.

## Files to Touch
- `cmd/thermal/main.go`
- `cmd/thermal/args.go`
- `cmd/thermal/help.go`
- `internal/render/help.go`
- `cmd/thermal/version_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-version-verbose-flag-contract/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-version-verbose-flag-contract/plan.md`](plan.md)

## Requirements

1. **One meaning per spelling.**
   - `-V` and `--version` print version metadata and exit `0`.
   - `-v` and `--verbose` enable verbose diagnostics and do not print version metadata.
   - `thermal version` prints the same metadata as `--version`.
   - No invocation may change the meaning of `-v` based on argument count or position.
   - `-V` must be accepted in any position; it is a known flag, not a parse error.
   - `-v` must remain accepted in any position and must not be consumed as a value for another flag.

2. **Precedence.**
   - Extend the global precedence from `cli-help-interception-and-scoped-help`: `help` > `version` > argument validation > command execution.
   - `thermal -V daily` prints version and exits `0`; the command is not executed.
   - `thermal daily -V` prints version and exits `0`.
   - `thermal -v -V` prints version; version wins over verbose.
   - `thermal -V --bogus` prints version and exits `0` (version > validation), documented in help.
   - `thermal --version --json` and `thermal version --json` emit the same JSON object.
   - `thermal -v --json` is the normal JSON output for the resolved command, with diagnostics on stderr; it is not a version request.

3. **Version payload.**
   - Text form:
     ```
     thermal dev (built from source)
       commit: 5f652d4
       built:  2026-10-29T10:19:25Z
     ```
     (Omit `commit`/`built` only when unknown, as today.)
   - JSON form, exactly one object:
     ```json
     {"type":"version","version":"dev (built from source)","commit":"5f652d4","date":"2026-10-29T10:19:25Z","generatedAt":"2026-10-29T19:00:00Z"}
     ```
   - `-V --json`, `--version --json`, and `version --json` must be byte-identical apart from `generatedAt` if it is included; prefer omitting `generatedAt` from version to keep it deterministic, or test with a fixed clock.
   - Exit `0`; stdout only; no stderr; no data scan, no network, no update check.
   - Version output must not be affected by `--verbose`, `--no-color`, `--width`, or `COLUMNS` (except normal wrapping if a line exceeds the width).

4. **Help and discoverability.**
   - Global help lists both:
     ```
     -V, --version   Print version, commit, and build timestamp
     -v, --verbose   Output diagnostic warnings to stderr
     ```
   - The `version` command page shows the same payload and the `--json` form.
   - A precedence note: `Help and version requests win over validation and execution.`
   - The help must not say `-v` is version.
   - `thermal help version` works via the help goal.

5. **Compatibility and migration.**
   - `thermal -v` changes from version to verbose (leaderboard). This is a behavior change and must be documented.
   - MIGRATION six-part entry: what breaks (`-v` no longer prints version), who it hits (scripts and aliases), detection (`thermal -v | head -1`), fix (`thermal -V` or `thermal --version`), escape hatch (`thermal version`), and the version in which `-v` was overloaded.
   - README flag table updated; `project.jsonc` updated.
   - The changelog entry must call out the `-v` change as a fix, not a feature.

6. **Tests.**
   - Invocation matrix: `-v`, `--verbose`, `-V`, `--version`, `version`, `-v daily`, `daily -v`, `-V daily`, `daily -V`, `-v -V`, `-V --bogus`, `--json` variants, empty HOME.
   - Assert `-v` never prints version metadata and `-V` never executes a command.
   - JSON parity with a fixed clock or by comparing all keys except `generatedAt`.
   - Help test: both spellings present, no contradictory text.
   - No-side-effect test: version does not read data or hit the network.
   - A property test: for every position of `-V` in a valid argv, version is printed and exit `0`.
   - A grep test: no help string maps `-v` to version.

7. **Documentation and gate.**
   - `scripts/simulated_user_gate.sh`: assert AC-001 through AC-007.
   - README, MIGRATION, project.jsonc.
   - Help screenshots/examples updated if present.

## Non-Goals
- Changing the version string format or ldflags.
- Changing the `version` command's JSON schema beyond making it consistent.
- Adding build metadata (Go version, platform) to the output; a follow-up if desired.
- Changing `--verbose` diagnostics content.
- Changing help precedence (owned by the help goal).

## Acceptance Criteria
- AC-001: `thermal -v` exits `0` and prints the leaderboard (`Token Warriors`), not `thermal dev`.
- AC-002: `thermal -v daily` exits `0` and prints the daily report; `thermal daily -v` behaves identically.
- AC-003: `thermal -V` exits `0` and prints `thermal dev (built from source)`.
- AC-004: `thermal --version` and `thermal version` produce output identical to `-V`.
- AC-005: `thermal -V --json | jq -e '.type=="version"'` exits `0`; `thermal version --json` emits the same object.
- AC-006: `thermal -v --json` emits the resolved command's JSON (leaderboard), not the version object.
- AC-007: `thermal --help` contains `-V, --version` and `-v, --verbose`; it does not describe `-v` as version.
- AC-008: `thermal -v -V` prints version and exits `0`.
- AC-009: `thermal -V daily` prints version and exits `0`; the daily table is absent.
- AC-010: `thermal -V --bogus` prints version and exits `0` (documented precedence).
- AC-011: `HOME=$(mktemp -d) thermal -v` exits `1` with the empty-state message, not version output.
- AC-012: `go test -race ./cmd/thermal/... ./internal/render/...` exits `0`, including `TestVersionFlagMatrix`, `TestVerboseNeverPrintsVersion`, `TestVersionJSONParity`, and `TestVersionPrecedence`.
- AC-013: `./scripts/simulated_user_gate.sh` exits `0` and includes the version/verbose section.

## Negative Cases & Failure Modes
| Invocation | Required result |
| --- | --- |
| `-v` | verbose normal output, no version |
| `--verbose` | same |
| `-V` | version, exit 0 |
| `--version` | version, exit 0 |
| `version` | version, exit 0 |
| `-v daily` | verbose daily |
| `daily -v` | verbose daily |
| `-V daily` | version, command ignored |
| `daily -V` | version, command ignored |
| `-v -V` | version wins |
| `-V --bogus` | version wins, exit 0 |
| `-V --json` | version JSON |
| `version --json` | same JSON |
| `-v --json` | command JSON |
| `-V` empty HOME | version, exit 0 |
| `-v` empty HOME | empty-state error, exit 1 |

## Metrics & Regression Guards
- Context independence: `-v` behavior is identical in every position; a matrix test enforces it.
- Version purity: `-V`/`--version`/`version` produce the same text/JSON; a parity test enforces it.
- Help truth: no help string maps `-v` to version; a grep test enforces it.
- No side effects: version path performs no loader, pricing, cache, update-check, or network call.
- Gate: simulated user gate asserts AC-001, AC-003, AC-005, AC-007, and AC-009.

## Risks & Mitigations
- **Risk:** Users and scripts expect `-v` to print the version. **Mitigation:** MIGRATION entry, changelog callout, and a clear help line; `-V` is the new short form.
- **Risk:** A global version pre-scan conflicts with the help pre-scan. **Mitigation:** one precedence table and one resolver; test `-h -V`, `-V -h`, and `--help --version`.
- **Risk:** `-V --bogus` succeeding hides a real typo. **Mitigation:** document version > validation; users who want validation should omit `-V`. This matches git/node behavior.
- **Risk:** JSON parity tests fail on `generatedAt` differences. **Mitigation:** omit `generatedAt` from version JSON or inject a clock.
- **Risk:** `-v` with no command on an empty HOME now exits `1` instead of `0`. **Mitigation:** that is the correct no-data behavior; document in the same MIGRATION entry.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "-v version verbose overload"` and `sila recall "flag length special case"`.
- **Writes**: record a `case` for the "one spelling one meaning + precedence resolver" mechanism, grounded to `cmd/thermal/main.go`; record a `lesson` for the context-dependent flag.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "version/verbose contract shipped" --open="version JSON determinism"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-version-verbose-flag-contract/goal.md)` under a `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the invocation matrix.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
