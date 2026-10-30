# Thermal CLI Contract

This is the stable process and channel contract for `thermal`. Error text and
human wording may improve between releases; the exit codes and JSON envelopes
are the machine contract.

## Exit codes

| Code | Meaning | Example |
| --- | --- | --- |
| `0` | Success | `thermal version` |
| `1` | Data, IO, or network failure | `thermal opencode` with no database |
| `2` | Usage error | `thermal --tool open` |
| `3` | Refused non-interactive mutation | `thermal upgrade --json` |

`--require-data` changes a no-data report from `0` to `1` without changing its
stdout payload. `SIGTERM` during `thermal live` exits `143`; a second `SIGINT`
while shutdown is still in progress exits `130`.

## Channels

- stdout carries the requested payload only.
- stderr carries diagnostics, deprecation warnings, and human remediation.
- `--json` stdout is valid JSON; warnings stay on stderr.
- `live --json --stream` is NDJSON, one object per line.
- No ANSI on a pipe, under `--no-color`, `NO_COLOR`, or `TERM=dumb`.

## Global flags on every command

`--json`, `--no-color`, and `--verbose` are honored where they make sense.
`--offline` never touches the network; `upgrade --offline` fails closed because
release metadata is not cached. A known global flag is honored, a no-op, or
explicitly not applicable with an alternative; it is never "unknown".

## Envelopes

Every `--json` report returns the `type` for the requested verb. No-data
reports keep that type, use empty collections, and add `state:"no-data"` plus a
`next` object. `thermal sources --json` is the only `type:"sources"` producer.
See `docs/JSON-CONTRACT.md` for keys and precision.
