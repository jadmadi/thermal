# CLI UX Review → Goal Traceability

Source review: `cli-reviewer` three-pass audit of `/tmp/thermal-cli-review`
(`thermal dev`) on 2026-10-30. Every reported friction point maps to exactly one
primary goal; the goal's acceptance criteria are the verification contract.

## Coverage Matrix

| ID | Finding | Goal | Proving AC |
| --- | --- | --- | --- |
| B1 | Flagship help example dead-ends on missing OpenCode | cli-data-access-error-guidance | AC-001, AC-010 |
| B2 | Deprecation deadline `v0.14.0` is seven releases past | cli-version-and-deprecation-truth | AC-007, AC-009 |
| B3 | Fresh HOME exits 1 with a 28-line stderr inventory | cli-empty-data-success-contract | AC-001, AC-004 |
| B4 | Sources table truncates the useful path and detail | cli-terminal-output-integrity | AC-005, AC-006 |
| B5 | `version` prints `dev` despite a SemVer/commit promise | cli-version-and-deprecation-truth | AC-001, AC-004 |
| B6 | `-v` produces no verbose signal | cli-version-and-deprecation-truth | AC-005, AC-006 |
| B7 | `help streak` / `help token` are unknown commands | cli-help-glossary-and-routing | AC-001, AC-002 |
| B8 | `doctor` / `config` answered as unknown tools | cli-help-glossary-and-routing | AC-005 |
| B9 | `top` suggests the two-letter alias `oc` | cli-help-glossary-and-routing | AC-006 |
| B10 | Populated leaderboard has no next step | cli-help-glossary-and-routing | AC-010 |
| B11 | Legend wraps between a marker and its gloss | cli-terminal-output-integrity | AC-004 |
| C1 | `--last` means days/weeks/months per report | cli-window-flag-semantics | AC-001–AC-004 |
| C2 | Repeated `--tool` silently keeps the last value | cli-window-flag-semantics | AC-006 |
| C3 | Stale deprecation deadline (duplicate of B2) | cli-version-and-deprecation-truth | AC-007, AC-009 |
| C4 | `--tool open` has no hint; `code` suggests `ccode` | cli-help-glossary-and-routing | AC-007, AC-008 |
| C5 | `database not found` has no next command | cli-data-access-error-guidance | AC-001, AC-002 |
| C6 | Ctrl+C in `live` reports a failure and exits 1 | cli-sigint-lifecycle-contract | AC-001, AC-002 |
| C7 | A newline in a path tears the report frame | cli-terminal-output-integrity | AC-007 |
| C8 | Symlink loop and permission denied share one message | cli-data-access-error-guidance | AC-005, AC-006 |
| C9 | `--` produces `unknown tool "--offline"` | cli-window-flag-semantics | AC-008 |
| C10 | `<path>` examples break when pasted | cli-window-flag-semantics | AC-009, AC-010 |
| S1 | `COLUMNS=40` refuses a non-TTY report | cli-terminal-output-integrity | AC-001–AC-003 |
| S2 | Empty `--json` swaps to `type:"sources"` | cli-empty-data-success-contract | AC-002, AC-003 |
| S3 | `upgrade --check --offline` says unknown flag | cli-automation-exit-and-upgrade-parity | AC-001, AC-002 |
| S4 | `statusline` ignores `NO_COLOR`; `--plain` keeps ANSI | cli-terminal-output-integrity | AC-008–AC-010 |
| S5 | Truncated JSONL drops records silently | jsonl-partial-read-observability | AC-001–AC-003 |
| S6 | No-data statusline is empty and unexplained | cli-empty-data-success-contract | AC-006–AC-008 |
| S7 | Exit codes 0/1/2/3 are undocumented | cli-automation-exit-and-upgrade-parity | AC-007–AC-010 |
| T1 | `prompt` emits truecolor into a pipe | cli-terminal-output-integrity | AC-008, AC-009 |
| T2 | `sources` ignores width below 60 columns | cli-terminal-output-integrity | AC-006 |

## Green Flags to Preserve

These behaviors passed the review and must not regress while implementing the
goals above. Each implementing goal adds at least one guard for the rows it
touches.

- Typo recovery: `daly`, `weeky`, `projets`, `modles`, `licenc`, `stat`, and
  `--jsno` all suggest the right word, exit `2`, and print no stack trace.
- Malformed values name a range and a runnable example; reversed windows are
  corrected with a `did you mean` line.
- Piped reports emit zero ANSI and zero stderr bytes; `--verbose` warnings stay
  on stderr; `--json --verbose` remains valid JSON.
- `--json` is width-independent; 16 documented surfaces parse; NDJSON is one
  object per line.
- `weekly | head -n 1` exits cleanly with no broken-pipe text; 16 concurrent
  runs are byte-identical with valid cache JSON.
- SIGINT restores alt-screen, cursor, and bracketed paste; no lock/pid/temp
  files remain; stream interruption exits `0`.

## Execution Order

1. `cli-empty-data-success-contract` (activation; unblocks the JSON contract).
2. `cli-data-access-error-guidance` (recovery; reuses the inventory).
3. `cli-window-flag-semantics` (correct numbers; parser-local).
4. `cli-terminal-output-integrity` (rendering; touches the same surfaces).
5. `cli-sigint-lifecycle-contract` (interruption; live/stream only).
6. `cli-help-glossary-and-routing` (discoverability; after empty/data errors).
7. `cli-version-and-deprecation-truth` (trust; metadata and warnings).
8. `cli-automation-exit-and-upgrade-parity` (scripting; exit-code docs).
9. `jsonl-partial-read-observability` (data trust; loader-local).

The sequence numbers in `sila goals` are 255–263. Dependencies are recorded on
each package; none is blocked.

## Notes

- These goals extend, not replace, the verified packages
  `cli-first-run-source-inventory`,
  `cli-nontty-and-statusline-output-contract`,
  `cli-version-verbose-flag-contract`,
  `cli-argument-parser-strictness`,
  `cli-db-override-scope-and-errors`,
  `cli-terminal-width-adaptation`,
  `cli-leaderboard-ranking-and-legend`,
  `cli-upgrade-nondestructive-safety-flags`, and
  `jsonl-scan-diagnostics`.
- Two contracts are superseded in part:
  `cli-empty-data-success-contract` replaces the empty-state exit-`1` and
  `type:"sources"` swap from `cli-first-run-source-inventory`;
  `cli-terminal-output-integrity` tightens the `COLUMNS` and statusline-colour
  behavior from `cli-nontty-and-statusline-output-contract`.
- Each package carries `goal.md`, `facts.md`, and `plan.md` with observed
  output, file-level root cause, interfaces, requirements, acceptance criteria,
  negative cases, metrics, risks, memory plan, and done condition.
