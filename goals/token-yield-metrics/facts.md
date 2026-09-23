# Facts: Token Yield & Code Output Delta Telemetry ('thermal yield')

## Architectural Invariants & Constraints
- Clean-Room Implementation: Natively implemented in Go. Zero third-party code reuse. 100% sole copyright by Jad Madi.
- Sanitized Metadata Only: Yield metrics compute only quantitative numeric diffs (+lines, -lines, files touched); raw source code and diff contents are discarded at the parser boundary.
- Defensive Division by Zero: Net lines calculation handles zero-delta sessions (pure exploration/debugging) gracefully with infinity/unmeasured badges rather than division errors.

## File & Interface Contracts
- Relevant files:
  - `cmd/thermal/main.go`
  - `internal/thermal/yield.go`
  - `internal/thermal/yield_test.go`
  - `internal/render/yield.go`
