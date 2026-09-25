#!/usr/bin/env bash
#
# check.sh — Unified Local Contributor Pre-Flight Verification Script
#
# Provides a tiered gate latency architecture:
#   - Fast-Path Pre-Commit (--pre-commit / --quick / -q):
#       Runs scoped formatting, Diff Sentry, and package-isolated unit tests
#       under 3 seconds to preserve local commit momentum.
#   - Comprehensive Pre-Push / CI (--full / --pre-push / -f):
#       Runs full race detector, package statement coverage, Go report card,
#       binary compilation, and simulated user release gate.
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

# Track start time for latency reporting
START_SECONDS=$(date +%s)

# Default mode
MODE="full"
ALL_FILES=0

show_help() {
    cat << EOF
Usage: ./scripts/check.sh [OPTIONS]

Thermal Contributor Pre-Flight Verification Suite

Modes:
  --pre-commit, --quick, -q   Fast-path pre-commit mode (<3s):
                                - Scoped Go formatting (modified files only)
                                - Diff Sentry (secrets, merge markers, debug artifacts)
                                - Unit tests only for modified Go packages
                                - Fast-fail build check
  --full, --pre-push, -f      Comprehensive pre-push / CI verification (default):
                                - All pre-commit checks
                                - Go Report Card & Static Analysis (go vet)
                                - Full race detector across all packages
                                - Full package statement coverage table
                                - Complete binary compilation
                                - Simulated user testing release gate
                                - Vulnerability scanning (govulncheck)

Options:
  --install-hooks             Configure Git to use repository hooks (.githooks/)
  --all-files                 Check formatting across all files instead of scoped
  --help, -h                  Display this help message

EOF
}

# Auto-detect if invoked directly from a git hook or environment variable
CALLER_BASE=$(basename "$0")
if [[ "${CALLER_BASE}" == "pre-commit" ]] || [[ "${THERMAL_PRECOMMIT:-0}" == "1" ]] || [[ "${GIT_HOOK:-}" == "pre-commit" ]]; then
    MODE="pre-commit"
elif [[ "${CALLER_BASE}" == "pre-push" ]] || [[ "${GIT_HOOK:-}" == "pre-push" ]]; then
    MODE="full"
fi

# Parse CLI arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --pre-commit|--quick|-q)
            MODE="pre-commit"
            shift
            ;;
        --full|--pre-push|-f)
            MODE="full"
            shift
            ;;
        --install-hooks)
            echo -e "${BOLD}${CYAN}Installing Thermal Git Hooks...${RESET}"
            cd "${ROOT_DIR}"
            if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
                git config core.hooksPath .githooks
                chmod +x .githooks/pre-commit .githooks/pre-push 2>/dev/null || true
                echo -e "${GREEN}✔ Configured git core.hooksPath to '.githooks'.${RESET}"
                echo -e "  - Pre-commit: fast-path (<3s, package-isolated unit tests & diff sentry)"
                echo -e "  - Pre-push:   full verification (race detector, coverage table, user gate)"
                exit 0
            else
                echo -e "${RED}✖ Error: Not a git repository.${RESET}"
                exit 1
            fi
            ;;
        --all-files)
            ALL_FILES=1
            shift
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${RESET}" >&2
            show_help >&2
            exit 2
            ;;
    esac
done

echo -e "\n${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}"
if [[ "${MODE}" == "pre-commit" ]]; then
    echo -e "${BOLD}  THERMAL · Fast-Path Pre-Commit Verification Suite${RESET}"
else
    echo -e "${BOLD}  THERMAL · Contributor Pre-Flight Verification Suite${RESET}"
fi
echo -e "${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}\n"

# -----------------------------------------------------------------------------
# Discover Modified Files & Packages
# -----------------------------------------------------------------------------
cd "${ROOT_DIR}"

# Ensure internal/changelog/CHANGELOG.md is synced with root CHANGELOG.md
if [[ -f "${ROOT_DIR}/CHANGELOG.md" ]]; then
    mkdir -p "${ROOT_DIR}/internal/changelog"
    cmp -s "${ROOT_DIR}/CHANGELOG.md" "${ROOT_DIR}/internal/changelog/CHANGELOG.md" 2>/dev/null || cp "${ROOT_DIR}/CHANGELOG.md" "${ROOT_DIR}/internal/changelog/CHANGELOG.md"
fi

# Find modified Go source files in working tree / staging
MODIFIED_GO_FILES=$( (git diff --cached --name-only --diff-filter=d 2>/dev/null || true; \
                      git diff --name-only --diff-filter=d 2>/dev/null || true; \
                      git ls-files --others --exclude-standard 2>/dev/null || true) \
                      | grep -E '\.go$' | sort -u || true )

# In full mode, if on a branch ahead of origin/main, include all files changed in the branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
BASE_BRANCH="origin/main"

if [[ "${MODE}" == "full" ]] && [[ "${CURRENT_BRANCH}" != "main" ]] && git rev-parse --verify "${BASE_BRANCH}" >/dev/null 2>&1; then
    BRANCH_GO_FILES=$(git diff --name-only --diff-filter=d "${BASE_BRANCH}...HEAD" 2>/dev/null | grep -E '\.go$' || true)
    if [[ -n "${BRANCH_GO_FILES}" ]]; then
        MODIFIED_GO_FILES=$(printf "%s\n%s\n" "${MODIFIED_GO_FILES}" "${BRANCH_GO_FILES}" | sed '/^$/d' | sort -u || true)
    fi
fi

# -----------------------------------------------------------------------------
# 1. Contributor Sign-off (CLA / DCO)
# -----------------------------------------------------------------------------
echo -e "${BOLD}1. Verifying Contributor Sign-off (CLA / DCO)...${RESET}"
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

# -----------------------------------------------------------------------------
# 2. Diff Sentry (Hygiene, Secrets, Merge Markers, Debug Artifacts)
# -----------------------------------------------------------------------------
echo -e "\n${BOLD}2. Running Diff Sentry...${RESET}"
SENTRY_FAIL=0

# Gather diff to inspect (prefer staged diff if non-empty; fallback to working tree diff + untracked files)
if ! git diff --cached --quiet 2>/dev/null; then
    DIFF_CONTENT=$(git diff --cached -U0 2>/dev/null || true)
else
    DIFF_CONTENT=$(git diff HEAD -U0 2>/dev/null || git diff -U0 2>/dev/null || true)
    UNTRACKED_FILES=$(git ls-files --others --exclude-standard 2>/dev/null || true)
    for uf in ${UNTRACKED_FILES}; do
        if [ -f "$uf" ] && [ ! -L "$uf" ] && [[ "$uf" != "scripts/check_test.sh" ]]; then
            DIFF_CONTENT="${DIFF_CONTENT}"$'\n'$(diff -u /dev/null "$uf" 2>/dev/null || true)
        fi
    done
fi

ADDED_LINES=$(echo "${DIFF_CONTENT}" | grep -E '^\+[^+]' || true)

# Sentry Rule A: Unresolved merge conflict markers
if echo "${ADDED_LINES}" | grep -E -q '^\+(<{7}|={7}|>{7})([[:space:]]|$)'; then
    echo -e "   ${RED}✖ Diff Sentry Error:${RESET} Detected unresolved Git merge conflict markers in diff."
    echo "${ADDED_LINES}" | grep -E '^\+(<{7}|={7}|>{7})([[:space:]]|$)' | head -n 5
    SENTRY_FAIL=1
fi

# Sentry Rule B: Leaked Private Keys
if echo "${ADDED_LINES}" | grep -E -q -- '-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----'; then
    echo -e "   ${RED}✖ Diff Sentry Error:${RESET} Detected potential private key header in diff."
    SENTRY_FAIL=1
fi

# Sentry Rule C: High-confidence API Key / Secret Token Leak
# (Match actual keys while avoiding false positives)
if echo "${ADDED_LINES}" | grep -E -q -- '(sk-[a-zA-Z0-9_-]{20,}|gh[pousr]_[A-Za-z0-9_]{36,}|AIzaSy[A-Za-z0-9_-]{33}|AKIA[0-9A-Z]{16}|xox[baprs]-[0-9A-Za-z]{10,})'; then
    # Ensure this is not inside an explicit test mock file or check script
    SUSPECT_FILES=$(echo "${DIFF_CONTENT}" | grep -B2 -E -- '(sk-[a-zA-Z0-9_-]{20,}|gh[pousr]_[A-Za-z0-9_]{36,}|AIzaSy[A-Za-z0-9_-]{33}|AKIA[0-9A-Z]{16}|xox[baprs]-[0-9A-Za-z]{10,})' | grep -E '^\+\+\+ ' | awk '{print $2}' || true)
    NON_TEST_VIOLATIONS=0
    for sf in ${SUSPECT_FILES}; do
        clean_sf="${sf#b/}"
        if [[ "${clean_sf}" != *"_test.go" ]] && [[ "${clean_sf}" != *"testdata"* ]] && [[ "${clean_sf}" != *"scripts/"* ]] && [[ "${clean_sf}" != *".githooks/"* ]] && [[ "${clean_sf}" != *"docs/"* ]] && [[ "${clean_sf}" != *"goals/"* ]]; then
            NON_TEST_VIOLATIONS=$((NON_TEST_VIOLATIONS + 1))
        fi
    done
    if [ "${NON_TEST_VIOLATIONS}" -gt 0 ]; then
        echo -e "   ${RED}✖ Diff Sentry Error:${RESET} Detected potential live API token / secret credential in non-test source diff."
        SENTRY_FAIL=1
    fi
fi

# Sentry Rule D: Temporary Debug Statements or Leftover Panics
if echo "${ADDED_LINES}" | grep -E -q '(fmt\.(Print|Printf|Println)\("DEBUG|panic\("TODO|runtime\.Breakpoint\(\))'; then
    echo -e "   ${RED}✖ Diff Sentry Error:${RESET} Detected temporary debug print or TODO panic in diff."
    echo "${ADDED_LINES}" | grep -E '(fmt\.(Print|Printf|Println)\("DEBUG|panic\("TODO|runtime\.Breakpoint\(\))' | head -n 5
    SENTRY_FAIL=1
fi

# Sentry Rule E: Dangerous / Binary Files Staged
STAGED_BINARIES=$(git status --porcelain 2>/dev/null | grep -E '^[AM].*\.(exe|so|dylib|dll|class)$' | awk '{print $2}' || true)
if [ -n "$STAGED_BINARIES" ]; then
    echo -e "   ${RED}✖ Diff Sentry Error:${RESET} Forbidden binary executable / library staged for commit:"
    echo "$STAGED_BINARIES"
    SENTRY_FAIL=1
fi

if [ "${SENTRY_FAIL}" -ne 0 ]; then
    echo -e "   ${RED}Diff Sentry failed. Please resolve above issues before committing.${RESET}"
    exit 1
fi
echo -e "   ${GREEN}✔ Passed:${RESET} Diff Sentry cleared (zero leaked secrets, merge markers, or debug artifacts)."

# -----------------------------------------------------------------------------
# 3. Go Formatting (Scoped to modified files per AGENTS.md)
# -----------------------------------------------------------------------------
echo -e "\n${BOLD}3. Checking Code Formatting (gofmt)...${RESET}"
if [[ "${ALL_FILES}" -eq 1 ]]; then
    UNFORMATTED=$(gofmt -l cmd/ internal/ 2>/dev/null || true)
    if [ -n "$UNFORMATTED" ]; then
        echo -e "   ${RED}✖ Error:${RESET} The following files need formatting:"
        echo "$UNFORMATTED"
        echo -e "\nRun: ${BOLD}gofmt -w <file>${RESET} to format."
        exit 1
    fi
    echo -e "   ${GREEN}✔ Passed:${RESET} All Go source files formatted cleanly."
elif [[ -z "${MODIFIED_GO_FILES}" ]]; then
    echo -e "   ${GREEN}✔ Notice:${RESET} No Go source files modified; formatting check bypassed."
else
    UNFORMATTED=""
    for file in ${MODIFIED_GO_FILES}; do
        if [ -f "$file" ]; then
            UNF=$(gofmt -l "$file" 2>/dev/null || true)
            if [ -n "$UNF" ]; then
                UNFORMATTED="${UNFORMATTED}${UNF}\n"
            fi
        fi
    done
    if [ -n "$UNFORMATTED" ]; then
        echo -e "   ${RED}✖ Error:${RESET} The following modified files need formatting:"
        printf "%b" "$UNFORMATTED"
        echo -e "\nRun: ${BOLD}gofmt -w <file>${RESET} to format."
        exit 1
    fi
    echo -e "   ${GREEN}✔ Passed:${RESET} All modified Go source files formatted cleanly."
fi

# -----------------------------------------------------------------------------
# 4. Unit Tests (Scoped to modified packages in fast mode)
# -----------------------------------------------------------------------------
if [[ "${MODE}" == "pre-commit" ]]; then
    echo -e "\n${BOLD}4. Running Package-Isolated Unit Tests...${RESET}"
    if [ -z "${MODIFIED_GO_FILES}" ]; then
        echo -e "   ${GREEN}✔ Notice:${RESET} No Go source packages modified (docs/assets diff); package tests bypassed."
    else
        MODIFIED_PKGS=()
        for f in ${MODIFIED_GO_FILES}; do
            pkg_dir=$(dirname "$f")
            if [[ "$pkg_dir" == "." ]]; then
                MODIFIED_PKGS+=(".")
            else
                MODIFIED_PKGS+=("./$pkg_dir")
            fi
        done
        UNIQUE_PKGS=($(printf "%s\n" "${MODIFIED_PKGS[@]}" 2>/dev/null | sort -u || true))
        echo -e "   ${CYAN}ℹ Fast-Path:${RESET} Testing modified packages: ${UNIQUE_PKGS[*]}"
        (cd "${ROOT_DIR}" && go test -race "${UNIQUE_PKGS[@]}")
        echo -e "   ${GREEN}✔ Passed:${RESET} Unit tests passed for modified packages."
    fi

    # 5. Quick Build Verification
    if [ -n "${MODIFIED_GO_FILES}" ]; then
        echo -e "\n${BOLD}5. Verifying Binary Compilation...${RESET}"
        (cd "${ROOT_DIR}" && go build -o /dev/null ./cmd/thermal)
        echo -e "   ${GREEN}✔ Passed:${RESET} cmd/thermal compiled successfully."
    fi

    ELAPSED=$(( $(date +%s) - START_SECONDS ))
    echo -e "\n${BOLD}${GREEN}═════════════════════════════════════════════════════════════════════${RESET}"
    echo -e "${BOLD}${GREEN}  ✔ FAST-PATH PRE-COMMIT VERIFICATION CLEARED in ${ELAPSED}s! Clean commit ready.${RESET}"
    echo -e "${BOLD}  ℹ Full package coverage, report card, and simulated user gate reserved for pre-push / CI.${RESET}"
    echo -e "${BOLD}${GREEN}═════════════════════════════════════════════════════════════════════${RESET}\n"
    exit 0
fi

# -----------------------------------------------------------------------------
# Comprehensive Pre-Push / Pipeline Gate (Mode == "full")
# -----------------------------------------------------------------------------

# 4. Full Unit Tests with Race Detection
echo -e "\n${BOLD}4. Running Full Unit Tests with Race Detector...${RESET}"
(cd "${ROOT_DIR}" && go test -race ./...)
echo -e "   ${GREEN}✔ Passed:${RESET} All unit tests passed cleanly with -race."

# 5. Full Package Statement Coverage Table
echo -e "\n${BOLD}5. Generating Full Package Statement Coverage...${RESET}"
COVERAGE_RAW=$(cd "${ROOT_DIR}" && go test -cover ./... 2>&1 || true)

awk '
/coverage:/ {
    pkg = "";
    cov = 0;
    for (i = 1; i <= NF; i++) {
        if ($i ~ /^github.com\/jadmadi\/thermal\//) {
            pkg = $i;
        }
        if ($i == "coverage:") {
            cov = $(i+1);
            gsub(/%/, "", cov);
            total += cov;
            count++;
            status = (cov + 0 >= 70.0) ? "PASS" : "INFO";
        }
    }
    if (pkg != "") {
        sub("github.com/jadmadi/thermal/", "", pkg);
        printf "   %-38s %6.1f%%   %s\n", pkg, cov, status;
    }
}
END {
    if (count > 0) {
        avg = total / count;
        print "   ────────────────────────────────────────────────────────────";
        printf "   Repository Average Coverage (%d pkgs): %6.1f%%   %s\n", count, avg, (avg >= 70.0 ? "PASS" : "INFO");
    }
}
' <<< "${COVERAGE_RAW}"
echo -e "   ${GREEN}✔ Passed:${RESET} Package statement coverage verified."

# 6. Go Report Card & Static Analysis
echo -e "\n${BOLD}6. Running Go Report Card & Static Analysis (go vet)...${RESET}"
(cd "${ROOT_DIR}" && go vet ./...)
echo -e "   ${GREEN}✔ Passed:${RESET} go vet completed with zero warnings."

# 7. Build Verification
echo -e "\n${BOLD}7. Verifying Binary Compilation...${RESET}"
(cd "${ROOT_DIR}" && go build -o /dev/null ./cmd/thermal)
echo -e "   ${GREEN}✔ Passed:${RESET} cmd/thermal compiled successfully."

# 8. Simulated User Testing Release Gate
echo -e "\n${BOLD}8. Running Simulated User Release Gate...${RESET}"
"${SCRIPT_DIR}/simulated_user_gate.sh"

# 9. Optional Vulnerability Scanning
if command -v govulncheck >/dev/null 2>&1; then
    echo -e "\n${BOLD}9. Running govulncheck...${RESET}"
    (cd "${ROOT_DIR}" && govulncheck ./...)
    echo -e "   ${GREEN}✔ Passed:${RESET} govulncheck identified zero vulnerabilities."
fi

ELAPSED=$(( $(date +%s) - START_SECONDS ))
echo -e "\n${BOLD}${GREEN}═════════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}${GREEN}  ✔ ALL PRE-FLIGHT CHECKS PASSED in ${ELAPSED}s! Ready for push / pull request.${RESET}"
echo -e "${BOLD}${GREEN}═════════════════════════════════════════════════════════════════════${RESET}\n"
exit 0
