# Facts: Declare AI content usage preferences with Content Signals in robots.txt

## Issue
No Content Signals found in `robots.txt`.

## Fix & Requirements
Add `Content-Signal` directives to `docs/pages/robots.txt` declaring preferences for `ai-train`, `search`, and `ai-input`. For example:
```
User-agent: *
Allow: /

Content-Signal: ai-train=no, search=yes, ai-input=no
```

## Architectural Context & References
- Skill: `https://isitagentready.com/.well-known/agent-skills/content-signals/SKILL.md`
- Content Signals: `https://contentsignals.org/`
- IETF Draft: `https://datatracker.ietf.org/doc/draft-romm-aipref-contentsignals/`
