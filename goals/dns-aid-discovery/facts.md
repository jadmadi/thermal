# Facts: Publish DNS for AI Discovery (DNS-AID) records for DNS-based agent discovery

## Issue
DNS for AI Discovery (DNS-AID) well-known entrypoint records not found.

## Fix & Requirements
Publish DNS for AI Discovery (DNS-AID) records under your domain (e.g. `_index._agents.thermal.jadmadi.net` or `_a2a._agents.thermal.jadmadi.net`) using ServiceMode SVCB/HTTPS records with `alpn` and `endpoint` parameters. Sign the public discovery zone with DNSSEC so validating resolvers return authenticated data.

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/dns-aid/SKILL.md`
- Draft Spec: `https://datatracker.ietf.org/doc/draft-mozleywilliams-dnsop-dnsaid/`
- RFC 9460 (SVCB and HTTPS RR Types): `https://www.rfc-editor.org/rfc/rfc9460`
- Infrastructure: Cloudflare DNS zone management for `jadmadi.net`.
