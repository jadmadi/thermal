# Goal: Make thermal upgrade previewable, scriptable, and non-destructive by default

## Goal Description
Priority: P0 (safety). `thermal upgrade` is the only command that mutates the user's filesystem outside the cache, yet it has no option parsing at all. `cmd/thermal/main.go:585-586` calls `os.Exit(runUpgrade())`, and `runUpgrade()` ignores every argument. Unknown flags are silently accepted by the hand-rolled post-positional scanner, so every safety word a user might try is a no-op that authorizes the destructive path.

Observed on the current source build (2026-10-29), using disposable binary copies:

- `thermal upgrade --help` downloaded `thermal_0.21.0_linux_amd64.tar.gz`, replaced the binary, printed `✓ upgraded to v0.21.0`, and exited `0`; md5 changed from `43110b4…` to `ced2913…`.
- `thermal upgrade --dry-run` performed the same real upgrade; the binary changed.
- `thermal upgrade --json` performed the same real upgrade and printed human prose on stdout; no JSON was emitted.
- `thermal upgrade --check` on an up-to-date v0.21.0 printed `already up to date — v0.21.0` and exited `0`; on a `dev` build the same invocation upgrades.
- No confirmation prompt exists, and no non-TTY guard prevents a piped or CI invocation from replacing the binary.

The outcome of this goal is an upgrade command with an explicit, documented state machine: **preview (`--check`, `--dry-run`) never downloads or mutates; mutation requires either an interactive confirmation or `--yes`; `--json` produces exactly one machine-readable object on stdout; unknown flags fail with exit 2; and every failure leaves the existing binary intact.**

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 243
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 1
- **Related implemented goals (do not regress)**: `cli-upgrade-file-integrity` owns archive extraction deduplication and `Close()` error propagation; `cli-staged-deprecation-and-upgrade` owns migration notices; `cli-daily-update-check` owns the background update cache. This goal owns flag semantics, confirmation, exit codes, and JSON.
- **Related new goal**: `cli-help-interception-and-scoped-help` must guarantee `upgrade --help` is intercepted before `runUpgrade` is called; this goal must still defend in depth.

## Files to Touch
- `cmd/thermal/upgrade.go`
- `cmd/thermal/upgrade_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/help.go`
- `internal/render/help.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-upgrade-nondestructive-safety-flags/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-upgrade-nondestructive-safety-flags/plan.md`](plan.md)
- **Evidence standard**: every AC must record the command, exit code, stdout/stderr, and binary md5 before/after.

## Requirements

1. **Explicit option parsing.**
   - Replace the argument-free `runUpgrade()` with `runUpgrade(opts upgradeOptions) int`.
   - `upgradeOptions` must carry: `Check bool`, `DryRun bool`, `JSON bool`, `Yes bool`, `Force bool`, `Out io.Writer`, `Err io.Writer`, `Client *http.Client`, `APIURL string`, `Executable func() (string, error)`, and the current `version.Info`.
   - Parse only the documented flags: `--check`, `--dry-run`, `--json`, `--yes`, `--force`, `-h/--help` (delegated to the help goal).
   - Any other token, including `-v`, `--quiet`, `--no-color`, or a positional, exits `2` with:
     `thermal: unknown flag for upgrade: --quiet` and `run thermal upgrade --help`.
   - `--check` and `--dry-run` are mutually exclusive only if their semantics differ; define and test the combination (recommendation: `--check --dry-run` behaves as `--check` and both are listed as compatible).
   - `--json` composes with `--check`, `--dry-run`, and a confirmed upgrade.

2. **State machine and default behavior.**
   - `--check`: fetch release metadata, compare versions, print/emit current, latest, update availability, and exit `0`; no download, no extraction, no mutation.
   - `--dry-run`: resolve the release, the matching asset name/URL, the executable path, and the exact mutation steps; print/emit the plan and exit `0`; no download, no extraction, no mutation, no `.old`/`.new` files.
   - `thermal upgrade` on a TTY: show current/latest and prompt `Upgrade to vX.Y.Z? [y/N]`; `y`/`yes` proceeds, anything else exits `3` with `upgrade cancelled`; no mutation on decline.
   - `thermal upgrade` with stdout or stdin not a TTY and no `--yes`: exit `3` with `thermal: refusing to replace the binary without a terminal; re-run with --yes or use --check`; no network mutation path is entered.
   - `--yes`: proceed without a prompt; required for non-TTY mutation.
   - `--force`: allow reinstalling the same version or downgrading to the latest release; without it, current == latest prints `already up to date` and exits `0`.
   - A `dev` build is treated as older than any release and must still require confirmation/`--yes`; `--check` on `dev` reports the latest release without mutating.

3. **Output and JSON contract.**
   - Human progress and diagnostics go to stderr; the final result line may go to stdout only in text mode.
   - `--json` writes exactly one JSON object to stdout and nothing else. Schema:
     ```json
     {
       "type": "upgrade",
       "action": "check|dry-run|upgraded|up-to-date|declined|error",
       "current": "dev|0.21.0",
       "latest": "v0.21.0",
       "asset": "thermal_0.21.0_linux_amd64.tar.gz",
       "target": "/usr/local/bin/thermal",
       "mutated": false,
       "migration": "https://jadmadi.net/project/thermal/migration",
       "error": ""
     }
     ```
   - `error` is empty on success; on failure `action` is `error`, `error` is a single actionable sentence, and exit is `1` (or `2` for usage).
   - JSON output must remain valid when stderr also contains progress; stdout must not contain ANSI.
   - Text mode must not print `✓ upgraded` before the swap has succeeded.

4. **Mutation safety.**
   - Download, extract, chmod, and (if implemented) checksum/signature verification must all complete before the first rename.
   - Keep the existing atomic swap (`current → .old`, `new → current`) and restore `.old` on failure.
   - Refuse to mutate when the target directory is not writable, when the target is a symlink whose resolved path is outside the expected prefix, or when the process lacks permission; print the exact path and remedy.
   - On any failure (HTTP non-200, truncated archive, extraction error, chmod error, rename error), the original binary must remain byte-identical; verify with md5.
   - Remove `.old` and `.new` on success; leave no partial artifacts after failure.
   - Do not write to the target when `--check` or `--dry-run` is set, even if a stale `.old` exists.

5. **Testability without the network.**
   - All HTTP access must go through the injected `Client` and `APIURL`; no test may call `api.github.com`.
   - Provide a test helper that builds a release JSON and a valid `.tar.gz` in a temp dir, and serves them from `httptest.Server`.
   - Allow tests to inject the executable path so the swap happens inside a temp dir, never on the test binary.
   - Add a timeout to every request and keep the existing 30 s/120 s split or document a replacement.
   - Tests must cover: `--help` (no network), `--check`, `--dry-run`, `--json`, unknown flag, non-TTY refusal, TTY `n`, TTY `y` with stub, 404, truncated archive, read-only target, same-version, `--force`.

6. **Help, docs, and migration.**
   - The upgrade help page lists every flag, the three exit codes, the non-TTY rule, and two examples (`thermal upgrade --check`, `thermal upgrade --yes`).
   - `README.md` documents preview-before-upgrade and the `--yes` requirement for automation.
   - `docs/MIGRATION.md` gets a six-part entry: non-TTY now refuses without `--yes`; scripts must add `--yes` or use `--check`; detection is `thermal upgrade </dev/null` exiting `3`; fix is `thermal upgrade --yes`; escape hatch is `--check`/`--dry-run`.
   - `project.jsonc` lists the new flags and behavior.
   - `scripts/simulated_user_gate.sh` asserts: `upgrade --help` exit 0, `--check --json` parses, `--dry-run` leaves md5 unchanged, unknown flag exits 2, and no network is attempted for help/unknown flags.

## Non-Goals
- Changing the release asset naming, GoReleaser configuration, or provenance workflow.
- Adding a package-manager-aware upgrade strategy (brew/apt) in this goal.
- Changing the daily background update check or its cache schema.
- Implementing signature verification beyond what already exists; if a checksum is added, it is a smaller follow-up goal.

## Acceptance Criteria
- AC-001: `thermal upgrade --help` exits `0`, contains `--check`, `--dry-run`, `--json`, `--yes`, and leaves a disposable binary's md5 unchanged.
- AC-002: `thermal upgrade --check --json` exits `0`; stdout is one valid JSON object with `type:"upgrade"`, `action:"check"`, non-empty `current` and `latest`; no download request reaches the stub server.
- AC-003: `thermal upgrade --dry-run --json` exits `0`, `action:"dry-run"`, `mutated:false`, and no `.old`/`.new` files exist next to the target.
- AC-004: In the `httptest` upgrade harness, `upgrade --json --yes` with the stub serving the current version exits `0` with `action:"up-to-date"` and no mutation.
- AC-005: In the `httptest` upgrade harness, `upgrade --json --yes` with the stub serving a newer version exits `0` with `action:"upgraded"`, `mutated:true`, and the target file is replaced (verified by content hash inside the temp dir).
- AC-006: `thermal upgrade --force-now` exits `2`, names `--force-now` as unknown, and makes no HTTP request.
- AC-007: `thermal upgrade </dev/null` (non-TTY, no `--yes`) exits `3`, prints `refusing to replace the binary without a terminal`, and leaves the binary unchanged.
- AC-008: In a PTY, `thermal upgrade` followed by `n` exits `3`, prints `upgrade cancelled`, and leaves the binary unchanged; `y` proceeds.
- AC-009: In the `httptest` harness, a stub 404 exits `1`, prints an actionable error, and leaves the binary byte-identical.
- AC-010: In the `httptest` harness, a truncated `.tar.gz` exits `1`, leaves the binary byte-identical, and removes `.new`/`.old`.
- AC-011: `go test -race ./cmd/thermal/...` exits `0`, including `TestUpgradeCheckNoMutation`, `TestUpgradeDryRunNoMutation`, `TestUpgradeJSONContract`, `TestUpgradeUnknownFlag`, `TestUpgradeNonTTYRefusal`, `TestUpgradeDecline`, `TestUpgradeFailureRollback`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the upgrade safety section.

## Negative Cases & Failure Modes
| Invocation / condition | Required result |
| --- | --- |
| `upgrade --help` | Help, exit 0, no network, no mutation |
| `upgrade --check` | Metadata + plan, exit 0, no mutation |
| `upgrade --dry-run` | Plan with asset/target, exit 0, no download |
| `upgrade --json` on a TTY | Prompt or JSON? Define: `--json` implies non-interactive; require `--yes` or exit 3 with JSON error |
| `upgrade </dev/null` | Exit 3, no mutation |
| `upgrade --yes` with no network | Exit 1, actionable error |
| Unknown flag | Exit 2, no network |
| Same version | `up-to-date`, exit 0, no mutation |
| `dev` build | Report latest, require confirmation/`--yes` |
| Read-only target dir | Exit 1, path + remedy, no mutation |
| Rename failure | Restore original, exit 1, no `.new` |
| Truncated archive | Exit 1, original intact |
| `--check --json` with stderr progress | stdout remains one valid JSON object |

## Metrics & Regression Guards
- Mutation count: every non-mutating invocation leaves the target's inode and md5 unchanged; assert in tests.
- Network count: the stub server records requests; `--help`, unknown flags, and `--check` with a cached/known response make zero download requests.
- JSON purity: `thermal upgrade --check --json 2>/dev/null | jq -e .` succeeds; `... | jq -e . >/dev/null` must not see ANSI.
- Exit-code table is frozen in `cmd/thermal/upgrade_test.go` and mirrored in help.
- The simulated user gate checks AC-001, AC-002, AC-006, and AC-007 without network.

## Risks & Mitigations
- **Risk:** Defaulting to a prompt breaks automated installers that rely on `thermal upgrade`. **Mitigation:** `--yes` is the documented migration path; MIGRATION entry and a changelog note.
- **Risk:** A `--json` failure could emit partial JSON before the error. **Mitigation:** buffer the object and write once at the end; test with a failing stub.
- **Risk:** Tests accidentally mutate the test binary. **Mitigation:** inject `Executable` and always operate in `t.TempDir()`; add a guard that refuses to run tests with the real path.
- **Risk:** `--check` still hits the network, which is expected, but a proxy failure could look like a mutation failure. **Mitigation:** distinguish network errors (`cannot reach release API`) from mutation errors and document exit 1 for both.
- **Risk:** `--force` downgrades unexpectedly. **Mitigation:** print current → target and require the same confirmation/`--yes`; never force without an explicit flag.

## Memory Plan
- **Prime**: run `sila prime` at task start and record the briefing timestamp.
- **Recall**: run `sila recall "self updater unknown flags destructive"` and `sila recall "atomic binary swap rollback"`; record match counts.
- **Writes**: record a `case` for the injectable upgrade state machine (client, API URL, executable path), grounded to `cmd/thermal/upgrade.go`; record a `lesson` for the "unknown flag authorized a destructive action" failure.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "upgrade safety flags shipped" --open="checksum verification"` at session end.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-upgrade-nondestructive-safety-flags/goal.md)` under a `fix(cli):` or `feat(cli):` subject.
3. Goal is certified (`certification.json` with `verified: true`) and `evidence.md` contains a standalone `Verdict: PASS` plus the md5-before/after table for every non-mutating path.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
