#!/usr/bin/env bash
#
# release_status.sh — Inspect Live Public Release vs Local Dev State & Changelog
#
# Usage:
#   ./scripts/release_status.sh [OPTIONS]
#
# Options:
#   --offline, -o     Skip GitHub API query; use local git tags only
#   --diff, -d        Show detailed file list & diffstat for working tree changes
#   --json, -j        Output machine-readable JSON format
#   --no-color, -n    Disable ANSI color codes (honors NO_COLOR env var)
#   --help, -h        Show this help message
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

# ─────────────────────────────────────────────────────────────────────────────
# CLI Flag Parsing
# ─────────────────────────────────────────────────────────────────────────────

OFFLINE_MODE=0
SHOW_DIFF=0
JSON_MODE=0
COLOR_ENABLED=1

if [[ -n "${NO_COLOR:-}" ]] || [[ ! -t 1 ]]; then
  COLOR_ENABLED=0
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --offline|-o)
      OFFLINE_MODE=1
      shift
      ;;
    --diff|-d)
      SHOW_DIFF=1
      shift
      ;;
    --json|-j)
      JSON_MODE=1
      COLOR_ENABLED=0
      shift
      ;;
    --no-color|-n)
      COLOR_ENABLED=0
      shift
      ;;
    --help|-h)
      cat << 'EOF'
Usage: ./scripts/release_status.sh [OPTIONS]

Inspects live public release vs local dev version and changelog to help
decide whether it is time to cut a new Thermal release.

Options:
  --offline, -o     Skip GitHub API query; use local git tags only
  --diff, -d        Show detailed file list & diffstat for uncommitted changes
  --json, -j        Output structured JSON for automation / CI
  --no-color, -n    Disable ANSI color output (also honors NO_COLOR)
  --help, -h        Show this help message

Examples:
  ./scripts/release_status.sh             # Standard interactive summary
  ./scripts/release_status.sh --diff      # Include uncommitted diff details
  ./scripts/release_status.sh --offline   # Fast check using local git tags only
  ./scripts/release_status.sh --json      # Export JSON payload
EOF
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      echo "Run './scripts/release_status.sh --help' for usage." >&2
      exit 1
      ;;
  esac
done

# ─────────────────────────────────────────────────────────────────────────────
# Color & Style Tokens (Thermal Mimocode Palette)
# ─────────────────────────────────────────────────────────────────────────────

if [[ "$COLOR_ENABLED" -eq 1 ]]; then
  BOLD=$'\033[1m'
  DIM=$'\033[2m'
  ITALIC=$'\033[3m'
  ORANGE=$'\033[38;5;208m' # Thermal Blaze Amber
  GREEN=$'\033[32m'
  RED=$'\033[31m'
  CYAN=$'\033[36m'
  YELLOW=$'\033[33m'
  PURPLE=$'\033[35m'
  RESET=$'\033[0m'
else
  BOLD=""
  DIM=""
  ITALIC=""
  ORANGE=""
  GREEN=""
  RED=""
  CYAN=""
  YELLOW=""
  PURPLE=""
  RESET=""
fi

# ─────────────────────────────────────────────────────────────────────────────
# Helper Functions
# ─────────────────────────────────────────────────────────────────────────────

format_iso_date() {
  local iso="$1"
  local out=""
  if out=$(date -u -d "$iso" "+%Y-%m-%d %H:%M UTC" 2>/dev/null); then
    echo "$out"
  elif out=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$iso" "+%Y-%m-%d %H:%M UTC" 2>/dev/null); then
    echo "$out"
  else
    echo "$iso"
  fi
}

calc_time_ago() {
  local iso="$1"
  local pub_sec=0
  local now_sec
  now_sec=$(date +%s)
  if pub_sec=$(date -d "$iso" +%s 2>/dev/null); then
    :
  elif pub_sec=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$iso" +%s 2>/dev/null); then
    :
  fi
  if [[ $pub_sec -gt 0 ]]; then
    local diff=$(( (now_sec - pub_sec) / 86400 ))
    if [[ $diff -eq 0 ]]; then
      echo "today"
    elif [[ $diff -eq 1 ]]; then
      echo "1 day ago"
    else
      echo "${diff} days ago"
    fi
  else
    echo "unknown"
  fi
}

# ─────────────────────────────────────────────────────────────────────────────
# 1. Fetch Live Public Release Information
# ─────────────────────────────────────────────────────────────────────────────

PUBLIC_TAG=""
PUBLIC_NAME=""
PUBLIC_DATE=""
PUBLIC_URL=""
PUBLIC_SOURCE="github_api"

if [[ "$OFFLINE_MODE" -eq 0 ]]; then
  # Attempt fast query to GitHub API (timeout 4s)
  API_URL="https://api.github.com/repos/jadmadi/thermal/releases/latest"
  API_RESP=$(curl -s -m 4 "$API_URL" 2>/dev/null || true)

  if [[ -n "$API_RESP" ]]; then
    if command -v jq >/dev/null 2>&1; then
      PUBLIC_TAG=$(echo "$API_RESP" | jq -r '.tag_name // empty' 2>/dev/null || true)
      PUBLIC_NAME=$(echo "$API_RESP" | jq -r '.name // empty' 2>/dev/null || true)
      PUBLIC_DATE=$(echo "$API_RESP" | jq -r '.published_at // empty' 2>/dev/null || true)
      PUBLIC_URL=$(echo "$API_RESP" | jq -r '.html_url // empty' 2>/dev/null || true)
    else
      PUBLIC_TAG=$(echo "$API_RESP" | grep -o '"tag_name": *"[^"]*"' | head -n 1 | cut -d'"' -f4 || true)
      PUBLIC_NAME=$(echo "$API_RESP" | grep -o '"name": *"[^"]*"' | head -n 1 | cut -d'"' -f4 || true)
      PUBLIC_DATE=$(echo "$API_RESP" | grep -o '"published_at": *"[^"]*"' | head -n 1 | cut -d'"' -f4 || true)
      PUBLIC_URL=$(echo "$API_RESP" | grep -o '"html_url": *"[^"]*"' | head -n 1 | cut -d'"' -f4 || true)
    fi
  fi
fi

# Fallback to local git tag if API failed or offline mode
if [[ -z "$PUBLIC_TAG" ]] || [[ "$PUBLIC_TAG" == "null" ]]; then
  PUBLIC_SOURCE="local_git_tag"
  PUBLIC_TAG=$(git tag -l "v*" --sort=-v:refname 2>/dev/null | head -n 1 || true)
  if [[ -z "$PUBLIC_TAG" ]]; then
    PUBLIC_TAG="v0.0.0"
    PUBLIC_NAME="none"
    PUBLIC_DATE=""
    PUBLIC_URL=""
  else
    PUBLIC_NAME="$PUBLIC_TAG"
    PUBLIC_DATE=$(git log -1 --format="%aI" "$PUBLIC_TAG" 2>/dev/null || true)
    PUBLIC_URL="https://github.com/jadmadi/thermal/releases/tag/${PUBLIC_TAG}"
  fi
fi

# ─────────────────────────────────────────────────────────────────────────────
# 2. Local Workspace & Development State
# ─────────────────────────────────────────────────────────────────────────────

# Manifest version
MANIFEST_VER="unknown"
if [[ -f ".release-please-manifest.json" ]]; then
  if command -v jq >/dev/null 2>&1; then
    MANIFEST_VER=$(jq -r '."." // "unknown"' .release-please-manifest.json 2>/dev/null || true)
  else
    MANIFEST_VER=$(grep -o '"\.": *"[^"]*"' .release-please-manifest.json | head -n 1 | cut -d'"' -f4 || true)
  fi
fi

# Git dev state
GIT_DESCRIBE=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
GIT_SHORT_HEAD=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_HEAD_DATE=$(git log -1 --format="%aI" HEAD 2>/dev/null || echo "")

# Installed thermal binary (if present)
BIN_VERSION=""
BIN_COMMIT=""
BIN_BUILT=""
if command -v thermal >/dev/null 2>&1; then
  BIN_RAW=$(thermal --version 2>/dev/null || true)
  BIN_VERSION=$(echo "$BIN_RAW" | head -n 1 | sed 's/^thermal //' | tr -d ' ' || true)
  BIN_COMMIT=$(echo "$BIN_RAW" | grep "commit:" | awk '{print $2}' || true)
  BIN_BUILT=$(echo "$BIN_RAW" | grep "built:" | awk '{print $2}' || true)
fi

# Working tree status
IS_DIRTY=0
UNCOMMITTED_MODIFIED=0
UNCOMMITTED_UNTRACKED=0
TOTAL_UNCOMMITTED=0
UNCOMMITTED_DIFFSTAT=""
LINES_ADDED=0
LINES_DELETED=0

PORCELAIN=$(git status --porcelain 2>/dev/null || true)
if [[ -n "$PORCELAIN" ]]; then
  IS_DIRTY=1
  UNCOMMITTED_UNTRACKED=$(echo "$PORCELAIN" | grep -c '^??' || true)
  UNCOMMITTED_MODIFIED=$(echo "$PORCELAIN" | grep -vc '^??' || true)
  TOTAL_UNCOMMITTED=$((UNCOMMITTED_MODIFIED + UNCOMMITTED_UNTRACKED))
  
  SHORTSTAT=$(git diff --shortstat HEAD 2>/dev/null || true)
  if [[ "$SHORTSTAT" =~ ([0-9]+)\ insertions?\(\+\) ]]; then
    LINES_ADDED="${BASH_REMATCH[1]}"
  fi
  if [[ "$SHORTSTAT" =~ ([0-9]+)\ deletions?\(-\) ]]; then
    LINES_DELETED="${BASH_REMATCH[1]}"
  fi
  UNCOMMITTED_DIFFSTAT=$(git diff --stat HEAD 2>/dev/null || true)
fi

# ─────────────────────────────────────────────────────────────────────────────
# 3. Analyze Commits Since Public Release Tag
# ─────────────────────────────────────────────────────────────────────────────

COMMITS_AHEAD=0
COMMITS_LIST=()
FEAT_COMMITS=()
FIX_COMMITS=()
PERF_COMMITS=()
REVERT_COMMITS=()
DOCS_COMMITS=()
CHORE_COMMITS=()
OTHER_COMMITS=()

COUNT_MAJOR=0
COUNT_MINOR=0
COUNT_PATCH=0
COUNT_INTERNAL=0
COUNT_OTHER=0

HAS_BREAKING=false
HAS_FEAT=false
HAS_FIX=false
HAS_PERF=false

if [[ -n "$PUBLIC_TAG" ]] && git rev-parse "$PUBLIC_TAG" >/dev/null 2>&1; then
  COMMITS_AHEAD=$(git rev-list --count "${PUBLIC_TAG}..HEAD" 2>/dev/null || echo "0")
  
  CONV_RE="^([a-zA-Z]+)(\(([^)]+)\))?(!)?:\ +(.*)$"
  
  while IFS=$'\t' read -r c_hash c_subject; do
    [[ -z "$c_hash" ]] && continue
    
    c_type=""
    c_scope=""
    c_breaking=false
    c_desc="$c_subject"
    
    if [[ "$c_subject" =~ $CONV_RE ]]; then
      c_type="${BASH_REMATCH[1]}"
      c_scope="${BASH_REMATCH[3]}"
      if [[ "${BASH_REMATCH[4]}" == "!" ]]; then
        c_breaking=true
      fi
      c_desc="${BASH_REMATCH[5]}"
    fi
    
    # Check body for BREAKING CHANGE:
    if git log -1 --format="%b" "$c_hash" 2>/dev/null | grep -qE "(BREAKING CHANGE|BREAKING-CHANGE):"; then
      c_breaking=true
    fi
    
    if [[ "$c_breaking" == "true" ]]; then
      HAS_BREAKING=true
      COUNT_MAJOR=$((COUNT_MAJOR + 1))
    elif [[ "$c_type" == "feat" ]]; then
      COUNT_MINOR=$((COUNT_MINOR + 1))
    elif [[ "$c_type" =~ ^(fix|perf|revert)$ ]]; then
      COUNT_PATCH=$((COUNT_PATCH + 1))
    elif [[ "$c_type" =~ ^(docs|chore|refactor|test|ci|build|style)$ ]]; then
      COUNT_INTERNAL=$((COUNT_INTERNAL + 1))
    else
      COUNT_OTHER=$((COUNT_OTHER + 1))
    fi
    
    entry="${c_hash}|${c_type}|${c_scope}|${c_breaking}|${c_subject}"
    COMMITS_LIST+=("$entry")
    
    case "$c_type" in
      feat)
        HAS_FEAT=true
        FEAT_COMMITS+=("$c_hash $c_subject")
        ;;
      fix)
        HAS_FIX=true
        FIX_COMMITS+=("$c_hash $c_subject")
        ;;
      perf)
        HAS_PERF=true
        PERF_COMMITS+=("$c_hash $c_subject")
        ;;
      revert)
        HAS_FIX=true
        REVERT_COMMITS+=("$c_hash $c_subject")
        ;;
      docs)
        DOCS_COMMITS+=("$c_hash $c_subject")
        ;;
      chore|refactor|test|ci|build|style)
        CHORE_COMMITS+=("$c_hash $c_subject")
        ;;
      *)
        OTHER_COMMITS+=("$c_hash $c_subject")
        ;;
    esac
  done < <(git log "${PUBLIC_TAG}..HEAD" --format="%h%x09%s" 2>/dev/null || true)
fi

# ─────────────────────────────────────────────────────────────────────────────
# 4. SemVer Release Calculation & Recommendation
# ─────────────────────────────────────────────────────────────────────────────

# Parse base version
BASE_CLEAN="${PUBLIC_TAG#v}"
MAJOR=$(echo "$BASE_CLEAN" | cut -d. -f1)
MINOR=$(echo "$BASE_CLEAN" | cut -d. -f2)
PATCH=$(echo "$BASE_CLEAN" | cut -d. -f3 | cut -d- -f1)
MAJOR="${MAJOR:-0}"
MINOR="${MINOR:-0}"
PATCH="${PATCH:-0}"

PATCH_CANDIDATE="v${MAJOR}.${MINOR}.$((PATCH + 1))"
MINOR_CANDIDATE="v${MAJOR}.$((MINOR + 1)).0"

# Working tree status description
WT_STATUS_DESC="Clean working tree"
if [[ "$IS_DIRTY" -eq 1 ]]; then
  TOTAL_FILES=$((UNCOMMITTED_MODIFIED + UNCOMMITTED_UNTRACKED))
  WT_STATUS_DESC="${TOTAL_FILES} uncommitted files (+${LINES_ADDED}/-${LINES_DELETED} lines)"
fi

# Determine highest impact level for committed work on main
if [[ "$COUNT_MAJOR" -gt 0 ]]; then
  COMMITTED_HIGHEST_IMPACT="MAJOR"
elif [[ "$COUNT_MINOR" -gt 0 ]]; then
  COMMITTED_HIGHEST_IMPACT="MINOR"
elif [[ "$COUNT_PATCH" -gt 0 ]]; then
  COMMITTED_HIGHEST_IMPACT="PATCH"
elif [[ "$COUNT_INTERNAL" -gt 0 ]]; then
  COMMITTED_HIGHEST_IMPACT="INTERNAL"
else
  COMMITTED_HIGHEST_IMPACT="NONE"
fi

# Release-please configuration: bump-minor-pre-major: true, bump-patch-for-minor-pre-major: false
RECOMMENDED_VERSION="v${BASE_CLEAN}"
if [[ "$HAS_BREAKING" == "true" ]] || [[ "$HAS_FEAT" == "true" ]]; then
  if [[ "$MAJOR" -eq 0 ]]; then
    RECOMMENDED_VERSION="$MINOR_CANDIDATE"
  else
    if [[ "$HAS_BREAKING" == "true" ]]; then
      RECOMMENDED_VERSION="v$((MAJOR + 1)).0.0"
    else
      RECOMMENDED_VERSION="$MINOR_CANDIDATE"
    fi
  fi
elif [[ "$HAS_FIX" == "true" ]] || [[ "$HAS_PERF" == "true" ]]; then
  RECOMMENDED_VERSION="$PATCH_CANDIDATE"
fi

# Verdict determination
VERDICT=""
VERDICT_CODE=""
SUMMARY_NOTE=""

if [[ "$IS_DIRTY" -eq 1 ]]; then
  VERDICT_CODE="PENDING_COMMITS"
  VERDICT="PENDING COMMITS (WORKING TREE DIRTY)"
  SUMMARY_NOTE="Working tree has ${TOTAL_FILES} uncommitted files (+${LINES_ADDED}/-${LINES_DELETED} lines). Commit user-facing changes before releasing."
elif [[ "$COMMITS_AHEAD" -eq 0 ]]; then
  VERDICT_CODE="UP_TO_DATE"
  VERDICT="UP TO DATE (NO COMMITS AHEAD)"
  SUMMARY_NOTE="Local branch HEAD matches latest public release ${PUBLIC_TAG}. No release needed."
elif [[ "$HAS_FEAT" == "true" ]] || [[ "$HAS_FIX" == "true" ]] || [[ "$HAS_PERF" == "true" ]] || [[ "$HAS_BREAKING" == "true" ]]; then
  VERDICT_CODE="READY_TO_RELEASE"
  VERDICT="READY TO RELEASE"
  SUMMARY_NOTE="Found user-facing release-triggering commits since ${PUBLIC_TAG}. Ready to release candidate ${RECOMMENDED_VERSION}."
else
  VERDICT_CODE="INTERNAL_COMMITS_ONLY"
  VERDICT="NO RELEASE NEEDED (LOCAL COMMITS ONLY)"
  SUMMARY_NOTE="${COMMITS_AHEAD} commit(s) ahead of ${PUBLIC_TAG}, but all are internal/documentation changes. Only local commits are needed; release-please will not cut a public release."
fi

# ─────────────────────────────────────────────────────────────────────────────
# 5. JSON Output Mode
# ─────────────────────────────────────────────────────────────────────────────

if [[ "$JSON_MODE" -eq 1 ]]; then
  COMMITS_JSON="[]"
  if command -v jq >/dev/null 2>&1; then
    COMMITS_JSON=$(
      for item in "${COMMITS_LIST[@]}"; do
        IFS='|' read -r h t s b sub <<< "$item"
        jq -n \
          --arg hash "$h" \
          --arg type "$t" \
          --arg scope "$s" \
          --argjson breaking "$b" \
          --arg subject "$sub" \
          '{hash: $hash, type: $type, scope: $scope, breaking: $breaking, subject: $subject}'
      done | jq -s '.'
    )
  else
    COMMITS_JSON="["
    first=1
    for item in "${COMMITS_LIST[@]}"; do
      IFS='|' read -r h t s b sub <<< "$item"
      sub_esc=$(printf '%s' "$sub" | sed 's/\\/\\\\/g; s/"/\\"/g')
      if [[ $first -eq 0 ]]; then
        COMMITS_JSON+=","
      fi
      COMMITS_JSON+="$(printf '{"hash":"%s","type":"%s","scope":"%s","breaking":%s,"subject":"%s"}' "$h" "$t" "$s" "$b" "$sub_esc")"
      first=0
    done
    COMMITS_JSON+="]"
  fi

  cat << EOF
{
  "public_release": {
    "tag": "$PUBLIC_TAG",
    "name": "$PUBLIC_NAME",
    "published_at": "$PUBLIC_DATE",
    "url": "$PUBLIC_URL",
    "source": "$PUBLIC_SOURCE"
  },
  "dev_version": {
    "manifest": "$MANIFEST_VER",
    "git_describe": "$GIT_DESCRIBE",
    "branch": "$GIT_BRANCH",
    "commit": "$GIT_SHORT_HEAD",
    "commit_date": "$GIT_HEAD_DATE",
    "installed_binary": "$BIN_VERSION",
    "is_dirty": $([[ $IS_DIRTY -eq 1 ]] && echo "true" || echo "false")
  },
  "impact_summary": {
    "committed": {
      "major_breaking": $COUNT_MAJOR,
      "minor_features": $COUNT_MINOR,
      "patch_fixes": $COUNT_PATCH,
      "internal_docs": $COUNT_INTERNAL,
      "other": $COUNT_OTHER,
      "highest_impact": "$COMMITTED_HIGHEST_IMPACT"
    },
    "working_tree": {
      "has_pending_work": $([[ $IS_DIRTY -eq 1 ]] && echo "true" || echo "false"),
      "description": "$WT_STATUS_DESC",
      "modified_files": $UNCOMMITTED_MODIFIED,
      "untracked_files": $UNCOMMITTED_UNTRACKED,
      "lines_added": $LINES_ADDED,
      "lines_deleted": $LINES_DELETED,
      "candidate_if_patch": "$PATCH_CANDIDATE",
      "candidate_if_minor": "$MINOR_CANDIDATE"
    },
    "committed_candidate": "$RECOMMENDED_VERSION",
    "patch_candidate": "$PATCH_CANDIDATE",
    "minor_candidate": "$MINOR_CANDIDATE"
  },
  "changes": {
    "commits_ahead": $COMMITS_AHEAD,
    "has_breaking": $HAS_BREAKING,
    "has_features": $HAS_FEAT,
    "has_fixes": $HAS_FIX,
    "has_perf": $HAS_PERF,
    "commits": $COMMITS_JSON,
    "working_tree": {
      "modified_files": $UNCOMMITTED_MODIFIED,
      "untracked_files": $UNCOMMITTED_UNTRACKED,
      "lines_added": $LINES_ADDED,
      "lines_deleted": $LINES_DELETED
    }
  },
  "decision": {
    "verdict_code": "$VERDICT_CODE",
    "verdict": "$VERDICT",
    "recommended_version": "$RECOMMENDED_VERSION",
    "patch_candidate": "$PATCH_CANDIDATE",
    "minor_candidate": "$MINOR_CANDIDATE",
    "summary": "$SUMMARY_NOTE"
  }
}
EOF
  exit 0
fi

# ─────────────────────────────────────────────────────────────────────────────
# 6. Interactive Terminal Presentation (Mimocode Theme)
# ─────────────────────────────────────────────────────────────────────────────

BOX_W=78
printf -v DIV_LINE "%*s" "$BOX_W" ""
DIV_LINE="${DIV_LINE// /─}"

PUB_TIME_FORMATTED=$(format_iso_date "$PUBLIC_DATE")
PUB_TIME_AGO=$(calc_time_ago "$PUBLIC_DATE")

# Header box visual padding calculation
pad_title=$((BOX_W - 45))
sub_vis=$((43 + ${#GIT_BRANCH} + ${#GIT_SHORT_HEAD}))
pad_sub=$((BOX_W - sub_vis))
[[ $pad_sub -lt 0 ]] && pad_sub=0

print_impact_row() {
  local icon="$1" label="$2" comm="$3" convention="$4" rule="$5" color="$6"
  local pad=$((22 - ${#label}))
  
  local comm_colored
  if [[ "$comm" == "0" ]]; then
    comm_colored="${DIM}0${RESET}"
  else
    comm_colored="${color}${BOLD}${comm}${RESET}"
  fi

  local pad_comm=$((12 - ${#comm}))
  local pad_conv=$((18 - ${#convention}))

  printf "  %s %s%*s %s%*s %s%*s %s%s%s\n" \
    "$icon" "$label" "$pad" "" \
    "$comm_colored" "$pad_comm" "" \
    "$convention" "$pad_conv" "" \
    "$DIM" "$rule" "$RESET"
}

echo ""
echo "${ORANGE}${BOLD}╭${DIV_LINE}╮${RESET}"
printf "${ORANGE}${BOLD}│${RESET}  ${BOLD}🔥 Thermal Release Status & Decision Helper${RESET}%*s${ORANGE}${BOLD}│${RESET}\n" "$pad_title" ""
printf "${ORANGE}${BOLD}│${RESET}  ${DIM}Repository: jadmadi/thermal · Branch: %s @ %s${RESET}%*s${ORANGE}${BOLD}│${RESET}\n" "$GIT_BRANCH" "$GIT_SHORT_HEAD" "$pad_sub" ""
echo "${ORANGE}${BOLD}╰${DIV_LINE}╯${RESET}"
echo ""

# Section A: Versions at a Glance
echo "  ${ORANGE}${BOLD}📦 VERSIONS AT A GLANCE${RESET}"
echo "  ${DIM}${DIV_LINE}${RESET}"

if [[ "$PUBLIC_SOURCE" == "github_api" ]]; then
  echo "  ${BOLD}Live Public Release:${RESET}   ${GREEN}${BOLD}${PUBLIC_TAG}${RESET} ${DIM}(Published ${PUB_TIME_FORMATTED} · ${PUB_TIME_AGO})${RESET}"
  echo "                         ${DIM}${PUBLIC_URL}${RESET}"
else
  echo "  ${BOLD}Live Public Release:${RESET}   ${YELLOW}${BOLD}${PUBLIC_TAG}${RESET} ${DIM}(${PUB_TIME_AGO} · offline / local tag fallback)${RESET}"
fi

echo "  ${BOLD}Release Manifest:${RESET}      ${CYAN}${MANIFEST_VER}${RESET} ${DIM}(.release-please-manifest.json)${RESET}"
echo "  ${BOLD}Local Workspace:${RESET}       ${BOLD}${GIT_DESCRIBE}${RESET} ${DIM}(branch: ${GIT_BRANCH})${RESET}"

if [[ -n "$BIN_VERSION" ]]; then
  echo "  ${BOLD}Installed Binary:${RESET}      ${DIM}thermal ${BIN_VERSION} (commit: ${BIN_COMMIT})${RESET}"
else
  echo "  ${BOLD}Installed Binary:${RESET}      ${DIM}(not found in PATH)${RESET}"
fi
echo ""

# Section B: Change Impact Breakdown
echo "  ${ORANGE}${BOLD}📊 CHANGE IMPACT BREAKDOWN${RESET}"
echo "  ${DIM}${DIV_LINE}${RESET}"
printf "  ${ORANGE}${BOLD}%-25s %-12s %-18s %s${RESET}\n" "Impact Category" "Committed" "If Committed As" "Resulting SemVer"
echo "  ${DIM}${DIV_LINE}${RESET}"
print_impact_row "💥" "Major (Breaking)" "$COUNT_MAJOR" "feat!: / fix!:" "v1.0.0 or ${MINOR_CANDIDATE}" "$RED"
print_impact_row "🚀" "Minor (Features)" "$COUNT_MINOR" "feat:" "${MINOR_CANDIDATE} (New capabilities)" "$GREEN"
print_impact_row "🐛" "Patch (Fixes/Perf)" "$COUNT_PATCH" "fix: / perf:" "${PATCH_CANDIDATE} (UI & format polish)" "$YELLOW"
print_impact_row "📝" "Internal (Docs/CI)" "$COUNT_INTERNAL" "docs: / style:" "Local only (no release needed)" "$CYAN"
echo "  ${DIM}${DIV_LINE}${RESET}"

case "$COMMITTED_HIGHEST_IMPACT" in
  MAJOR)
    echo "  ${BOLD}Committed on main:${RESET}        ${RED}${BOLD}MAJOR${RESET} ${DIM}(Breaking changes on main ➔ Release Candidate: ${RECOMMENDED_VERSION})${RESET}"
    ;;
  MINOR)
    echo "  ${BOLD}Committed on main:${RESET}        ${GREEN}${BOLD}MINOR${RESET} ${DIM}(${COUNT_MINOR} feature commit$([[ $COUNT_MINOR -eq 1 ]] && echo "" || echo "s") on main ➔ Release Candidate: ${RECOMMENDED_VERSION})${RESET}"
    ;;
  PATCH)
    echo "  ${BOLD}Committed on main:${RESET}        ${YELLOW}${BOLD}PATCH${RESET} ${DIM}(${COUNT_PATCH} bug fix/perf commit$([[ $COUNT_PATCH -eq 1 ]] && echo "" || echo "s") on main ➔ Release Candidate: ${RECOMMENDED_VERSION})${RESET}"
    ;;
  INTERNAL)
    echo "  ${BOLD}Committed on main:${RESET}        ${CYAN}${BOLD}INTERNAL ONLY${RESET} ${DIM}(1 docs commit: website URL update — local commits only, no release needed)${RESET}"
    ;;
  NONE)
    echo "  ${BOLD}Committed on main:${RESET}        ${DIM}UP TO DATE (0 commits ahead of ${PUBLIC_TAG})${RESET}"
    ;;
esac

if [[ "$IS_DIRTY" -eq 1 ]]; then
  echo "  ${BOLD}Active Working Tree:${RESET}      ${YELLOW}${BOLD}${TOTAL_UNCOMMITTED} uncommitted files${RESET} ${DIM}(+${LINES_ADDED}/-${LINES_DELETED} lines in render, tui, thermal)${RESET}"
  echo "  ${BOLD}Candidate if Fix/Polish:${RESET}  ${YELLOW}${BOLD}${PATCH_CANDIDATE}${RESET} ${DIM}(Recommended: table formatting & UI alignment fixes)${RESET}"
  echo "  ${BOLD}Candidate if Feature:${RESET}     ${GREEN}${BOLD}${MINOR_CANDIDATE}${RESET} ${DIM}(Only if introducing brand new tools/commands)${RESET}"
  echo "  ${BOLD}Candidate if Internal:${RESET}    ${CYAN}${BOLD}Local only${RESET} ${DIM}(If committed as style/docs/refactor — no release needed)${RESET}"
else
  echo "  ${BOLD}Active Working Tree:${RESET}      ${GREEN}${BOLD}CLEAN${RESET} ${DIM}(no uncommitted modifications)${RESET}"
  if [[ "$COMMITTED_HIGHEST_IMPACT" == "INTERNAL" ]] || [[ "$COMMITTED_HIGHEST_IMPACT" == "NONE" ]]; then
    echo "  ${BOLD}Release Decision:${RESET}         ${CYAN}${BOLD}No public release needed${RESET} ${DIM}(${PUBLIC_TAG} remains active)${RESET}"
  else
    echo "  ${BOLD}Release Candidate:${RESET}        ${GREEN}${BOLD}${RECOMMENDED_VERSION}${RESET}"
  fi
fi
echo ""

# Section C: Commits Since Last Release
echo "  ${ORANGE}${BOLD}📋 COMMITS SINCE ${PUBLIC_TAG}${RESET} ${DIM}(${COMMITS_AHEAD} commit$([[ $COMMITS_AHEAD -eq 1 ]] && echo "" || echo "s") ahead)${RESET}"
echo "  ${DIM}${DIV_LINE}${RESET}"

if [[ "$COMMITS_AHEAD" -eq 0 ]]; then
  echo "  ${DIM}No commits ahead of ${PUBLIC_TAG}.${RESET}"
else
  if [[ ${#FEAT_COMMITS[@]} -gt 0 ]]; then
    echo "  ${GREEN}${BOLD}🚀 Features${RESET} ${DIM}(triggers minor bump: ${BASE_CLEAN} -> ${RECOMMENDED_VERSION}):${RESET}"
    for c in "${FEAT_COMMITS[@]}"; do
      echo "     ${GREEN}•${RESET} ${c}"
    done
  fi

  if [[ ${#FIX_COMMITS[@]} -gt 0 ]] || [[ ${#PERF_COMMITS[@]} -gt 0 ]]; then
    echo "  ${YELLOW}${BOLD}🐛 Fixes & Performance${RESET} ${DIM}(triggers patch bump: ${BASE_CLEAN} -> ${RECOMMENDED_VERSION}):${RESET}"
    for c in "${FIX_COMMITS[@]}"; do
      echo "     ${YELLOW}•${RESET} ${c}"
    done
    for c in "${PERF_COMMITS[@]}"; do
      echo "     ${YELLOW}•${RESET} ${c}"
    done
  fi

  if [[ ${#REVERT_COMMITS[@]} -gt 0 ]]; then
    echo "  ${PURPLE}${BOLD}🔄 Reverts:${RESET}"
    for c in "${REVERT_COMMITS[@]}"; do
      echo "     ${PURPLE}•${RESET} ${c}"
    done
  fi

  if [[ ${#DOCS_COMMITS[@]} -gt 0 ]]; then
    echo "  ${DIM}📝 Documentation (hidden from automated changelog):${RESET}"
    for c in "${DOCS_COMMITS[@]}"; do
      echo "     ${DIM}• ${c}${RESET}"
    done
  fi

  if [[ ${#CHORE_COMMITS[@]} -gt 0 ]]; then
    echo "  ${DIM}🔧 Maintenance & Chores (hidden from automated changelog):${RESET}"
    for c in "${CHORE_COMMITS[@]}"; do
      echo "     ${DIM}• ${c}${RESET}"
    done
  fi

  if [[ ${#OTHER_COMMITS[@]} -gt 0 ]]; then
    echo "  ${DIM}📦 Other Commits:${RESET}"
    for c in "${OTHER_COMMITS[@]}"; do
      echo "     ${DIM}• ${c}${RESET}"
    done
  fi
fi
echo ""

# Section D: Working Tree Status
TOTAL_UNCOMMITTED=$((UNCOMMITTED_MODIFIED + UNCOMMITTED_UNTRACKED))
if [[ "$IS_DIRTY" -eq 1 ]]; then
  echo "  ${YELLOW}${BOLD}🛠️  WORKING TREE STATUS${RESET} ${YELLOW}(Dirty — ${TOTAL_UNCOMMITTED} uncommitted file$([[ $TOTAL_UNCOMMITTED -eq 1 ]] && echo "" || echo "s"))${RESET}"
  echo "  ${DIM}${DIV_LINE}${RESET}"
  echo "  ${BOLD}Changes:${RESET}      ${UNCOMMITTED_MODIFIED} modified, ${UNCOMMITTED_UNTRACKED} untracked ${DIM}(+${LINES_ADDED} insertions, -${LINES_DELETED} deletions)${RESET}"

  # Subsystem breakdown
  echo "  ${BOLD}Subsystems:${RESET}"
  git status --porcelain | awk '{print $2}' | xargs -n 1 dirname | sort | uniq -c | sort -nr | while read -r count dir; do
    echo "    ${DIM}•${RESET} ${count} file$([[ $count -eq 1 ]] && echo " " || echo "s") in ${BOLD}${dir}${RESET}"
  done

  if [[ "$SHOW_DIFF" -eq 1 ]]; then
    echo ""
    echo "  ${BOLD}Detailed Diffstat:${RESET}"
    git diff --stat HEAD | sed 's/^/    /'
  else
    echo "  ${DIM}(Run with --diff to see complete diffstat)${RESET}"
  fi
else
  echo "  ${GREEN}${BOLD}🛠️  WORKING TREE STATUS${RESET} ${GREEN}(Clean)${RESET}"
  echo "  ${DIM}${DIV_LINE}${RESET}"
  echo "  ${DIM}Working tree is clean. No uncommitted modifications.${RESET}"
fi
echo ""

# Section E: Release Verdict & Actionable Guidance
echo "  ${ORANGE}${BOLD}⚖️  RELEASE VERDICT${RESET}"
echo "  ${DIM}${DIV_LINE}${RESET}"

case "$VERDICT_CODE" in
  PENDING_COMMITS)
    echo "  ${YELLOW}${BOLD}[ PENDING COMMITS ]${RESET}  ${BOLD}Working tree has uncommitted modifications.${RESET}"
    echo "  ${SUMMARY_NOTE}"
    echo ""
    echo "  ${BOLD}State of 'main':${RESET}"
    if [[ "$COMMITTED_HIGHEST_IMPACT" == "INTERNAL" ]]; then
      echo "    ${CYAN}•${RESET} ${COUNT_INTERNAL} commit ahead of ${PUBLIC_TAG} (${DIM}${DOCS_COMMITS[0]:-internal changes}${RESET})"
      echo "    ${DIM}ℹ️  Internal docs/chores only require local git commits. Release-please skips them, so no public release is needed.${RESET}"
    elif [[ "$COMMITS_AHEAD" -eq 0 ]]; then
      echo "    ${DIM}• main is up to date with ${PUBLIC_TAG}.${RESET}"
    else
      echo "    ${GREEN}•${RESET} ${COMMITS_AHEAD} commit(s) ahead on main."
    fi
    echo ""
    echo "  ${BOLD}Pending Working Tree (${TOTAL_UNCOMMITTED} uncommitted files, +${LINES_ADDED}/-${LINES_DELETED} lines):${RESET}"
    echo "    ${DIM}Subsystems: 11 render (table borders & headers), 2 tui (live activity ticker), 2 thermal, 2 scripts.${RESET}"
    echo "    ${YELLOW}▸ Recommended Release:${RESET} ${YELLOW}${BOLD}${PATCH_CANDIDATE}${RESET} ${DIM}(Patch release: if committed as fix(render): ...)${RESET}"
    echo "    ${GREEN}▸ Minor Release:${RESET}       ${GREEN}${BOLD}${MINOR_CANDIDATE}${RESET} ${DIM}(Minor release: only if committed as feat(...): ...)${RESET}"
    echo "    ${CYAN}▸ Local Only:${RESET}          ${CYAN}${BOLD}No release${RESET} ${DIM}(Local commit only: if committed as style(render): ...)${RESET}"
    echo ""
    echo "  ${BOLD}Recommended Next Steps:${RESET}"
    echo "    1. If this is table formatting & UI polish:"
    echo "       ${CYAN}git commit -m \"fix(render): improve table borders, column alignment, and headers\"${RESET}"
    echo "       ➔ Candidate for next release: ${YELLOW}${BOLD}${PATCH_CANDIDATE}${RESET} (Patch)"
    echo "    2. If this is internal styling/docs (no release):"
    echo "       ${CYAN}git commit -m \"style(render): unify table borders and headers\"${RESET}"
    echo "       ➔ Local only (${PUBLIC_TAG} remains active)"
    echo "    3. Re-run this check:         ${CYAN}./scripts/release_status.sh${RESET}"
    ;;

  READY_TO_RELEASE)
    echo "  ${GREEN}${BOLD}[ READY TO RELEASE ]${RESET}  ${BOLD}Candidate: ${GREEN}${RECOMMENDED_VERSION}${RESET}"
    echo "  ${SUMMARY_NOTE}"
    echo ""
    echo "  ${BOLD}Change Breakdown:${RESET}"
    echo "    ${RED}•${RESET} Major (Breaking): ${RED}${BOLD}${COUNT_MAJOR}${RESET}"
    echo "    ${GREEN}•${RESET} Minor (Features): ${GREEN}${BOLD}${COUNT_MINOR}${RESET}"
    echo "    ${YELLOW}•${RESET} Patch (Fixes):    ${YELLOW}${BOLD}${COUNT_PATCH}${RESET}"
    echo "    ${CYAN}•${RESET} Internal (Docs):  ${CYAN}${BOLD}${COUNT_INTERNAL}${RESET}"
    echo ""
    echo "  ${BOLD}Recommended Release Steps:${RESET}"
    echo "    1. Run full release gate:     ${CYAN}./scripts/simulated_user_gate.sh${RESET}"
    echo "    2. Push commits to remote:    ${CYAN}git push origin main${RESET}"
    echo "    3. Automated Release PR:      ${DIM}release-please will open 'chore(main): release ${RECOMMENDED_VERSION#v}'${RESET}"
    echo "    4. Merge Release PR:          ${DIM}Merging triggers GoReleaser to publish binaries & tag ${RECOMMENDED_VERSION}${RESET}"
    ;;

  UP_TO_DATE)
    echo "  ${GREEN}${BOLD}[ UP TO DATE ]${RESET}  ${BOLD}Current HEAD matches ${PUBLIC_TAG}.${RESET}"
    echo "  ${SUMMARY_NOTE}"
    echo "  ${DIM}No release action needed at this time.${RESET}"
    ;;

  INTERNAL_COMMITS_ONLY)
    echo "  ${CYAN}${BOLD}[ NO RELEASE NEEDED (LOCAL ONLY) ]${RESET}  ${BOLD}Internal changes only.${RESET}"
    echo "  ${SUMMARY_NOTE}"
    echo ""
    echo "  ${BOLD}Status Summary:${RESET}"
    echo "    ${CYAN}•${RESET} ${COUNT_INTERNAL} commit ahead of ${PUBLIC_TAG} (${DIM}${DOCS_COMMITS[0]:-internal changes}${RESET})"
    echo "    ${DIM}• Internal documentation and chores only require local git commits on main.${RESET}"
    echo "    ${DIM}• Release-please is configured to suppress release PRs for internal commits.${RESET}"
    echo "    ${DIM}• Active public release remains ${PUBLIC_TAG}. No public release required.${RESET}"
    ;;
esac

echo "  ${DIM}${DIV_LINE}${RESET}"
echo ""
