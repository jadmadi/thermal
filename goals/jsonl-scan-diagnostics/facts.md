# Facts: A scan that stops early is reported, never silent

## The gap

Eight scan sites in seven files use `bufio.Scanner` (agy has two). None of them
checks `scanner.Err()`. The only `.Err()` calls in `internal/loaders` are sqlite
`rows.Err()` at `sqlite.go:343`, `:641`, and `:783`, which are a different
thing.

When a scan stops early, for a line over the ceiling or a read error, the loop
ends, the loader keeps the lines it already read, and nothing says the file was
cut short. The sibling goal makes the ceilings fit real data; this goal is the
safety net under them.

## The current diagnostics contract

- The only warning today is a fatal load error: `cmd/thermal/main.go` prints
  `thermal: warning: failed loading <tool>: <err>` to stderr under `--verbose`
  and skips the whole tool (around lines 471 and 625).
- Loader functions return `(thermal.Summary, []thermal.DailyRow,
  []thermal.ProjectDay, error)`, and 13 tools share the signature, so the error
  cannot carry a per-file note without churn.
- `AGENTS.md` section 3.3: verbose warnings go to stderr only and must never
  pollute stdout or `--json`.
- Also silent, same class, inferred from the code and not observed on this
  machine yet: a worker whose `os.Open` fails returns an empty result with no
  note (grok, claude, droid, command-code; agy returns zero steps the same way).

## Chosen mechanism

- `Summary.Warnings []string` with `json:"-"`: loaders already return the
  summary, so no signature changes, and the field can never reach JSON output.
- Per-file results carry at most one warning; the loader merges them after
  `wg.Wait()`, so no mutex and no interleaving.
- `cmd/thermal` prints them only under `--verbose`:
  `thermal: warning: <tool>: <text>` on stderr.
- One shared formatter for the text. On `bufio.ErrTooLong` it names the path
  and the ceiling and says the rest of the file was skipped; any other error
  names the path and the error.
- The ceiling is a package variable, not a const, so a test can lower it and
  exercise the truncation path without writing 32 MiB of fixture.

Rejected: widening the loader signature to return warnings; a larger diff than
the problem, and the summary is already the per-tool result.

## What a warning must not do

- No stdout, ever; `--json` output stays byte-identical.
- No stderr without `--verbose`.
- No effect on counts: the partial data stays, as today.

Out of scope: TUI surfaces for warnings and any change to the `--json`
contract.
