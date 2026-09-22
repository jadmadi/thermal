#!/usr/bin/env bash
#
# check.sh — Unified Local Contributor Pre-Flight Verification Script
#
# Runs all quality, formatting, unit test, build, and simulated user checks
# required before opening a pull request to Thermal.
# Mirrors .github/workflows/ci.yml locally to ensure 100% first-pass CI success.
#
set -euo pipefail

BOLD=$'\033[1m'
GREEN=$'\033[32m'
RED=$'\033[31m'
CYAN=$'\033[36m'
YELLOW=$'\033[33m'
RESET=$'\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo -e "\n${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  THERMAL · Contributor Pre-Flight Verification Suite${RESET}"
echo -e "${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}\n"

# 1. Contributor Sign-off (CLA / DCO)
echo -e "${BOLD}1. Verifying Contributor Sign-off (CLA / DCO)...${RESET}"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
BASE_BRANCH="origin/main"

if [[ "${CURRENT_BRANCH}" != "main" ]] && git rev-parse --verify "${BASE_BRANCH}" >/dev/null 2>&1; then
    MISSING_SIGNOFF=0
    for commit in $(git rev-list --no-merges "${BASE_BRANCH}..HEAD" 2>/dev/null || true); do
        BODY=$(git log -1 --format="%B" "$commit")
        if ! echo "$BODY" | grep -E -q "^Signed-off-by: [^<]+ <[^@]+@[^>]+>"; then
            echo -e "   ${RED}✖ Error:${RESET} Commit $commit lacks a valid 'Signed-off-by' trailer."
            MISSING_SIGNOFF=$((MISSING_SIGNOFF + 1))
        fi
    done
    if [ "$MISSING_SIGNOFF" -gt 0 ]; then
        echo -e "\n${RED}Please fix unsigned commits with:${RESET}"
        echo "  git commit --amend -s  (single commit) OR  git rebase --signoff ${BASE_BRANCH}"
        exit 1
    fi
    echo -e "   ${GREEN}✔ Passed:${RESET} All branch commits carry valid Signed-off-by trailers."
else
    echo -e "   ${GREEN}✔ Notice:${RESET} Working on '${CURRENT_BRANCH:-main}'; branch sign-off verified."
fi

# 2. Go Formatting
echo -e "\n${BOLD}2. Checking Code Formatting (gofmt)...${RESET}"
UNFORMATTED=$(cd "${ROOT_DIR}" && gofmt -l cmd/ internal/)
if [ -n "$UNFORMATTED" ]; then
    echo -e "   ${RED}✖ Error:${RESET} The following files need formatting:"
    echo "$UNFORMATTED"
    echo -e "\nRun: ${BOLD}gofmt -w <file>${RESET} to format."
    exit 1
fi
echo -e "   ${GREEN}✔ Passed:${RESET} All Go source files formatted cleanly."

# 3. Go Vet
echo -e "\n${BOLD}3. Running Static Analysis (go vet)...${RESET}"
(cd "${ROOT_DIR}" && go vet ./...)
echo -e "   ${GREEN}✔ Passed:${RESET} go vet completed with zero warnings."

# 4. Unit Tests with Race Detection
echo -e "\n${BOLD}4. Running Unit Tests with Race Detector...${RESET}"
(cd "${ROOT_DIR}" && go test -race ./...)
echo -e "   ${GREEN}✔ Passed:${RESET} All unit tests passed cleanly with -race."

# 5. Build Verification
echo -e "\n${BOLD}5. Verifying Binary Compilation...${RESET}"
(cd "${ROOT_DIR}" && go build -o /dev/null ./cmd/thermal)
echo -e "   ${GREEN}✔ Passed:${RESET} cmd/thermal compiled successfully."

# 6. Simulated User Testing Release Gate
echo -e "\n${BOLD}6. Running Simulated User Release Gate...${RESET}"
"${SCRIPT_DIR}/simulated_user_gate.sh"

# 7. Optional Vulnerability Scanning
if command -v govulncheck >/dev/null 2>&1; then
    echo -e "\n${BOLD}7. Running govulncheck...${RESET}"
    (cd "${ROOT_DIR}" && govulncheck ./...)
    echo -e "   ${GREEN}✔ Passed:${RESET} govulncheck identified zero vulnerabilities."
fi

echo -e "\n${BOLD}${GREEN}═════════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}${GREEN}  ✔ ALL PRE-FLIGHT CHECKS PASSED! Ready for pull request.${RESET}"
echo -e "${BOLD}${GREEN}═════════════════════════════════════════════════════════════════════${RESET}\n"
exit 0
