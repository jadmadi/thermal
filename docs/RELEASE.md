# Thermal Release Engineering & Architecture

This document describes the automated release pipeline, licensing compliance packaging, UPX compression invariants, historical release archives, and legacy release retirement procedures for **Thermal** (`github.com/jadmadi/thermal`).

---

## 1. Automated Release Pipeline

Thermal uses a fully automated, continuous-delivery release pipeline powered by **`release-please`** and **GoReleaser**:

```
[Developer Commits] (Conventional Commits: feat:, fix:, chore:)
         │
         ▼
[release-please PR] (Automatic SemVer calculation & CHANGELOG.md generation)
         │  (Merged to main)
         ▼
[GitHub Release Tag] (vx.y.z created by release-please)
         │
         ▼
[.github/workflows/release.yml]
         │
         ├─► Multi-platform Go compilation (linux, darwin, windows)
         ├─► UPX binary compression (linux, windows only)
         ├─► Licensing asset bundling (LICENSE, DUAL-LICENSE.md, NOTICES.md)
         └─► Release asset upload to GitHub Releases
```

### Key Release Invariants
1. **Conventional Commits**: Every commit affecting user features must follow Conventional Commits (`feat:`, `fix:`, `refactor:`, `perf:`, `chore:`). Documentation-only commits (`docs:`) and internal test commits (`test:`) are ignored by release-please to avoid unnecessary version bumps.
2. **License Inclusions**: Every release archive (`.tar.gz` and `.zip`) must package:
   - `LICENSE` (GNU AGPL-3.0)
   - `DUAL-LICENSE.md` (Commercial Enterprise license terms and Community Reciprocity Covenant)
   - `NOTICES.md` (Third-party open-source attributions)
3. **UPX Compression Policy**:
   - **Linux & Windows**: UPX compression (`upx --best`) is applied to minimize download bandwidth and memory footprint.
   - **macOS (Darwin)**: UPX compression is strictly disabled. Stripped macOS binaries must remain uncompressed so macOS Gatekeeper, code signing, and binary format verifiers do not reject the executable.

---

## 2. Historical Release Archive

Before the formal AGPL-3.0 dual-licensing baseline was sealed, the following releases were published:

| Tag | Published Date | Milestone Summary |
| :--- | :--- | :--- |
| **`v0.12.0`** | 2026-09-21 | Replay simulation engine, subscription limit comparison (`--against`, `--compare`), interactive TUI mix/models views |
| **`v0.11.0`** | 2026-09-21 | Disjoint token attribution (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`), Nous Hermes loader |
| **`v0.10.0`** | 2026-09-21 | DeepSeek (DSH) session projcache loader, enhanced pricing catalog sync |
| **`v0.9.1`** | 2026-09-20 | Terminal rendering fixes, color palette tuning |
| **`v0.9.0`** | 2026-09-20 | Charting analytics (`mix`, `stats`, `trend`), Unicode bar graphs |
| **`v0.8.0`** | 2026-09-19 | Project attribution normalization (`ProjectKey`), multi-tool project aggregation |
| **`v0.7.0`** | 2026-09-19 | Strict report flag validation, JSON schema consolidation |
| **`v0.6.0`** | 2026-09-16 | Claude and Droid JSONL transcript loaders |
| **`v0.5.0`** | 2026-09-15 | Grok and ZCode loaders, `models.dev` price catalog integration |
| **`v0.4.0`** | 2026-09-15 | Codex rollout log ingestion and session parsing |
| **`v0.3.0`** | 2026-09-15 | Devin and OpenCode v2 SQLite loaders, memory-mapped I/O |
| **`v0.2.3`** | 2026-09-13 | Performance fixes, memory-mapped SQLite optimization |
| **`v0.2.2`** | 2026-09-12 | Terminal leaderboard formatting and color themes |
| **`v0.2.1`** | 2026-09-11 | Command-code and Agy session loaders |
| **`v0.2.0`** | 2026-07-18 | Initial multi-tool architecture |
| **`v0.1.0`** | 2026-06-12 | Initial prototype release (MIT License) |

---

## 3. Legacy Release Retirement & Reset Procedure

To ensure that distributed binaries reflect the current **GNU AGPL-3.0 and Commercial Dual-Licensing** terms, Thermal maintains a release retirement script:

```bash
# 1. Audit published releases in safe dry-run mode (zero changes):
./scripts/retire_legacy_releases.sh --dry-run

# 2. To permanently retire legacy releases on GitHub (requires explicit authorization):
./scripts/retire_legacy_releases.sh --confirm
```

### Safety Principles:
- **Default Dry-Run**: `retire_legacy_releases.sh` defaults to read-only inspection. It never contacts GitHub or mutates releases without `--confirm`.
- **Historical Integrity**: The commit history of all past tags remains permanently preserved in git. Retiring GitHub release assets only stops distribution of outdated pre-compiled binaries from the releases UI.
- **Continuous Distribution**: Users building from source via `go install github.com/jadmadi/thermal/cmd/thermal@latest` always receive the latest licensed release line.
