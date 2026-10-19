# Facts: Clamp project activity deltas to tool-level totals (live-tracker-delta-clamping)

## Architectural Invariants & Constraints
- Sub-parts of an aggregate can never exceed the aggregate: the sum of project burst deltas must be bounded by the tool's reported turn and token deltas.
- No allocation on the polling fast path.

## Problem Evidence & Root Causes
- Location: `internal/thermal/live.go:654-670`
- Code:
  ```go
  if tokSeen {
      pDTok = acc.tokens - pPrevTok
  } else {
      pDTok = acc.tokens
      if pDTok > deltaTok {
          pDTok = deltaTok
      }
  }
  if turnSeen {
      pDTurns = acc.turns - pPrevTurns
  } else {
      pDTurns = acc.turns
      if pDTurns > deltaTurns {
          pDTurns = deltaTurns
      }
  }
  ```
- Contrast: Notice that the `else` branch (first time seeing a project) correctly clamps `pDTok > deltaTok` and `pDTurns > deltaTurns`.
- Defect: The `if tokSeen` and `if turnSeen` branches do NOT clamp to `deltaTok` and `deltaTurns`.
- Consequence: If project baselines drift, get corrupted, or get clobbered, a single tool turn can cause a project delta of tens of thousands of steps to be emitted into the ticker.
- Fix: Add clamping in both branches:
  ```go
  if tokSeen {
      pDTok = acc.tokens - pPrevTok
  } else {
      pDTok = acc.tokens
  }
  if pDTok > deltaTok {
      pDTok = deltaTok
  }

  if turnSeen {
      pDTurns = acc.turns - pPrevTurns
  } else {
      pDTurns = acc.turns
  }
  if pDTurns > deltaTurns {
      pDTurns = deltaTurns
  }
  ```

## File & Interface Contracts
- Edit: `internal/thermal/live.go` in project delta loop.
- Test: `internal/thermal/live_test.go`
- Document: `docs/MIGRATION.md`
