# DNS for AI Discovery (DNS-AID) Architecture & Zone Records

This document defines the **DNS-AID** (DNS for AI Discovery) specification, zone records, and DNSSEC validation for **Thermal** (`thermal.jadmadi.net`) per `draft-mozleywilliams-dnsop-dnsaid` and RFC 9460.

---

## 1. Protocol Purpose & Discovery Semantics

DNS-AID enables autonomous AI agents to discover machine endpoints, API catalogs, and agent-to-agent (A2A) interfaces via standard DNS resolution, bypassing initial HTTP overhead.

Using ServiceMode `HTTPS` and `SVCB` resource records, validating resolvers query standardized DNS subdomains to retrieve endpoint URLs, ALPN protocol capabilities, and cryptographic authenticity (via DNSSEC).

---

## 2. Configured Resource Records (`jadmadi.net`)

Under the authoritative DNS zone managed by Cloudflare for `jadmadi.net`:

### A. Primary Agent Index (`_index._agents.thermal.jadmadi.net`)
```dns
; Primary discovery record pointing to Thermal agent entrypoint
_index._agents.thermal.jadmadi.net.  300  IN  HTTPS  1  thermal.jadmadi.net. (
    alpn="h2,h3"
    port="443"
    mandatory="alpn"
)

; TXT fallback and capability attributes
_index._agents.thermal.jadmadi.net.  300  IN  TXT  "v=dns-aid1; api-catalog=https://thermal.jadmadi.net/.well-known/api-catalog; mcp=https://thermal.jadmadi.net/.well-known/mcp/server-card.json"
```

### B. Agent-to-Agent Service Record (`_a2a._agents.thermal.jadmadi.net`)
```dns
; A2A communication endpoint
_a2a._agents.thermal.jadmadi.net.    300  IN  HTTPS  1  thermal.jadmadi.net. (
    alpn="h2,h3"
    port="443"
)
_a2a._agents.thermal.jadmadi.net.    300  IN  TXT  "v=dns-aid1; type=cli-telemetry-engine; skills=https://thermal.jadmadi.net/.well-known/agent-skills/index.json"
```

---

## 3. DNSSEC Verification Runbook

All records within `jadmadi.net` are cryptographically signed using Cloudflare DNSSEC (Algorithm 13 - ECDSAP256SHA256).

### Verification Command:
```bash
# Query authenticated HTTPS record with DNSSEC flag (+dnssec)
dig +dnssec HTTPS _index._agents.thermal.jadmadi.net

# Verify AD (Authenticated Data) flag in response header:
# ;; flags: qr rd ra ad; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 1
```

---

## 4. References
- IETF Draft: [`draft-mozleywilliams-dnsop-dnsaid`](https://datatracker.ietf.org/doc/draft-mozleywilliams-dnsop-dnsaid/)
- RFC 9460: *Service Binding and Parameter Specification via the DNS (SVCB and HTTPS RRs)*
