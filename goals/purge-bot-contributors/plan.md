# Plan: Purge Devin Bot Contributor Attribution & Git History Rewrite

## Execution Steps
- [x] Phase 1 — Devin Commit Audit: Audit the git history to identify exact commit SHAs where devin-ai-integration[bot] is the author or committer.
- [x] Phase 2 — History Rewrite Script: Implement deterministic script (scripts/purge_devin_commits.sh) rewriting devin-ai-integration[bot] to Jad Madi <jadmadi@gmail.com> using git-filter-repo or native rebase.
- [x] Phase 3 — Verification & Push Instructions: Verify tree hashes and test suite (go test -race ./...), then provide exact git push --force-with-lease origin main command to refresh GitHub's contributor graph.
