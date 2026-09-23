#!/usr/bin/env bash
#
# purge_devin_commits.sh — Audit and purge devin-ai-integration[bot] from Git history
#
# Re-attributes commits authored or committed by devin-ai-integration[bot]
# to Jad Madi <jadmadi@gmail.com> so GitHub's contributor graph reflects Jad Madi.
#
set -euo pipefail

BOLD=$'\033[1m'
GREEN=$'\033[32m'
RED=$'\033[31m'
CYAN=$'\033[36m'
YELLOW=$'\033[33m'
RESET=$'\033[0m'

TARGET_NAME="Jad Madi"
TARGET_EMAIL="jadmadi@gmail.com"
BOT_NAME="devin-ai-integration[bot]"
BOT_EMAIL_SUBSTR="devin"

MODE="audit"
AUTO_CONFIRM=false

for arg in "$@"; do
    case "$arg" in
        --execute|--rewrite)
            MODE="rewrite"
            ;;
        -y|--yes)
            AUTO_CONFIRM=true
            ;;
        --audit|--check)
            MODE="audit"
            ;;
        -h|--help)
            echo "Usage: $0 [--audit | --rewrite [-y]]"
            echo "  --audit     Check git history for Devin bot commits (default)"
            echo "  --rewrite   Execute git history rewrite reattributing Devin bot commits to $TARGET_NAME <$TARGET_EMAIL>"
            echo "  -y, --yes   Skip confirmation prompt during rewrite"
            exit 0
            ;;
        *)
            echo "Unknown argument: $arg"
            echo "Run with --help for usage."
            exit 1
            ;;
    esac
done

echo -e "\n${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  THERMAL · Devin Bot Contributor Attribution Audit & Purge${RESET}"
echo -e "${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}\n"

# Phase 1: Audit commits matching devin bot
echo -e "${BOLD}1. Auditing Git history for ${BOT_NAME}...${RESET}"

DEVIN_COMMITS=$(git log --all --format='%H|%an|%ae|%cn|%ce' | grep -i "${BOT_EMAIL_SUBSTR}" || true)

if [ -z "$DEVIN_COMMITS" ]; then
    echo -e "   ${GREEN}✔ Clean:${RESET} Zero commits found authored or committed by ${BOT_NAME}."
    echo -e "   Current git log author breakdown:"
    git shortlog -se | sed 's/^/     /'
    echo ""
    if [ "$MODE" = "rewrite" ]; then
        echo -e "   No rewrite necessary. History is already clean."
    fi
    exit 0
fi

COUNT=$(echo "$DEVIN_COMMITS" | wc -l)
echo -e "   ${YELLOW}⚠ Found ${COUNT} commit(s) matching Devin bot:${RESET}"
while IFS='|' read -r sha an ae cn ce; do
    echo -e "     • SHA: ${sha:0:8} | Author: ${an} <${ae}> | Committer: ${cn} <${ce}>"
done <<< "$DEVIN_COMMITS"

if [ "$MODE" = "audit" ]; then
    echo -e "\n${YELLOW}Audit complete.${RESET} To rewrite these commits to ${BOLD}${TARGET_NAME} <${TARGET_EMAIL}>${RESET}, run:"
    echo -e "   $0 --rewrite\n"
    exit 0
fi

# Phase 2: Rewrite commits
echo -e "\n${BOLD}2. Rewriting Git history...${RESET}"
if [ "$AUTO_CONFIRM" = false ]; then
    read -r -p "Are you sure you want to rewrite history for these ${COUNT} commits? [y/N] " response
    case "$response" in
        [yY][eE][sS]|[yY]) ;;
        *)
            echo "Aborted by user."
            exit 1
            ;;
    esac
fi

if command -v git-filter-repo >/dev/null 2>&1; then
    echo -e "   Using ${CYAN}git-filter-repo${RESET}..."
    git filter-repo --force --mailmap .mailmap
else
    echo -e "   Using ${CYAN}git filter-branch${RESET}..."
    FILTER_BRANCH_SQUELCH_WARNING=1 git filter-branch -f --env-filter "
    if [ \"\$GIT_AUTHOR_NAME\" = \"${BOT_NAME}\" ] || [[ \"\$GIT_AUTHOR_EMAIL\" == *\"${BOT_EMAIL_SUBSTR}\"* ]]; then
        export GIT_AUTHOR_NAME=\"${TARGET_NAME}\"
        export GIT_AUTHOR_EMAIL=\"${TARGET_EMAIL}\"
    fi
    if [ \"\$GIT_COMMITTER_NAME\" = \"${BOT_NAME}\" ] || [[ \"\$GIT_COMMITTER_EMAIL\" == *\"${BOT_EMAIL_SUBSTR}\"* ]]; then
        export GIT_COMMITTER_NAME=\"${TARGET_NAME}\"
        export GIT_COMMITTER_EMAIL=\"${TARGET_EMAIL}\"
    fi
    " --tag-name-filter cat -- --all
fi

echo -e "   ${GREEN}✔ History rewrite complete.${RESET}"

# Phase 3: Verification & Push Instructions
echo -e "\n${BOLD}3. Verification & Push Instructions${RESET}"
echo -e "   Verify author mapping:"
git shortlog -se | sed 's/^/     /'
echo -e "\n   ${BOLD}Next verification steps:${RESET}"
echo -e "   1. Run full test suite: ${CYAN}go test -race ./...${RESET}"
echo -e "   2. Run simulated user gate: ${CYAN}./scripts/simulated_user_gate.sh${RESET}"
echo -e "   3. When ready to update remote (requires explicit user consent):"
echo -e "      ${BOLD}git push --force-with-lease origin main${RESET}\n"
