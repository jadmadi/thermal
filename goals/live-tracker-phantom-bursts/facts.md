# Facts: Fix live tracker idle baseline collapse (live-tracker-phantom-bursts)

## Architectural Invariants & Constraints
- Live monitor metrics must stay consistent: recent burst events cannot contradict session burned volume or today's volume.
- Zero allocation and sub-10ms polling invariants must be preserved.

## Problem Evidence & Root Causes
- Location: `internal/thermal/live.go:806-813`
- Code:
  ```go
  } else {
      // Keep project baselines updated even without tool-level delta
      for _, pd := range projDays {
          if pd.Tool == res.Name && pd.Day == lt.todayDay {
              key := pd.Tool + "\x00" + pd.Project
              lt.prevProjTokens[key] = pd.Tokens
              lt.prevProjTurns[key] = pd.Turns
          }
      }
  }
  ```
- Contrast: Initialization (`internal/thermal/live.go:335-339`) and delta tracking (`internal/thermal/live.go:616-635`) accumulate `pd.Tokens` and `pd.Turns` across ALL days.
- Mechanism: On idle ticks (no new tool tokens or turns), the `else` branch overwrites `lt.prevProjTurns[key]` with only today's turns (`pd.Day == lt.todayDay`).
- Failure Mode: When the tool emits a single new turn on the next poll, `acc.turns - pPrevTurns` diffs the cumulative lifetime turns against the collapsed single-day baseline (e.g., 125.6K steps for `lab`), repeatedly emitting massive phantom bursts.
- Solution: Maintain cumulative lifetime baselines across all days on idle ticks, matching the delta branch.

## File & Interface Contracts
- Edit: `internal/thermal/live.go` in `LiveTracker.Poll`
- Test: `internal/thermal/live_test.go`
- Document: `docs/MIGRATION.md`
