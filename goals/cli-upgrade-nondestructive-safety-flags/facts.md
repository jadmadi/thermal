# Facts: Make thermal upgrade previewable, scriptable, and non-destructive by default

## Observed Evidence (2026-10-29, current source build `thermal dev`)
All destructive observations were made on disposable copies (`/tmp/upgtest`, `/tmp/upg-help`); the repository binary was not modified.

- `thermal upgrade --dry-run` output:
  ```
    thermal dev (built from source)  checking for updates...
    current: dev (built from source)
    latest:  v0.21.0
    downloading thermal_0.21.0_linux_amd64.tar.gz...
    ✓ upgraded to v0.21.0
    restart thermal to use the new version.
  ```
  Exit `0`; md5 `43110b4bb5cefc11550fbdd6cdb40c0` → `ced291366c7a97d0a7bd927d0c80741c`.
- `thermal upgrade --help` produced the same upgrade; md5 changed identically.
- `thermal upgrade --json` produced the same human prose on stdout; no JSON, exit `0`.
- `thermal upgrade --check` on the upgraded v0.21.0 binary printed `already up to date — v0.21.0`, exit `0` (it did not prove that `--check` is non-mutating on an older build).
- `thermal --help` documents only `upgrade  Self-upgrade thermal to the latest release`; no flags, exit codes, or non-TTY rule.

## Root Cause
- `cmd/thermal/main.go:585-586`:
  ```go
  case "upgrade":
      os.Exit(runUpgrade())
  ```
  No options are parsed or passed.
- `cmd/thermal/upgrade.go:49-59`: `runUpgrade() int` takes no arguments and immediately prints `checking for updates` and creates a 30 s context for the GitHub API request.
- The post-positional scanner in `cmd/thermal/main.go:123-273` has no `upgrade`-specific cases and silently ignores unknown flags. `--help`, `--dry-run`, `--check`, and `--json` are therefore all no-ops on this path.
- `cmd/thermal/upgrade.go:120-190`: download, extract, chmod, copy to `.new`, rename current → `.old`, rename `.new` → current, remove `.old`. Safety depends entirely on the operator not invoking the command.
- `runUpgrade` writes progress with `fmt.Printf` (stdout), so a JSON consumer receives prose.
- There is no TTY detection in `runUpgrade`; `isatty` is used elsewhere but not here.

## Architectural Invariants & Constraints
- `thermal upgrade` is the only command allowed to mutate the installed binary; every other command is read-only (cache writes excepted).
- A preview must be provably non-mutating: no download, no extraction, no chmod, no rename, no `.old`/`.new`.
- A failed upgrade must leave the original binary byte-identical and remove partial artifacts.
- Machine output (`--json`) must be one valid JSON object on stdout, with diagnostics on stderr.
- Non-interactive execution must never mutate without an explicit `--yes`.
- The upgrade path must not depend on the current working directory or on `$PATH` resolution; use `os.Executable()` and resolve symlinks as today.
- Existing integrity behavior from `cli-upgrade-file-integrity` (single extraction branch, `Close()` error propagation) must be preserved.

## File & Interface Contracts
- `cmd/thermal/upgrade.go`:
  - `type upgradeOptions struct { Check, DryRun, JSON, Yes, Force bool; Out, Err io.Writer; Client *http.Client; APIURL string; Executable func() (string, error); Current version.Info }`
  - `func runUpgrade(opts upgradeOptions) int`
  - `func parseUpgradeArgs(args []string) (upgradeOptions, error)` — rejects unknown flags.
  - `func resolveUpgradePlan(...) (upgradePlan, error)` — current/latest/asset/assetURL/target.
  - `func performUpgrade(plan upgradePlan, opts upgradeOptions) error` — the only mutating function.
- `cmd/thermal/main.go`: pass `os.Args[2:]` and `os.Stdout`/`os.Stderr` into `runUpgrade`; keep `os.Exit` at the dispatch boundary.
- `cmd/thermal/upgrade_test.go`: `httptest.Server` serving `githubRelease` JSON and a `.tar.gz` built in `t.TempDir()`.
- `scripts/simulated_user_gate.sh`: shell-level assertions only; no network.

## JSON Contract
```json
{
  "type": "upgrade",
  "action": "check",
  "current": "dev",
  "latest": "v0.21.0",
  "asset": "thermal_0.21.0_linux_amd64.tar.gz",
  "target": "/tmp/thermal-copy",
  "mutated": false,
  "migration": "https://jadmadi.net/project/thermal/migration",
  "error": ""
}
```
- `action` ∈ `check | dry-run | upgraded | up-to-date | declined | error`.
- `mutated` is `false` for every action except `upgraded`.
- `error` is empty on success.
- The object is written once, at the end, after all progress; stdout contains no ANSI and no prose.

## Verification Contract
```bash
# Non-mutating previews
cp thermal /tmp/thermal-copy
before=$(md5sum /tmp/thermal-copy | cut -d' ' -f1)
/tmp/thermal-copy upgrade --check --json | jq -e '.type=="upgrade" and .action=="check" and .mutated==false'
/tmp/thermal-copy upgrade --dry-run --json | jq -e '.action=="dry-run" and .mutated==false'
after=$(md5sum /tmp/thermal-copy | cut -d' ' -f1)
test "$before" = "$after"
test ! -e /tmp/thermal-copy.old && test ! -e /tmp/thermal-copy.new

# Usage errors
/tmp/thermal-copy upgrade --force-now >/dev/null 2>/tmp/up.err; test $? -eq 2
grep -q -- '--force-now' /tmp/up.err

# Non-TTY refusal (no network mutation)
/tmp/thermal-copy upgrade </dev/null >/dev/null 2>/tmp/up2.err; test $? -eq 3
grep -q 'refusing to replace the binary' /tmp/up2.err

# Unit tests use httptest only
CGO_ENABLED=0 go test -race ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A release JSON fixture with `tag_name: "v0.21.0"` and a linux/amd64 asset.
- A tar.gz builder that writes a tiny executable named `thermal` into the archive.
- A temp target path so the swap never touches the test binary.
- A read-only target directory for the permission-failure case.
- A stub server that can return 200, 404, truncated bodies, and slow responses.
- A PTY harness for the `y`/`n` prompt cases.

## Open Questions
- Should `thermal upgrade --json` on a TTY prompt, or imply non-interactive and require `--yes`? Recommendation: `--json` implies non-interactive; without `--yes`, emit an `action:"declined"` JSON object and exit `3`.
- Should `--force` allow downgrades as well as reinstalls? Recommendation: yes, but print `current → target` and require confirmation.
- Should checksum/signature verification be added here or in a follow-up? Recommendation: follow-up goal; this goal must not weaken existing integrity.
- What exit code for a declined prompt? Recommendation: `3` (distinct from validation `1` and usage `2`), documented in help.
