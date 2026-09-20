# Plan: Shrink the unpriceable token line

## Execution Steps
- [x] **Task 1 — Find the Codex mapping gap** (`internal/loaders/codex.go`): work out why one day loses its model when its rollout log carries `gpt-6-astra`. Check the date the rollup uses against the directory the log lives in, since the 2026-09-14 day resolves from a 2026-09-08 directory. Write a test with a fixture that reproduces it before changing anything.
- [x] **Task 2 — Fix the mapping** so every parsed session contributes its model to the day its tokens land on, and add a regression test that a session whose directory month differs from its token day still attributes.
- [x] **Task 3 — Measure the line** before and after, and record both numbers in the goal report so the change is provable rather than asserted.
- [x] **Task 4 — Decide Cause 2** (decided: an activity-only row is one where Tokens equals Turns and everything else is empty; those stay out of token totals and remain in the leaderboard and streaks): whether activity-only loaders should keep writing their counts into `Tokens`. Options are a marker on `DailyRow` set by the loader, or a narrower rule in `hasTokenData`. Do not act without the decision, because it changes what reports show.
- [x] **Task 5 — Documentation**: the README cost section and the guide's cost section state what the line counts, and AGENTS.md records whatever decision Task 4 reaches.
- [x] **Task 6 — Verification and commit**: `go test -race ./...`, build, and a hand check of the footer figure. Commit with `Goal-Ref: unpriceable-token-sources` under a `fix:` subject, open a pull request into `main`, and do not cut a release from it.
