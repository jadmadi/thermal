# Thermal™ Dual Licensing & Architecture

**Thermal™ by Jad Madi — AI Agent Telemetry & Streak Engine**  
*Copyright (C) 2026 Jad Madi. All rights reserved.*

Thermal (`github.com/jadmadi/thermal`) is dual-licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)** and a **Commercial Use License** ($30 per developer per year), with separate **Enterprise & OEM** terms for embedding, hosted services, and support commitments.

---

## 1. Community & Open Source Edition (GNU AGPLv3)

Thermal is free and open-source software under the terms of the **[GNU Affero General Public License v3.0](LICENSE)**.

### What is permitted:
- Using Thermal as a local terminal CLI and TUI on your personal machine or development environment.
- Inspecting, modifying, and customizing Thermal's code for open-source and personal use.
- Distributing copies and modifications under the reciprocal terms of the AGPLv3.

### What is required under AGPLv3:
- **Reciprocal Source Disclosure**: If you distribute modified versions of Thermal (binaries or source), you must make the complete corresponding source code available under AGPLv3.
- **Network Interaction Clause (Section 13)**: If you run a modified version of Thermal that interacts with users remotely over a computer network (such as a hosted web dashboard, SaaS leaderboard, or centralized developer analytics service), you must prominently offer all remote users access to the complete corresponding source code under AGPLv3.
- **Notice Preservation**: You must retain all copyright notices, license headers, and disclaimers.

---

## 2. Commercial Use License ($30 per developer per year)

For organizations that need to use Thermal in commercial work, or whose internal policy restricts AGPL software, without embedding or redistributing it:

- **Internal commercial use** by one named developer for a 12-month term.
- **Use without AGPL-3.0 obligations** for that internal use only.
- **Updates** released during the term.
- **Community support** through GitHub Issues; no SLA.

**Price:** **$30 per developer per 12-month term**. Pay through PayPal or request an invoice; a license certificate follows by email.

**Terms:** [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md).

## 3. Enterprise & OEM License

For organizations, commercial software vendors, and cloud service providers that need more than internal use:

- **Embed Thermal into proprietary products**: Integrate Thermal or its internal parsers/loaders into closed-source commercial software, IDE extensions, or proprietary developer platforms without open-sourcing your product.
- **Run closed-source hosted SaaS / cloud services**: Provide hosted leaderboard, analytics, or team heatmap services powered by Thermal without disclosing your infrastructure or proprietary application source code under AGPL Section 13.
- **Negotiate fleet and security terms**: Enterprise deployment review, custom internal agent loaders, and deployment assistance.
- **Receive support commitments**: Dedicated technical support, response commitments, and custom deployment assistance under a separate agreement.

Enterprise terms are negotiated separately and are not included in the $30 Commercial Use License.

---

## 4. Community Reciprocity Covenant

To foster genuine contributor trust and eliminate the fear of unilateral corporate relicensing (e.g., historical transitions to BUSL, SSPL, or closed source seen across other dual-licensed ecosystems), Thermal is bound by this **Community Reciprocity Covenant**:

1. **Irrevocable Open-Source Guarantee**: The complete Thermal CLI binary, all tool-specific ingestion loaders (`internal/loaders/`), the analytics and streak engine (`internal/thermal/`), and the terminal TUI (`internal/tui/`) will remain available under an OSI-approved copyleft license (GNU AGPL-3.0) in perpetuity.
2. **No Retroactive Proprietary Relicensing**: Community contributions submitted under AGPL-3.0 will never be retroactively relicensed into a closed-source, source-available, or non-OSI proprietary product without explicit contributor agreement.
3. **Demarcation of Commercial Offerings**: Commercial licensing offered by Jad Madi is strictly demarcated to:
   - Internal commercial use and corporate compliance exemptions, available as the **$30 per developer per year Commercial Use License**.
   - Proprietary embedding into closed-source commercial IDEs or vendor platforms.
   - Closed-source multi-tenant enterprise fleet hosting and private SaaS dashboards.
   - Dedicated commercial support SLAs and custom private loader development.
4. **No Artificial Feature Gating**: Thermal will not cripple or artificially withhold local terminal features, loaders, performance optimizations, or analytics from the open-source community edition.

---

## 5. Commercial Licensing Inquiries

**Commercial Use License — $30 per developer per year**
* **Terms**: [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md)
* **Buy one developer**: [PayPal $30](https://paypal.me/Madise/30)
* **Invoice or multiple developers**: [contact@jadmadi.net](mailto:contact@jadmadi.net)

**Enterprise & OEM**
To obtain an enterprise quote, evaluate a fleet deployment, or discuss OEM/embedding terms:
* **Contact**: [contact@jadmadi.net](mailto:contact@jadmadi.net)
* **GitHub**: [@jadmadi](https://github.com/jadmadi)
* **Website**: [https://github.com/jadmadi/thermal](https://github.com/jadmadi/thermal)

---

## 6. Trademark & Forking Guidelines

### Forking Rights
In accordance with open-source philosophy and Karl Fogel's *Producing Open Source Software*, the credibility of open-source governance rests on the genuine right to fork. Any developer or organization may fork Thermal under the terms of the GNU AGPLv3:
- You may copy, modify, and distribute the source code under AGPL-3.0.
- You must retain all copyright notices, commit history, and original author attribution in documentation and source headers.

### Trademark & Naming Protection
To avoid confusion in package managers and developer communities:
1. **Distinct Identity**: Any fork or derivative distribution must use an independent, distinct project name (e.g., *AgentHeat*, not *Thermal-Fork* or *Thermal-NG*).
2. **Marks Removal**: You must remove the **Thermal™** trademark, flame logo assets, and the tagline **"Don't break the streak"** from the derivative product's binary output, repository branding, and marketing copy.
3. **Fair Attribution**: You may accurately and non-misleadingly state that your work is "Based on Thermal by Jad Madi" or "Derived from github.com/jadmadi/thermal under AGPL-3.0".

---

## 7. Repository Asset Scope & Licensing Boundaries

To maintain legal certainty across a mixed repository, the contents of `github.com/jadmadi/thermal` are licensed as follows:

1. **Software & CLI Source Code** (`cmd/`, `internal/`, `build.sh`): Licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)** for open-source community use, or under the **Commercial Use License** ($30 per developer per year) for internal commercial use, or under a separate **Enterprise & OEM License** for proprietary deployments.
2. **Documentation & Guides** (`docs/`, `*.md`): Licensed under the **Creative Commons Attribution 4.0 International (CC-BY-4.0)** license. You may share and adapt documentation provided appropriate attribution is preserved.
3. **Visual Branding, Screenshots & Logos** (`assets/*.webp`, `docs/pages/*.webp`, `docs/pages/og.png`, branding assets): **All Rights Reserved by Jad Madi**. Visual assets, screenshots, and logos represent proprietary branding and are excluded from open-source code and content licenses.

---

## 8. Licensing Architecture & Public Decision Record

This section records the architectural rationale, constraints, and eliminated alternatives for Thermal's licensing model:

- **Why GNU AGPLv3?** Thermal is an AI agent telemetry and streak engine. Developer analytics engines and session ingestion pipelines are frequent targets for closed-source, cloud-hosted SaaS exploitation (e.g., proprietary hosted team dashboards or platform extensions). AGPL Section 13 prevents cloud providers and competitors from wrapping Thermal into hosted services without contributing their improvements back to the community.
- **Why Dual Licensing?** To enable adoption in organizations whose internal Open Source Program Office (OSPO) policies restrict AGPL software, with a low-friction **$30 per developer per year** Commercial Use License for internal commercial use, and a separate Enterprise & OEM track for vendors embedding Thermal into proprietary IDEs, platforms, or hosted services.
- **Why Alternatives Were Eliminated**:
  - *Permissive (MIT / Apache-2.0)*: Eliminated for core releases beyond v0.1.0 because it allows proprietary SaaS wrappers to monetize Thermal's zero-allocation parsers and telemetry engine with zero reciprocity or contribution back to the maintainer.
  - *Source-Available (BUSL, SSPL, FSL)*: Eliminated because they are not OSI-approved open source. They provoke community backlash, fail enterprise procurement allowlists, and prevent packaging in standard Linux/BSD package ecosystems.
  - *Standard GPLv3*: Eliminated because it only triggers upon binary conveyance, leaving the hosted cloud/SaaS loophole unaddressed.
- **Version Boundaries & Past History**: Releases up to and including `v0.1.0` permanently remain under the MIT License for those who obtained them. Beginning with `v0.1.1` (commit `aa9e241`), all subsequent releases, commits, and contributions are governed by GNU AGPLv3 with Commercial Dual-Licensing.
