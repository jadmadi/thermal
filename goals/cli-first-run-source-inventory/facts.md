# Facts: Replace the empty-state dead end with a per-source inventory and next steps

## Observed Evidence (2026-10-29, current source build `thermal dev`, `HOME=$(mktemp -d)`)
- `thermal` → stderr `thermal: no supported tool data found`, exit `1`; stdout empty.
- `thermal --json` → same stderr line, stdout empty, exit `1`.
- `thermal --verbose` → the same line only; no probed paths, no state.
- `thermal stats` → `thermal: no supported tool data found`, exit `1`.
- `thermal live --json` → `thermal: no supported tool data found`, exit `1`.
- `thermal statusline` → `thermal: no supported tool data found`, exit `1`.
- `thermal share` → `thermal: no supported tool data found`, exit `1`.
- `thermal opencode` → `thermal: database not found: /tmp/…/.local/share/opencode/opencode.db`, exit `1`.
- `thermal --tool claude daily` → `thermal: data directory not found: /tmp/…/.claude`, exit `1`.
- `thermal audit` → works (audits instruction files in the CWD, not agent data).

## Root Cause
- `cmd/thermal/main.go:758`, `:917`, and `:1629` print the generic line and exit `1` when the loaded result set is empty. All three discard the information the registry already has.
- `cmd/thermal/main.go:861-870` (`toolHasData`) already checks `info.DBPath`/`info.DataDir` existence, but only for skipping tools; it does not expose state.
- `internal/loaders/registry.go:32-141` (`AllTools`) knows every default path and directory; no API surfaces it.
- There is no `sources`/`doctor` command in the dispatch or help.
- The per-tool loaders return specific errors (`database not found`, `data directory not found`) but the aggregate path never renders them.

## Architectural Invariants & Constraints
- The inventory is metadata only: `stat` and directory entry counts; never open a database, read a JSONL file, or parse a session.
- Zero telemetry: do not record activation events.
- The default command remains read-only and must not create the missing directories or files.
- Human output goes to stderr on the error path and stdout for the explicit `sources` command; JSON goes to stdout and is exactly one object.
- Paths must respect the same environment overrides as the loaders (`HERMES_HOME`, `DSH_HOME`, and any platform-specific home).
- The inventory must be fast enough for a shell prompt and a first run: <100 ms on an empty HOME.
- A corrupt database is still `found` at the metadata level; the report command owns the open error. Do not hide the path because the DB is unreadable.

## Source States
| State | Detection | Human detail |
| --- | --- | --- |
| `found` | path exists, readable, file size > 0, or directory has ≥1 entry | `expected file` / `N entries` |
| `empty` | path exists, readable, zero-byte file or empty directory | `no sessions yet` |
| `missing` | `os.Stat` returns not-exist | `expected <path>` |
| `unreadable` | permission denied or timeout | `check permissions` |
| `wrong-shape` | file for a directory tool, directory without the expected file for a file tool | `expected <shape>` |
| `unknown` | registry has no path | `no configured path` |

## File & Interface Contracts
- `internal/loaders/inventory.go` (new):
  ```go
  type SourceState string // found, empty, missing, unreadable, wrong-shape, unknown
  type SourceInfo struct {
      Tool, Name, Alias string
      Path, DisplayPath string
      Shape string // file, directory, either
      State SourceState
      Detail string
      Modified time.Time
  }
  func Inventory(home string) []SourceInfo
  func InventorySummary(items []SourceInfo) Summary
  ```
- `internal/render/sources.go` (new):
  ```go
  func RenderSources(items []SourceInfo, opts SourcesOptions) string
  func RenderSourcesJSON(items []SourceInfo, home string, now time.Time) string
  ```
- `cmd/thermal/main.go`: add `case "sources"`; replace the three generic empty-state sites with `renderEmptyState(cmd string, opts thermal.Options)`.
- `internal/render/help.go`: add `sources` to the command list and examples.
- `internal/loaders/inventory_test.go`, `cmd/thermal/sources_test.go`: fixtures.

## Text Layout (80 columns)
```
  Thermal · sources · 15 supported tools

  TOOL              ALIAS   PATH                                   STATE       DETAIL
  OpenCode          oc      ~/.local/share/opencode/opencode.db    found       1.2 GB
  Claude            —       ~/.claude                              empty       no sessions yet
  Codex             —       ~/.codex/state_5.sqlite                missing     expected file
  ...

  Summary: 2 found · 1 empty · 11 missing · 1 unreadable
  Legend: found = data available · empty = no sessions · missing = path absent
          unreadable = permission denied · wrong-shape = unexpected file or directory

  Next: run one prompt in any supported tool, then re-run thermal.
        thermal sources --json   machine-readable inventory
        thermal --help           supported tools and flags
```
- At narrow widths, drop the `DETAIL` column, then shorten paths from the left.
- State is a word, never color-only.
- `--no-color`/`NO_COLOR` produce no ANSI.
- JSON uses absolute paths and includes `displayPath` with `~`.

## JSON Contract
```json
{
  "type": "sources",
  "generatedAt": "2026-10-29T19:00:00Z",
  "home": "/home/user",
  "summary": {
    "total": 15, "found": 2, "empty": 1, "missing": 11,
    "unreadable": 1, "wrongShape": 0
  },
  "sources": [
    {
      "tool": "opencode",
      "name": "OpenCode",
      "alias": "oc",
      "path": "/home/user/.local/share/opencode/opencode.db",
      "displayPath": "~/.local/share/opencode/opencode.db",
      "shape": "file",
      "state": "found",
      "detail": "",
      "modified": "2026-10-29T18:00:00Z"
    }
  ]
}
```
- `sources` is always length 15, stable order.
- No raw errors by default; `--verbose` may add a `"detail"` string.
- Never include file contents or session metadata.

## Verification Contract
```bash
empty=$(mktemp -d)

HOME=$empty thermal >/tmp/o 2>/tmp/e; test $? -eq 1
test ! -s /tmp/o
grep -q 'OpenCode' /tmp/e
grep -q 'missing' /tmp/e
grep -q 'thermal sources' /tmp/e

HOME=$empty thermal --json >/tmp/j 2>/dev/null; test $? -eq 1
jq -e '.type=="sources" and .summary.total==15 and .summary.found==0 and (.sources|length)==15' /tmp/j

HOME=$empty thermal sources >/tmp/o 2>/tmp/e; test $? -eq 0
test -s /tmp/o
test ! -s /tmp/e

HOME=$empty thermal sources --json | jq -e '.sources|length==15'

# One-tool fixture
fixture=$(mktemp -d); mkdir -p "$fixture/.local/share/opencode"
sqlite3 "$fixture/.local/share/opencode/opencode.db" 'create table session(id text);'
HOME=$fixture thermal sources --json | jq -e '.summary.found==1 and (.sources[]|select(.tool=="opencode")|.state)=="found"'

# Privacy
printf 'SENTINEL-PROMPT-TEXT' > "$fixture/.claude/projects/x/session.jsonl" 2>/dev/null || true
HOME=$fixture thermal sources | grep -q 'SENTINEL-PROMPT-TEXT' && exit 1

# Performance
/usr/bin/time -f '%e' env HOME=$empty thermal sources >/dev/null

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/loaders/... ./internal/render/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- Empty temp HOME.
- One OpenCode SQLite fixture with a minimal `session` table.
- One Claude directory with a JSONL file containing a sentinel prompt.
- Empty `~/.claude` directory.
- Zero-byte `opencode.db`.
- Directory at the expected SQLite file path.
- File at the expected Claude directory path.
- Permission-denied directory (skip as root).
- `HERMES_HOME` and `DSH_HOME` pointing at fixtures.
- A symlinked data directory.

## Open Questions
- Should bare `thermal` print the inventory to stdout or stderr? Recommendation: stderr, to keep stdout empty on error and preserve pipe hygiene.
- Should `thermal --json` with no data emit the sources JSON or remain empty? Recommendation: emit `type:"sources"` on stdout and exit `1`, documented in MIGRATION; `thermal sources --json` is the exit-0 form.
- Should `thermal sources` exit `0` when every source is missing? Recommendation: yes; it successfully produced the inventory.
- Should the inventory include session counts? Recommendation: no; counting requires reading, violates the stat-only invariant, and belongs to the loaders.
- Should `sources` be a top-level verb or an alias of `audit`? Recommendation: top-level `sources`; `audit` owns instruction/MCP health and has a different scoring contract.
