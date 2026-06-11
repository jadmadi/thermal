# mimo-streak

GitHub-style contribution heatmap for [MiMoCode](https://github.com/xiaomi/mimo-code) usage.

See your coding streaks, daily token activity, and usage patterns rendered as a beautiful terminal heatmap.

## Install

```bash
go install github.com/user/mimo-streak@latest
```

Or build from source:

```bash
git clone https://github.com/user/mimo-streak
cd mimo-streak
go build -o mimo-streak .
```

## Usage

```bash
# Default: show 52-week heatmap
mimo-streak

# Show last 26 weeks
mimo-streak --weeks 26

# Output raw JSON
mimo-streak --json

# Disable colors
mimo-streak --no-color

# Use a custom database path
mimo-streak --db /path/to/mimocode.db
```

## Example Output

```
  MiMoCode activity  324M tokens / 12 weeks  ~/.local/share/mimocode/mimocode.db

      MaApr May  J
      □□□□□□□□□□□□
  Mon □□□□□□□□□□□□
      □□□□□□□□□□□□
  Wed □□□□□□□□░░░░
      □□░░□□□□████
  Fri □□□□□□□□□□□□
      □□□□□□□□□□□□
      Less □░▒▓█ More

  3 active days  |  2 day streak  |  2 best  |  324M all-time
```

## How It Works

mimo-streak reads your MiMoCode SQLite database (stored at `~/.local/share/mimocode/mimocode.db`) and:

1. Queries all assistant messages for token usage and cost data
2. Groups activity by calendar day
3. Computes current and longest streaks (consecutive active days)
4. Renders a GitHub-style heatmap with intensity based on relative token usage

The database is opened **read-only** — mimo-streak never modifies your MiMoCode data.

## Requirements

- Go >= 1.21

## License

MIT
