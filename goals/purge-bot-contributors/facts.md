# Facts: Purge Devin Bot Contributor Attribution & Git History Rewrite

## Architectural Invariants & Constraints
- Sole Target — Devin Only: History rewrite targets exclusively commits authored or committed by devin-ai-integration[bot]. github-actions[bot] and all other commit authors remain completely unaffected.
- GitHub Profile Attribution: All commits matching devin-ai-integration[bot] are rewritten to Jad Madi <jadmadi@gmail.com> (primary GitHub email) so commits link to Jad Madi's GitHub profile and purge Devin from graphs/contributors.
- Code & Tree Integrity: Exact file contents, tree hashes, commit messages, and commit timestamps are preserved with zero regressions.

## File & Interface Contracts
- Relevant files:
  - `scripts/purge_devin_commits.sh`
  - `.mailmap`
