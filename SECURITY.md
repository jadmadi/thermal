# Security Policy — Thermal (`thermal`)

The Thermal project takes the security, privacy, and integrity of developers' local environments and telemetry data seriously. This document outlines our vulnerability disclosure policy, supported versions, response SLAs, and security design boundaries.

---

## 1. Supported Versions

Security patches and bug fixes are applied to the active release line:

| Version Line | Supported | Notes |
| :--- | :--- | :--- |
| **v0.12.x** | ✅ Yes | Current active release branch |
| < v0.12.0 | ❌ No | Please upgrade to the latest release via `thermal upgrade` or `go install github.com/jadmadi/thermal/cmd/thermal@latest` |

---

## 2. Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub Issues or Pull Requests.**

If you discover or suspect a security vulnerability in Thermal, report it privately through either of the following channels:

1. **Email (Preferred)**: Send an encrypted or plain-text advisory to **[contact@jadmadi.net](mailto:contact@jadmadi.net)** with the subject line `[SECURITY] <Brief Description>`.
2. **GitHub Security Advisory**: Submit a private report via GitHub's [Private Vulnerability Reporting](https://github.com/jadmadi/thermal/security/advisories/new).

### What to Include
To expedite triage, please provide:
- A clear description of the vulnerability and affected component (e.g., specific loader, argument parser, project key resolver).
- Step-by-step instructions to reproduce the issue, including sample config or mock transcript if applicable.
- Potential impact (e.g., path traversal, arbitrary file read, denial of service).
- Any proposed patch, remediation, or suggested workaround.

---

## 3. Vulnerability Response Timeline & SLA

Thermal enforces a strict, time-bounded response commitment for security reports:

- **48 Hours (Initial Triage)**: Acknowledge receipt of the report, confirm communication channels, and begin preliminary verification.
- **7 Days (Severity Assessment)**: Complete impact analysis, determine Common Vulnerability Scoring System (CVSS) rating, and confirm reproducibility.
- **30 Days (Remediation & Coordinated Disclosure)**: Develop, review, and test a security patch. Coordinate release timing and public disclosure (including CVE assignment if applicable).

---

## 4. Security Model & Architectural Invariants

Thermal's design incorporates several foundational security principles:

1. **Strictly Read-Only Ingestion**: Thermal never writes to, modifies, or deletes user databases or transcripts. All SQLite connections use `?mode=ro` URI flags, preventing write operations, table locks, or accidental corruption.
2. **Zero Cloud Telemetry**: Thermal is completely offline-capable. It does not send user data, metrics, tokens, or paths to external cloud servers.
3. **Safe Path Normalization**: `ProjectKey` normalization verifies repository boundaries and resolves symbolic links defensively to prevent directory traversal outside repository trees.
4. **Bounded Concurrency & Ceilings**: JSONL scanners enforce a strict line ceiling (32 MiB) and worker semaphore pools to protect against heap exhaustion or denial of service from malformed log streams.

### Non-Vulnerabilities (Out of Scope)
The following scenarios are considered out of scope for security reports:
- Attacks requiring existing root or administrative privileges on the local machine.
- Man-in-the-middle attacks on third-party pricing catalog downloads when `--offline` flag is not used (catalog fetches use standard HTTPS with TLS certificate validation).
- Intentional local configuration corruption by the user in `~/.cache/thermal/`.
