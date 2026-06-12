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

With Go:

```bash
go install github.com/jadmadi/thermal@latest
```

Or download a pre-built binary from [Releases](https://github.com/jadmadi/thermal/releases).

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
