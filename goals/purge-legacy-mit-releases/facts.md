# Facts: Retire Legacy MIT Releases and Reset Release Line to AGPL-3.0

## Background & Legal Invariants
1. **Irrevocability of Prior MIT Grants**: Copies of binaries or source code downloaded by third parties prior to the license change remain governed by the MIT license under which they were obtained. Open-source licenses cannot be retroactively cancelled for prior recipients.
2. **Distribution Control & Brand Cleanliness**: As copyright owner, Jad Madi possesses the legal right to stop distributing MIT-licensed binaries and assets on the official GitHub repository (`github.com/jadmadi/thermal`). Deleting legacy GitHub releases prevents future downloads of MIT-labeled binaries and stops wrapper startups from pointing to official MIT binaries.
3. **Dual-License Baseline**: Starting with the next release (e.g. `v0.8.0`), all official binaries and releases will be strictly dual-licensed under AGPL-3.0 (Community Edition) and the Commercial Enterprise License.
4. **Safety & Non-Destructive Archiving**: Before deleting remote release assets or tags from GitHub, an archive manifest of past release notes, tags, and commits must be recorded locally in `docs/RELEASE.md` so historical changelog records are preserved.

## Relevant Files & Tools
- `scripts/retire_legacy_releases.sh`: Automation script with `--dry-run` and `--confirm` safety checks wrapping GitHub CLI (`gh release list`, `gh release delete`, `git push origin --delete`).
- `docs/RELEASE.md`: Documentation outlining release lifecycle, legacy release retirement record, and future release procedure.
- GitHub CLI (`gh`): Required for programmatic release and asset deletion.
