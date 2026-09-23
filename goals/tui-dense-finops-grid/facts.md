# Facts: Multi-Dimensional FinOps TUI Grid, Activity Taxonomy & Sub-Tool Decomposition

## Architectural Invariants & Constraints
- Clean-Room Implementation: 100% written from scratch in Go using Charmbracelet Bubble Tea & Lip Gloss. Zero code or styling assets copied from external projects. Retains 100% sole copyright by Jad Madi.
- Adaptive Terminal Layout: Dynamically adapts to terminal dimensions (150x40+ full 9-box grid, downsampling gracefully to 2-column or stacked cards on smaller viewports).
- Zero-Allocation Classification: Activity taxonomy and shell command extraction must run during streaming log ingestion without unbounded string allocations.

## File & Interface Contracts
- Relevant files:
  - `cmd/thermal/main.go`
  - `internal/tui/dense.go`
  - `internal/tui/dense_test.go`
  - `internal/thermal/taxonomy.go`
