# Facts: Shrink the unpriceable token line

## What the line is today

Every cost footer names the tokens no price covers. On this machine that is 2,897,537 tokens out of a 31.9B window, and the whole figure comes from one Codex day. Two separate causes are hiding inside it, and only one is an attribution gap.

Measured with `projects --json` over the default window:

```
projects total tokens   31,933,566,785
unpriceable tokens          2,897,537
```

## Cause 1: one Codex day lost its model (2.9M tokens, worth fixing)

```
codex daily --last 90 --json
  2026-09-14  tokens 2897537  turns 1  models null
```

The session exists in the rollout logs and it names its model:

```
~/.codex/sessions/2026/09/08/rollout-2026-09-08T12-21-11-01a08052-....
  models seen: ['gpt-6-astra']   usage tokens 32,448,681
~/.codex/sessions/2026/09/15/rollout-2026-09-15T05-22-35-01a0a2df-8fd4-....
  models seen: ['codex-auto-review']  usage tokens 1,165,687
```

So the Codex loader reads models from rollout logs for most days (99 percent of its tokens carry one) but misses at least one session. The day sits in a separate month directory from the log that carries it, which points at the file-to-day mapping rather than the parsing. Finding that is the first task.

## Cause 2: activity-only tools counted as tokens (about 72K, a design question)

```
agy            65072 tokens  0% carrying a model   (tokens == turns, 500 steps is 500 turns)
command-code    5347 tokens  0% carrying a model
droid             14 tokens  0% carrying a model
muse               1 token   0% carrying a model
```

Agy's daily rows set `Tokens` to the step count, so a step count is being read as a token total and then reported as unpriceable. That came from the `hasTokenData` change, which accepts a positive `Tokens` value on its own. Before that change these rows were dropped and the line did not exist.

This is a containment problem, not an attribution problem, and it needs a decision rather than a patch:

- Treating an activity count as a token count is wrong in a token total.
- Dropping them again would undo the fix that stopped session-total sources, such as codewhale, disappearing from reports.
- The distinguishing fact is available: activity-only loaders know they are activity-only, and the unified `DailyRow` is what erases that. A marker on the row, set by the loader, would let reports count tokens from token sources and leave activity counts to the leaderboard and streaks.

## Architectural Invariants & Constraints
- The footer line stays, whatever else changes: an estimated total must never read as complete when part of it could not be priced.
- `hasTokenData` must keep accepting a positive `Tokens` total from a token source, because that fix stopped session-total sources vanishing from reports.
- Token types stay disjoint and add up to the row total.
- Anything that changes reported totals needs a note in the release, in the README, and in the guide's cost section.

## File & Interface Contracts
- `internal/loaders/codex.go`: the day-to-rollout mapping that loses a session's model.
- `internal/loaders/agy.go`, `commandcode.go`, `droid.go`, `muse.go`: the activity-only loaders, if Cause 2 is fixed here.
- `internal/thermal/types.go` and `internal/thermal/report.go`: the row marker and `hasTokenData`, if Cause 2 is fixed here.
