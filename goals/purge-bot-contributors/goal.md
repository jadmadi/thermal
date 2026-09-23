# Goal: Purge Devin Bot Contributor Attribution & Git History Rewrite

## Goal Description
Eliminate devin-ai-integration[bot] from GitHub's contributor graph. Audit and rewrite git commit author/committer history exclusively for commits authored or committed by devin-ai-integration[bot], re-attributing them to Jad Madi <jadmadi@gmail.com> (the default GitHub account email). All other contributors (including github-actions[bot]) remain untouched.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: immediate

## Files to Touch
- scripts/purge_devin_commits.sh
- .mailmap

## References
- **Shared Understanding & Fact Sheet**: [`goals/purge-bot-contributors/facts.md`](facts.md)
- **Execution Plan**: [`goals/purge-bot-contributors/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
