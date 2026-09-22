# Thermal™ Dual Licensing & Architecture

**Thermal™ by Jad Madi — AI Agent Telemetry & Streak Engine**  
*Copyright (C) 2026 Jad Madi. All rights reserved.*

Thermal (`github.com/jadmadi/thermal`) is dual-licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)** and a **Commercial Enterprise License**.

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

## 2. Commercial & Enterprise Edition

For organizations, commercial software vendors, and cloud service providers who cannot or choose not to comply with AGPLv3, Jad Madi offers a **Commercial Enterprise License**.

### When do you need a Commercial License?
You require a commercial license if you wish to:
- **Embed Thermal into Proprietary Products**: Integrate Thermal or its internal parsers/loaders into closed-source commercial software, IDE extensions, or proprietary developer platforms without open-sourcing your product.
- **Run Closed-Source Hosted SaaS / Cloud Services**: Provide hosted leaderboard, analytics, or team heatmap services powered by Thermal without disclosing your infrastructure or proprietary application source code under AGPL Section 13.
- **Satisfy Corporate OSPO Policies**: Deploy Thermal across enterprise developer workstations where internal corporate policies prohibit AGPL-licensed software.
- **Receive Enterprise Support & SLAs**: Access dedicated technical support, priority loader requests for custom internal agent harnesses, and custom deployment assistance.

---

---

## 3. Community Reciprocity Covenant

To foster genuine contributor trust and eliminate the fear of unilateral corporate relicensing (e.g., historical transitions to BUSL, SSPL, or closed source seen across other dual-licensed ecosystems), Thermal is bound by this **Community Reciprocity Covenant**:

1. **Irrevocable Open-Source Guarantee**: The complete Thermal CLI binary, all tool-specific ingestion loaders (`internal/loaders/`), the analytics and streak engine (`internal/thermal/`), and the terminal TUI (`internal/tui/`) will remain available under an OSI-approved copyleft license (GNU AGPL-3.0) in perpetuity.
2. **No Retroactive Proprietary Relicensing**: Community contributions submitted under AGPL-3.0 will never be retroactively relicensed into a closed-source, source-available, or non-OSI proprietary product without explicit contributor agreement.
3. **Demarcation of Commercial Offerings**: Commercial enterprise licensing offered by Jad Madi is strictly demarcated to:
   - Proprietary embedding into closed-source commercial IDEs or vendor platforms.
   - Closed-source multi-tenant enterprise fleet hosting and private SaaS dashboards.
   - Enterprise corporate compliance exemptions for organizations with blanket internal AGPL workstation bans.
   - Dedicated commercial support SLAs and custom private loader development.
4. **No Artificial Feature Gating**: Thermal will not cripple or artificially withhold local terminal features, loaders, performance optimizations, or analytics from the open-source community edition.

---

## 4. Commercial Licensing Inquiries

To obtain a commercial quote, evaluate an enterprise deployment, or discuss OEM/embedding terms:

* **Contact**: [contact@jadmadi.net](mailto:contact@jadmadi.net)
* **GitHub**: [@jadmadi](https://github.com/jadmadi)
* **Website**: [https://github.com/jadmadi/thermal](https://github.com/jadmadi/thermal)

---

## 5. Trademark & Forking Guidelines

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

## 6. Repository Asset Scope & Licensing Boundaries

To maintain legal certainty across a mixed repository, the contents of `github.com/jadmadi/thermal` are licensed as follows:

1. **Software & CLI Source Code** (`cmd/`, `internal/`, `build.sh`): Licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)** for open-source community use, or under the **Commercial Enterprise License** for proprietary deployments.
2. **Documentation & Guides** (`docs/`, `*.md`): Licensed under the **Creative Commons Attribution 4.0 International (CC-BY-4.0)** license. You may share and adapt documentation provided appropriate attribution is preserved.
3. **Visual Branding, Screenshots & Logos** (`docs/pages/*.webp`, `docs/pages/og.png`, root `*.webp`, branding assets): **All Rights Reserved by Jad Madi**. Visual assets, screenshots, and logos represent proprietary branding and are excluded from open-source code and content licenses.

---

## 7. Licensing Architecture & Public Decision Record

This section records the architectural rationale, constraints, and eliminated alternatives for Thermal's licensing model:

- **Why GNU AGPLv3?** Thermal is an AI agent telemetry and streak engine. Developer analytics engines and session ingestion pipelines are frequent targets for closed-source, cloud-hosted SaaS exploitation (e.g., proprietary hosted team dashboards or platform extensions). AGPL Section 13 prevents cloud providers and competitors from wrapping Thermal into hosted services without contributing their improvements back to the community.
- **Why Dual Licensing?** To enable enterprise adoption in organizations whose internal Open Source Program Office (OSPO) policies maintain written bans on AGPL software (e.g., Google's corporate AGPL ban on developer workstations), and to provide a commercial vehicle for vendors embedding Thermal into proprietary IDEs or platforms.
- **Why Alternatives Were Eliminated**:
  - *Permissive (MIT / Apache-2.0)*: Eliminated for core releases beyond v0.1.0 because it allows proprietary SaaS wrappers to monetize Thermal's zero-allocation parsers and telemetry engine with zero reciprocity or contribution back to the maintainer.
  - *Source-Available (BUSL, SSPL, FSL)*: Eliminated because they are not OSI-approved open source. They provoke community backlash, fail enterprise procurement allowlists, and prevent packaging in standard Linux/BSD package ecosystems.
  - *Standard GPLv3*: Eliminated because it only triggers upon binary conveyance, leaving the hosted cloud/SaaS loophole unaddressed.
- **Version Boundaries & Past History**: Releases up to and including `v0.1.0` permanently remain under the MIT License for those who obtained them. Beginning with `v0.1.1` (commit `aa9e241`), all subsequent releases, commits, and contributions are governed by GNU AGPLv3 with Commercial Dual-Licensing.
