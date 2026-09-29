# Goal: Instant Terminal Statusline & Prompt Composable Helper

## Goal Description
Introduce an ultra-fast (<3ms) thermal statusline (or thermal prompt) command designed for shell prompt composability (starship, zsh, tmux) rendering current streak, daily token burn, and active tool without opening interactive TUIs.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 76
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- cmd/thermal/main.go
- internal/render/statusline.go
- internal/render/statusline_test.go
- docs/MIGRATION.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/terminal-statusline-helper/facts.md`](facts.md)
- **Execution Plan**: [`goals/terminal-statusline-helper/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
