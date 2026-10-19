# Facts: Fix live tracker phantom bursts and fractional speed display

## Architectural Invariants & Constraints
- Live monitor metrics must stay consistent: the sum of recent burst events cannot contradict session burned volume or today's volume.
- Zero allocation and sub-10ms polling invariants must be preserved.
- No-color and terminal width formatting constraints must be honored across all speeds.

## Problem Evidence & Root Causes

### 1. Cumulative Baseline Collapse in Live Tracker
- Problem: In `internal/thermal/live.go:806-813`, when `deltaTok == 0 && deltaTurns == 0`, the tracker updates project baselines using only today's rows:
  ```go
  if pd.Tool == res.Name && pd.Day == lt.todayDay {
      key := pd.Tool + "\x00" + pd.Project
      lt.prevProjTokens[key] = pd.Tokens
      lt.prevProjTurns[key] = pd.Turns
  }
  ```
- Contrast: Initialization (`internal/thermal/live.go:335-339`) and delta tracking (`internal/thermal/live.go:616-635`) accumulate `pd.Tokens` and `pd.Turns` across ALL days.
- Consequence: On idle ticks, `lt.prevProjTurns[key]` collapses to today's count. When the tool subsequently emits even 1 turn, `acc.turns - pPrevTurns` computes `lifetime turns - today's turns` (e.g., 125.6K steps for `lab`, 3.0K steps for `scripts`) and emits them repeatedly into `RecentEvents`.
- In addition, `pDTurns` and `pDTok` in `internal/thermal/live.go:654-670` do not clamp to `deltaTurns` and `deltaTok` when `turnSeen` and `tokSeen` are true.
- A similar baseline collapse exists on midnight transition in `internal/thermal/live.go:159-165`.

### 2. Speed Metric Integer Truncation
- Problem: In `internal/tui/live.go:586-603`, `formatSpeed` casts float speeds to `int64` before calling `thermal.CompactNumber`:
  ```go
  return fmt.Sprintf("%s step/s", thermal.CompactNumber(int64(turnSpeed)))
  ```
- Consequence: When `turnRate` is 40 step/min, `turnSpeed = 40.0 / 60.0 = 0.667 step/s`. `int64(0.667)` truncates to 0, displaying `0 step/s` right next to `Rate: 40 step/min` and non-zero flame intensity.
- Solution: Format sub-1.0 positive speeds with one decimal place (`fmt.Sprintf("%.1f step/s", turnSpeed)`).

## File & Interface Contracts
- `internal/thermal/live.go`: `LiveTracker.Poll` must accumulate lifetime project tokens and turns consistently in all branches.
- `internal/tui/live.go`: `formatSpeed` must display fractional speeds when `0 < speed < 1.0`.
- `internal/thermal/live_test.go`: Test consecutive idle polls followed by 1-turn increments to ensure project deltas do not re-emit lifetime baselines.
- `internal/tui/live_test.go`: Test `formatSpeed` with `turnSpeed = 0.667` asserting `0.7 step/s`.
