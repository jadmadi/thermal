# Project Maintainers — Thermal (`thermal`)

This document lists the active maintainers, project roles, organizational affiliations, and emeritus contributors for **Thermal**.

Thermal operates under a **Written Solo Authority** model led by Jad Madi. Governance mechanisms, review windows, and maintainer responsibilities are codified in [GOVERNANCE.md](GOVERNANCE.md).

---

## 1. Project Lead & Benevolent Dictator

| Name | GitHub | Role | Affiliation | Contact |
| :--- | :--- | :--- | :--- | :--- |
| **Jad Madi** | [@jadmadi](https://github.com/jadmadi) | Project Lead, Founder & Primary Maintainer | Independent | `jadmadi@gmail.com`<br>`contact@jadmadi.net` |

**Responsibilities**:
- Final architectural direction, roadmap decisions, and release tagging.
- Sole copyright holder and steward of dual-licensing (AGPL-3.0 and Commercial Enterprise).
- Final tiebreaker for RFCs and technical disagreements under the Justified Veto rule.
- Security vulnerability coordination and CVE disclosure management.

---

## 2. Active Maintainers & Reviewers

Maintainers have triage and review permissions. They review PRs, evaluate new tool loaders against [AGENTS.md](AGENTS.md), and help shepherd community contributions.

*Currently, all core review and merge operations are managed directly by Project Lead Jad Madi.*

Community contributors demonstrating sustained, high-quality contributions, deep familiarity with Thermal's zero-allocation performance invariants, and alignment with [GOVERNANCE.md](GOVERNANCE.md) may be nominated for maintainer status by the Project Lead.

---

## 3. Organizational Affiliations & Independence

Thermal is an independent open-source project. Maintainers participate in their individual capacities. Thermal is not owned by, sponsored by, or beholden to any single AI tooling vendor or cloud platform. This independence ensures neutral, unbiased telemetry and fair representation across all supported AI coding agents.

---

## 4. Maintainer Progression & Emeritus Status

- **Becoming a Maintainer**: Detailed in [GOVERNANCE.md](GOVERNANCE.md). Requirements include consistent contributions across loaders, core algorithms, or docs, plus adherence to review SLAs (72-hour lazy consensus).
- **Emeritus Status**: Maintainers who step down or become inactive for more than 6 months transition to Emeritus status with public appreciation and revocation of repository write access.

### Emeritus Maintainers
*None currently.*

---

## 5. Privileged Access & Credential Inventory

In accordance with post-xz supply chain security principles and [GOVERNANCE.md](GOVERNANCE.md) Section 7, repository credentials and distribution secrets are strictly managed:

| Subsystem / Resource | Access Level | Authorized Roles | Current Key Holders |
| :--- | :--- | :--- | :--- |
| **GitHub Repository (`main`)** | Push / Merge | Maintainers, Project Lead | `@jadmadi` |
| **GitHub Release Secrets** | GoReleaser Token | Key Holders, Project Lead | `@jadmadi` |
| **Domain & Hosting DNS** | Administrative | Project Lead | `@jadmadi` |
| **Package Registries** | Publishing | Key Holders, Project Lead | `@jadmadi` |

Any revocation or maintainer transition requires secret rotation within 30 days.

