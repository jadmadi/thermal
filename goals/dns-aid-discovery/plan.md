# Plan: Publish DNS for AI Discovery (DNS-AID) records for DNS-based agent discovery

## Execution Steps
- [x] Document the DNS-AID record specification for `thermal.jadmadi.net` in `docs/DNS-AID.md`.
- [x] Define SVCB/HTTPS resource records for `_index._agents.thermal.jadmadi.net` with `alpn=h2,h3` and `port=443`.
- [x] Configure zone DNS records in Cloudflare DNS management with DNSSEC enabled.
- [x] Test DNS resolution with `dig +dnssec HTTPS _index._agents.thermal.jadmadi.net`.
- [x] Commit with goal reference `(goals/dns-aid-discovery/goal.md)`.
