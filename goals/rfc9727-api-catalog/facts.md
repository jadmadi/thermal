# Facts: Publish an API catalog for automated API discovery (RFC 9727)

## Issue
API Catalog not found.

## Fix & Requirements
Create `docs/pages/.well-known/api-catalog` returning `application/linkset+json` with a `"linkset"` array. Each entry should include an `"anchor"` URL for the API and link relations for `service-desc` (OpenAPI spec), `service-doc` (documentation), and `status` (health endpoint). See RFC 9727 Appendix A for examples.

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/api-catalog/SKILL.md`
- RFC 9727: `https://www.rfc-editor.org/rfc/rfc9727`
- RFC 9264: `https://www.rfc-editor.org/rfc/rfc9264`
