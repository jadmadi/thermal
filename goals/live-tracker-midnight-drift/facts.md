# Facts: Preserve cumulative project baselines across midnight transitions (live-tracker-midnight-drift)

## Architectural Invariants & Constraints
- Real-time monitors that run across midnight boundaries must seamlessly roll over without emitting false bursts from prior days.
- In-memory state transitions must be idempotent and zero-allocation.

## Problem Evidence & Root Causes
- Location: `internal/thermal/live.go:159-165`
- Code:
  ```go
  if lt.todayDay != "" && currentDay != lt.todayDay {
      // Midnight transition: day boundary crossed
      lt.todayDay = currentDay
      lt.prevProjTokens = make(map[string]int64)
      lt.prevProjTurns = make(map[string]int)
      for _, pd := range projDays {
          if pd.Day == currentDay {
              key := pd.Tool + "\x00" + pd.Project
              lt.prevProjTokens[key] = pd.Tokens
              lt.prevProjTurns[key] = pd.Turns
          }
      }
      lt.activeProject = ""
      lt.activeModel = ""
  }
  ```
- Failure Mechanism: When midnight passes, `lt.prevProjTokens` and `lt.prevProjTurns` are wiped and re-populated with ONLY the new day's rows (`pd.Day == currentDay`).
- Consequence: As soon as a tool emits activity in the new day, the delta calculation `acc.turns - pPrevTurns` sums all days in `toolProjTotals` and subtracts the wiped/single-day baseline, re-emitting all historical lifetime turns into `RecentEvents`.
- Fix: On midnight transition, accumulate all days for `prevProjTokens` and `prevProjTurns`, or retain the existing cumulative map without wiping historical counts.

## File & Interface Contracts
- Edit: `internal/thermal/live.go` in midnight transition logic.
- Test: `internal/thermal/live_test.go` simulating `t1 = 23:59:59` and `t2 = 00:00:01`.
- Document: `docs/MIGRATION.md`
