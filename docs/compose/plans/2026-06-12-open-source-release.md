# Open Source Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prepare thermal for public release as `github.com/jadmadi/thermal` with Go package install and cross-platform binaries.

**Architecture:** Rename the Go module to the full GitHub import path, add LICENSE/README metadata, and set up GoReleaser + GitHub Actions for automated binary releases on tag push.

**Tech Stack:** Go, GoReleaser, GitHub Actions

---

### Task 1: Rename Go Module

**Files:**
- Modify: `go.mod`
- Modify: `cmd/thermal/main.go:12-14`
- Modify: `internal/loaders/sessions.go:12`
- Modify: `internal/loaders/jsonl.go:13`
- Modify: `internal/loaders/sqlite.go:6`
- Modify: `internal/loaders/registry.go:9`
- Modify: `internal/render/leaderboard.go:9`
- Modify: `internal/render/dashboard.go:8`
- Modify: `internal/render/heatmap.go:6`

- [ ] **Step 1: Update go.mod module path**

Change line 1 of `go.mod` from:
```
module thermal
```
to:
```
module github.com/jadmadi/thermal
```

- [ ] **Step 2: Update all internal imports**

Replace all occurrences of `"thermal/` with `"github.com/jadmadi/thermal/` across these files:
- `cmd/thermal/main.go` (lines 12-14)
- `internal/loaders/sessions.go` (line 12)
- `internal/loaders/jsonl.go` (line 13)
- `internal/loaders/sqlite.go` (line 6)
- `internal/loaders/registry.go` (line 9)
- `internal/render/leaderboard.go` (line 9)
- `internal/render/dashboard.go` (line 8)
- `internal/render/heatmap.go` (line 6)

- [ ] **Step 3: Verify build compiles**

Run: `go build ./cmd/thermal`
Expected: builds successfully with no errors

- [ ] **Step 4: Commit**

```bash
git add go.mod cmd/thermal/main.go internal/
git commit -m "chore: rename module to github.com/jadmadi/thermal"
```

### Task 2: Add MIT License

**Files:**
- Create: `LICENSE`

- [ ] **Step 1: Create LICENSE file**

```
MIT License

Copyright (c) 2026 Jafar Madi

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 2: Commit**

```bash
git add LICENSE
git commit -m "docs: add MIT license"
```

### Task 3: Update README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Replace README content**

Update install URLs from `github.com/user/thermal` to `github.com/jadmadi/thermal`. Add author/support section at the bottom.

Replace the full content of `README.md` with:

```markdown
# Thermal

> Don't break the streak.

GitHub-style contribution heatmap for AI coding tools.

See your coding streaks, daily activity, and usage patterns rendered as a beautiful terminal heatmap. Default mode shows a **leaderboard** ranking all your installed tools.

## Supported Tools

| Tool | Data Source | Metrics |
|------|-------------|---------|
| **MiMoCode** | SQLite DB | Token usage, sessions, cost |
| **OpenCode** | SQLite DB | Token usage, sessions, cost |
| **Codex** | JSONL history | Command activity |
| **Agy** | Transcript JSONL | Conversation steps |
| **command-code** | JSONL history | Command activity |
| **codewhale** | JSON sessions | Session activity |

## Install

```bash
go install github.com/jadmadi/thermal@latest
```

Or build from source:

```bash
git clone https://github.com/jadmadi/thermal
cd thermal
go build -o thermal ./cmd/thermal
```

## Usage

```bash
# Default: leaderboard showing all installed tools
thermal

# Show a specific tool's heatmap
thermal --tool opencode
thermal --tool codex

# Show last 26 weeks
thermal --weeks 26

# Output raw JSON
thermal --json

# Disable colors
thermal --no-color
```

## Example: Leaderboard

```
  THERMAL — Don't break the streak.
  Thursday, June 11 2026

  #  Tool            Streak    Best      Active    Activity
  ──────────────────────────────────────────────────────────────
  🥇 OpenCode        4d        13d       40d       1.6B tok
     ████████████████████ 🔥🔥
  🥈 command-code    2d        9d        23d       374 cmd
     ██████████
  🥉 MiMoCode        2d        2d        3d        339M tok
     ██████████
     Codex           0d        2d        4d        62 msg
     Agy             0d        1d        1d        57 step
     codewhale       0d        0d        0d        0 sess

  🏆 OpenCode is on fire with a 4-day streak!

  Keep the heat going. Don't break the streak.
```

## Example: Single Tool

```
  OpenCode activity  1.2B tokens / 8 weeks  ~/.local/share/opencode/opencode.db

      ApMay  J
      ░░█▒░░░░
  Mon ░░█▓░░▒
      ░░▒░░░░█
  Wed █▒▒▒░░░█
      ▓█▓░░░░▓
  Fri ░▒░░░░▓
      ░█▓░░░▒
      Less □░▒▓█ More

  28 active days  |  4 day streak  |  13 best  |  1.6B all-time
```

## How It Works

Thermal reads usage data from installed AI coding tools:

- **MiMoCode / OpenCode**: Queries SQLite databases for assistant message token usage
- **Codex / command-code**: Parses JSONL command history for activity timestamps
- **Agy**: Reads transcript JSONL files from conversation logs
- **codewhale**: Reads JSON session files

Databases are opened **read-only** — Thermal never modifies your data.

## Requirements

- Go >= 1.21

## Author

**Jafar Madi** — [jadmadi.net](https://jadmadi.net) · [@jadmadi](https://x.com/jadmadi) · [jadmadi@gmail.com](mailto:jadmadi@gmail.com)

## Support

If you find Thermal useful, consider supporting:

[![PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?logo=paypal&logoColor=white)](https://paypal.me/Madise)

## License

[MIT](LICENSE)
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: update README with correct URLs and author info"
```

### Task 4: Add GoReleaser Config

**Files:**
- Create: `.goreleaser.yml`

- [ ] **Step 1: Create .goreleaser.yml**

```yaml
version: 2

builds:
  - main: ./cmd/thermal
    binary: thermal
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

- [ ] **Step 2: Commit**

```bash
git add .goreleaser.yml
git commit -m "ci: add GoReleaser config for cross-platform builds"
```

### Task 5: Add GitHub Actions Release Workflow

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create release workflow**

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Commit**

```bash
mkdir -p .github/workflows
git add .github/workflows/release.yml
git commit -m "ci: add GitHub Actions release workflow"
```

### Task 6: Tag Initial Release

**Files:**
- None (git operations only)

- [ ] **Step 1: Verify everything builds**

Run: `go build ./cmd/thermal && go vet ./...`
Expected: no errors

- [ ] **Step 2: Create initial tag**

```bash
git tag -a v0.1.0 -m "v0.1.0: initial open source release"
```

- [ ] **Step 3: Push tag to trigger release**

```bash
git push origin v0.1.0
```

Note: The GitHub remote should be set to `github.com/jadmadi/thermal`. Set it up first if needed:
```bash
git remote set-url origin git@github.com:jadmadi/thermal.git
```
