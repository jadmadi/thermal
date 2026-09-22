# Agent Authentication & Interaction Policy (Auth.md)

This document declares machine-readable agent authentication, access policies, and registration instructions for **Thermal** (`https://thermal.jadmadi.net`).

## 1. Architectural Posture: Open Public Access

Thermal is a local-first, zero-allocation terminal tool and documentation website:
- **No Cloud Telemetry**: Thermal never uploads developer telemetry or token metrics anywhere. All scans execute strictly read-only on the local machine.
- **Public Documentation & Metadata**: All documentation pages, `.well-known` discovery indices, and release assets are public and require **no authentication credentials, bearer tokens, or client registration**.
- **Rate Limits**: Standard Cloudflare DDoS and fair-use protections apply. Machine agents should include an informative `User-Agent` string.

## 2. Agent Registration

Automated agents consuming public Thermal resources do not need prior API registration:

```json
{
  "auth_required": false,
  "agent_registration": "open",
  "supported_identity_types": ["anonymous", "did:web"],
  "documentation": "https://thermal.jadmadi.net/llms-full.txt",
  "contact": "contact@jadmadi.net"
}
```

## 3. Commercial Dual-Licensing & Inquiries

For commercial enterprise licenses, multi-tenant fleet distribution rights, or proprietary embeddings beyond the terms of GNU AGPL-3.0:
- **Contact**: `contact@jadmadi.net`
- **License Terms**: [`DUAL-LICENSE.md`](https://github.com/jadmadi/thermal/blob/main/DUAL-LICENSE.md)
- **Governance**: [`GOVERNANCE.md`](https://github.com/jadmadi/thermal/blob/main/GOVERNANCE.md)

## 4. Protected Resource Metadata

See [`/.well-known/oauth-protected-resource`](/.well-known/oauth-protected-resource) for OAuth 2.0 Protected Resource Metadata (RFC 9728).
