# Facts: JSONL line ceilings that fit real logs

## What is wrong, in one line

A `bufio.Scanner` stops at the first line longer than its ceiling: `Scan`
returns false, the rest of the file is never read, and nothing is printed.
Seven loader files scan JSONL this way, and every ceiling is smaller than a
real log line can be. One is already over it.

## The live case: command-code

| Item | Value |
| --- | --- |
| File | `~/.commandcode/projects/home-jad-debian-temp/30ce078b-....jsonl` |
| Lines | 320, every one carries a timestamp |
| Longest line | line 283, 409,473 bytes, a `tool_result` message |
| Ceiling | 262,144 bytes |
| Read today | lines 1..282, all on 2026-08-01 (local day, +03) |
| Never read | lines 283..320: 14 messages on 2026-08-01, 24 on 2026-08-02 |

Exactly one of the 1,690 command-code transcript files is over its ceiling.
The undercount is visible in the tool's own JSON:

```json
{"total":10707,"window":[{"day":"2026-08-01","turns":294},{"day":"2026-08-02","turns":113}]}
```

Expected after the ceiling fits: total 10,707 -> 10,745, 2026-08-01
294 -> 308, 2026-08-02 113 -> 137. The deltas equal the file's unread lines,
14 and 24.

`thermal command-code daily` cannot show this: activity-only rows are excluded
from period reports since 0.9.1, so the tool's own JSON and the leaderboard
total are the proof channel.

## Every scanner, its ceiling, and the longest real line

Measured 2026-09-20 across the exact globs the loaders read.

| Loader | File set | Files | Ceiling | Longest line | Headroom |
| --- | --- | ---: | ---: | ---: | ---: |
| codex | rollout logs via `state_5.sqlite` | n/a | 32 MiB | 4,535,419 | 29 MiB |
| command-code | `~/.commandcode/projects/*/*.jsonl` | 1,690 | 256 KiB | 409,473 | over by 147,329 |
| grok | `~/.grok/sessions/*/*/updates.jsonl` | 2 | 256 KiB | 73,393 | 188,751 |
| droid | `~/.factory/sessions/*/*.jsonl` | 3 | 256 KiB | 57,837 | 204,307 |
| claude | `~/.claude/projects/*/*.jsonl` | 7 | 256 KiB | 45,976 | 216,168 |
| agy | `brain/*/.system_generated/logs/transcript.jsonl` | 213 | 256 KiB | 10,393 | 251,751 |
| jsonl fallback | `~/.codex/history.jsonl`, `~/.commandcode/history.jsonl` | 2 | 64 KiB default | 33,661 | 31,875 |

The other ceilings are not safe, only lucky. grok uses 28% of its cap today,
and one large paste grows a single line; the fallback scanners never call
`Buffer` at all, so they carry `bufio`'s 64 KiB default.

## Why 32 MiB, and where it goes

- `bufio.Scanner` grows its buffer only as far as the current line needs, so
  the ceiling is a bound, not an allocation.
- 32 MiB is 7x the largest line observed on this machine, and it is what codex
  has shipped since 0.9.1.
- The same shape hit codex before 0.9.1: a 1.3 MB rollout line stopped the
  scan before the `token_count` frames, and a 2.9M-token day lost its model.
- One shared constructor and constant in `internal/loaders`, adopted at every
  scan site, so the next loader cannot quietly pick a smaller number.

## Decisions

- Shared ceiling of 32 MiB.
- What happens when a scan still stops early is out of scope here; it is
  `jsonl-scan-diagnostics`.
- Partial data stays, as today; this goal changes what gets read, not what
  gets reported.

## Reproduction

```bash
F=~/.commandcode/projects/home-jad-debian-temp/30ce078b-782c-4d0c-be85-2898a0f1ff9f.jsonl
awk 'NR>=283' "$F" | grep -o '"timestamp":"[^"]*"' | cut -d'"' -f4 \
  | while read -r ts; do date -d "$ts" +%F; done | sort | uniq -c
#     14 2026-08-01
#     24 2026-08-02

thermal command-code --json | jq -c '{total: ([.daily[].turns] | add),
  window: [.daily[] | select(.day >= "2026-08-01" and .day <= "2026-08-02") | {day, turns}]}'
# {"total":10707,"window":[{"day":"2026-08-01","turns":294},{"day":"2026-08-02","turns":113}]}
```
