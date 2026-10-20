#!/usr/bin/env bash
#
# output_gallery_fixture.sh — hermetic fixture for the reproducible output gallery.
#
# Prints the fixture HOME on stdout. The caller owns cleanup. Data is fixed
# (2026-01-15) so report totals do not drift between gallery sections.

set -euo pipefail

GALLERY_HOME="$(mktemp -d -t thermal-gallery-home-XXXXXX)"
SESSION_DATE="2026-01-15T12:00:00Z"
REPO_A="${GALLERY_HOME}/projects/repo-a"
REPO_B="${GALLERY_HOME}/projects/repo-b"
WORK="${GALLERY_HOME}/work"

mkdir -p "${REPO_A}/.git" "${REPO_B}/.git" "${WORK}/.git"
mkdir -p "${GALLERY_HOME}/.codewhale/sessions"
mkdir -p "${GALLERY_HOME}/.claude/projects/gallery"
mkdir -p "${GALLERY_HOME}/.claude"
mkdir -p "${GALLERY_HOME}/.cache/thermal"
mkdir -p "${GALLERY_HOME}/.gemini/antigravity/brain/gallery/.system_generated/logs"

cat > "${GALLERY_HOME}/.codewhale/sessions/session_gallery.json" <<EOF
{
  "session_id": "gallery-sess-1",
  "metadata": {
    "created_at": "${SESSION_DATE}",
    "updated_at": "${SESSION_DATE}",
    "message_count": 10,
    "total_tokens": 750000,
    "cost": { "session_cost_usd": 1.75 },
    "model": "claude-3-5-sonnet",
    "mode": "chat",
    "workspace": "${REPO_A}"
  }
}
EOF

cat > "${GALLERY_HOME}/.claude/projects/gallery/session.jsonl" <<EOF
{"type":"assistant","timestamp":"${SESSION_DATE}","cwd":"${REPO_B}","message":{"id":"gallery-msg-1","model":"claude-3-5-haiku","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":200,"cache_read_input_tokens":8000}}}
EOF

cat > "${GALLERY_HOME}/.cache/thermal/pricing.json" <<EOF
{
  "version": 1,
  "fetchedAt": "${SESSION_DATE}",
  "source": "https://models.dev/api.json",
  "models": {
    "claude-3-5-sonnet": { "input": 3.0, "output": 15.0, "cacheRead": 0.3, "cacheWrite": 3.75 },
    "claude-3.5-sonnet": { "input": 3.0, "output": 15.0, "cacheRead": 0.3, "cacheWrite": 3.75 },
    "claude-3-5-haiku": { "input": 1.0, "output": 5.0, "cacheRead": 0.1, "cacheWrite": 1.25 },
    "claude-3.5-haiku": { "input": 1.0, "output": 5.0, "cacheRead": 0.1, "cacheWrite": 1.25 }
  }
}
EOF

cat > "${GALLERY_HOME}/.gemini/antigravity/brain/gallery/.system_generated/logs/transcript.jsonl" <<EOF
{"created_at":"${SESSION_DATE}","source":"USER_EXPLICIT","type":"USER_INPUT","content":"gallery request"}
{"created_at":"${SESSION_DATE}","source":"SYSTEM","type":"STEP","content":"Model Selection changed from None to Gemini 3.1 Pro (High)."}
EOF
for i in $(seq 1 10); do
  echo "{\"created_at\":\"${SESSION_DATE}\",\"source\":\"MODEL\",\"type\":\"PLANNER_RESPONSE\",\"content\":\"step ${i}\"}" >> "${GALLERY_HOME}/.gemini/antigravity/brain/gallery/.system_generated/logs/transcript.jsonl"
done

# Audit fixture: duplicate basenames in repo and home scopes.
for i in $(seq 1 160); do echo "rule ${i}"; done > "${WORK}/AGENTS.md"
printf '# Repo guidance\nKeep it short.\n' > "${WORK}/CLAUDE.md"
printf '/thermal\n/dist/\n*.exe\n' > "${WORK}/.gitignore"
printf '# Home guidance\n' > "${GALLERY_HOME}/.claude/CLAUDE.md"

printf '%s\n' "${GALLERY_HOME}"
