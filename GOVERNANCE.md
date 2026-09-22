# Project Governance — Thermal (`thermal`)

Thermal is an open-source terminal contribution heatmap and leaderboard tool for AI coding assistants, created and maintained by **Jad Madi** ([@jadmadi](https://github.com/jadmadi)).

This document defines Thermal's governance model, core values, architectural scope boundaries, decision-making mechanisms, review timers, and conflict resolution processes.

---

## 1. Governance Model: Written Solo Authority

Thermal operates under a **Written Solo Authority** model, informed by Nadia Eghbal's open-source project taxonomy (*Working in Public*):

- **Stadium Architecture**: Thermal serves a broad user base of developers running diverse AI tools, while maintaining high architectural coherence and a low maintainer overhead (< 30 minutes/month of administrative overhead).
- **Anti-Pattern Rejection ("No Governance Theatre")**: Thermal explicitly rejects complex multi-member elected steering committees or synthetic consensus bodies. For single-maintainer or high-focus developer utilities, bureaucratic governance introduces friction and stalls execution without improving software quality.
- **Transparency & Predictability**: In place of informal benevolent dictatorship, Thermal adopts a formal, written governance contract specifying explicit decision classes, review timeframes, justified veto rules, and defined project scope.

---

## 2. Mission, Scope & Anti-Goals

The primary responsibility of governance in Thermal is to defend maintainer attention and protect the software's architectural integrity from scope creep.

### In-Scope

Thermal focuses strictly on high-performance, terminal-first telemetry and visualization:

1. **Terminal Heatmaps & Dashboards**: Zero-allocation terminal heatmaps, ranking leaderboards, and period activity reports (`daily`, `weekly`, `monthly`, `projects`, `models`, `mix`, `stats`, `trend`, `replay`).
2. **Local-First, Strictly Read-Only Ingestion**: Reading local SQLite databases and session transcripts directly on the developer's workstation. All SQLite connections must open in read-only mode (`?mode=ro`) with memory-mapped I/O (`PRAGMA mmap_size`). Thermal **never** modifies user databases.
3. **Sub-10ms Invocation via Incremental Caching**: Utilizing delta-caching (`~/.cache/thermal/`) based on file modification times and transaction counters to ensure fast CLI responsiveness.
4. **Bounded Concurrency**: Ingestion engines scanning directories or multi-gigabyte log trees must use bounded worker pools (semaphore channels capped at ~16 workers) to ensure deterministic resource consumption.
5. **Canonical Normalization**: Normalized model identifiers (`modelName()`), disjoint token accounting (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`), and git-root project attribution (`ProjectKey`).
6. **Accurate Cost Estimation**: Catalog pricing lookups via `models.dev` that preserve exact recorded costs and clearly disclose estimates (`~$X.XX`) without fuzzy degradation.

### Explicit Anti-Goals (Out of Scope)

Proposals introducing any of the following are considered out of scope and will be rejected:

1. **Cloud-Hosted Telemetry & "Phone-Home" Analytics**: Thermal is 100% offline-capable and privacy-respecting. It does not phone home, report anonymous usage statistics, or require external cloud accounts.
2. **Background Daemons & Cron Services**: Thermal is an on-demand CLI tool. It does not install persistent background runners, system daemons, or recurring listener processes.
3. **Bloated Web/GUI Frameworks**: Thermal is a terminal-native tool. Embedding Electron, Tauri, or browser runtimes into the core binary is out of scope. (Documentation and web visualizations are maintained separately as static HTML in `docs/pages/`).
4. **Non-Local Network Scans**: Scanning remote servers, cloud agent APIs, or non-local storage during standard CLI execution is prohibited.
5. **Invasive Instrumentation**: Thermal reads existing logs and session databases; it does not inject hooks, wrappers, or interception layers into third-party AI tools.
6. **Performance Degradation**: Any feature that compromises Thermal's sub-10ms cached execution ceiling or introduces unbounded memory allocations will not be accepted.

---

## 3. Roles & Responsibilities

### Project Lead (Benevolent Dictator / Solo Authority)
**Jad Madi** ([@jadmadi](https://github.com/jadmadi)) is the founder and Project Lead:
- Sets the architectural vision, roadmap, and design philosophy.
- Holds sole copyright and serves as steward of Thermal's dual-licensing structure (AGPL-3.0 and Commercial Enterprise).
- Holds merge authority over the repository and serves as final tiebreaker in case of consensus deadlock.
- Appoints maintainers and delegates review responsibilities.

### Maintainers
Maintainers are trusted community members with triage or write permissions:
- Review pull requests against `AGENTS.md` and Thermal coding standards.
- Triage issues, reproduce bug reports, and assist contributors.
- Enforce Contributor License Agreement (CLA) compliance and code hygiene.

### Contributors
Anyone in the community who submits pull requests, files bug reports, proposes loader improvements, or enhances documentation:
- Must accept the Contributor License Agreement (CLA) via git `Signed-off-by:` trailers.
- Follow the guidelines in `CONTRIBUTING.md` and `AGENTS.md`.

---

## 4. Decision Classes & Review Timers

To prevent unbounded stalls or silent abandonment, every decision class operates on an explicit clock:

| Decision Class | Scope & Criteria | Review Window | Mechanism |
| :--- | :--- | :--- | :--- |
| **Class 1: Routine Fixes & Docs** | Bug fixes, typo corrections, documentation updates, and CI improvements that do not alter public APIs or data models. | **72 hours** | **Lazy Consensus**: 1 maintainer approval. If no maintainer objects within 72 hours, the PR is merged. |
| **Class 2: New Agent Loaders** | Adding ingestion engines for new AI coding tools or agent frameworks. | **7 days** | Must satisfy the **`AGENTS.md` Loader Checklist** (read-only SQLite, bounded concurrency, canonical models, disjoint tokens, test fixtures). Evaluated within 7 days. |
| **Class 3: Architectural & Breaking Changes** | Changes affecting CLI flags, JSON output schemas, caching formats, or performance invariants. | **7 days** | Requires an RFC issue describing the rationale and alternatives. 7-day public comment window; requires Project Lead approval. |
| **Class 4: Licensing & Governance** | Amendments to `LICENSE`, `DUAL-LICENSE.md`, `GOVERNANCE.md`, or the CLA. | — | Sole authority of the Project Lead (**Jad Madi**). |

---

## 5. Conflict Resolution & Justified Vetoes

Healthy disagreement is normal in open-source development. To keep technical debates productive and rational, Thermal enforces the **Justified Veto** rule (adopted from the Apache Software Foundation):

1. **Technical Rationale Required**: A negative review or veto (`-1`) cannot be arbitrary or personal. Every veto **must** cite a concrete technical justification, such as:
   - Violation of performance invariants (e.g., unbounded goroutines, memory regressions).
   - Violation of read-only safety rules (e.g., write locks on user databases).
   - Scope violations (falling into Anti-Goals).
   - Test suite failures or missing edge-case coverage.
2. **Actionable Remediation**: A veto must suggest an alternative approach or clear remediation path.
3. **Escalation & Tiebreaking**: If consensus cannot be achieved through discussion within 7 days, the Project Lead serves as the final tiebreaker.

---

## 6. Maintainer Availability & Maintenance SLA

To ensure healthy expectations across the open-source community:

- **Weekly Availability**: The Project Lead allocates approximately **2 to 4 hours per week** to Thermal issue triage, pull request review, and release management.
- **Capacity Recognition**: Slower response times during high-volume periods reflect bounded maintainer bandwidth, not project abandonment.
- **Communication Channels**:
  - Public technical inquiries, bug reports, and proposals: [GitHub Issues](https://github.com/jadmadi/thermal/issues) and [Pull Requests](https://github.com/jadmadi/thermal/pulls).
  - Security vulnerabilities: Report privately via email to [contact@jadmadi.net](mailto:contact@jadmadi.net) (see also [SECURITY.md](SECURITY.md)).
  - Commercial and enterprise licensing inquiries: [contact@jadmadi.net](mailto:contact@jadmadi.net).

---

## 7. Staged Trust Ladder & Post-xz Separation of Powers

In light of software supply-chain vulnerabilities such as the 2024 `xz-utils` compromise (CVE-2024-3094), Thermal establishes a strict **Separation of Powers** between code contribution, review rights, commit access, and cryptographic release signing. Commit rights never automatically confer release publishing authority.

Thermal recognizes five progressive stages of trust:

| Stage | Role | Responsibilities | Requirements & Progression Criteria | Access Granted |
| :--- | :--- | :--- | :--- | :--- |
| **Stage 1** | **Contributor** | Submits bug reports, documentation fixes, loader improvements. | Accept CLA via `Signed-off-by:` trailers; pass all automated test gates. | Standard PR submission. |
| **Stage 2** | **Triager** | Reproduces bugs, labels issues, closes stale duplicates. | $\ge 3$ verified, high-quality bug reproductions or issue triage actions. | Issue assignment, label management, PR triage. |
| **Stage 3** | **Reviewer** | Reviews PRs in designated subsystems (`loaders/`, `render/`). | $\ge 10$ substantive, technically sound PR reviews adhering to `AGENTS.md`. | PR review assignment, review approvals. |
| **Stage 4** | **Maintainer** | Merges approved PRs, coordinates feature branches. | $\ge 3$ months of sustained contribution and unanimous consensus of existing maintainers. | Repository write/merge permissions to `main`. |
| **Stage 5** | **Key Holder** | Signs release tags, manages GoReleaser secrets, publishes packages. | $\ge 6$ months active as Maintainer **and** completion of $\ge 1$ supervised release cycle. | GitHub Actions release secrets, domain administration. |

---

## 8. Inactivity, Emeritus Policy & Planned Hiatus

To maintain an accurate security boundary and avoid dormant accounts holding privileged access:

1. **Inactivity Window**: A maintainer who has not performed any code reviews, merged PRs, or participated in issue triage for **6 consecutive months** is deemed inactive.
2. **Emeritus Transition**: Following a 14-day grace notification, inactive maintainers transition to **Emeritus Maintainer** status. Repository write permissions and secret access are de-provisioned, and their historic contributions are permanently honored in `MAINTAINERS.md`.
3. **Credential Rotation SLA**: Any transition from Maintainer or Key Holder triggers an automated revocation and secret rotation within **30 days**.
4. **Planned Hiatus**: Maintainers may declare a planned hiatus of up to 12 months by notifying the Project Lead, temporarily suspending review expectations while retaining basic standing.
5. **Reinstatement**: An Emeritus maintainer may request reinstatement by demonstrating refreshed familiarity with current architectural standards, subject to Project Lead approval.

---

## 9. Emergency Succession & Continuity Protocol

Thermal is designed to survive individual maintainer absence without catastrophic disruption:

1. **Unreachable Lead Trigger**: If the Project Lead (**Jad Madi**) becomes incapacitated, unreachable, or unresponsive across all contact channels for more than **90 consecutive days**:
   - The active Maintainers group convenes to elect an **Interim Project Steward** by majority consensus to maintain critical security patches.
   - If no active maintainers exist, the repository defaults to a structured read-only archive state to preserve existing code and documentation integrity.
2. **Community Continuity**: Because Thermal is permanently copylefted under the **GNU AGPL-3.0** with the **Community Reciprocity Covenant**, the community retains the irrevocable legal right to fork and continue maintenance without licensing ambiguity.
