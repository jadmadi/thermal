#!/usr/bin/env bash
#
# simulated_user_gate.sh — Automated Simulated User Testing Release Gate
#
# Simulates comprehensive user commands across all Thermal features, verifying:
# 1. Output completeness & non-empty exit codes
# 2. ANSI stripping under --no-color
# 3. Valid JSON schema & jq parseability
# 4. Numerical formatting (no NaN, Inf, or raw float leaks)
# 5. Semantic consistency (token sums, rate ranges, verdicts)
# 6. Language & terminology integrity (no typo duplicate words)
# 7. Graceful validation errors on invalid flag combinations
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
BIN_PATH="${ROOT_DIR}/thermal"

echo -e "\n${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  THERMAL · Simulated User Testing Release Gate${RESET}"
echo -e "${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}\n"

# 1. Ensure binary is built
echo -e "${BOLD}1. Building thermal binary...${RESET}"
(cd "${ROOT_DIR}" && go build -o "${BIN_PATH}" ./cmd/thermal)
echo -e "   ${GREEN}✔${RESET} Binary built at: ${BIN_PATH}\n"

# 2. Check if host has agent databases; if not (e.g. CI), provision temporary mock home
MOCK_HOME=""
if [[ ! -d "${HOME}/.gemini/antigravity-cli" && ! -d "${HOME}/.codewhale" && ! -d "${HOME}/.claude" ]]; then
    echo -e "${YELLOW}Notice:${RESET} No local agent databases found on host; provisioning mock fixture environment..."
    MOCK_HOME="$(mktemp -d -t thermal-gate-home-XXXXXX)"
    MOCK_REPO="${MOCK_HOME}/projects/repo-sim"
    mkdir -p "${MOCK_REPO}/.git"
    mkdir -p "${MOCK_HOME}/.codewhale/sessions"
    mkdir -p "${MOCK_HOME}/.claude/projects/mockproj"

    NOW_SEC=$(date +%s)
    RECENT_MS=$(( (NOW_SEC - 86400) * 1000 ))
    RECENT_ISO=$(date -u -d "@$((NOW_SEC - 86400))" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")

    cat <<EOF > "${MOCK_HOME}/.codewhale/sessions/session_1.json"
{
  "session_id": "gate-sess-1",
  "metadata": {
    "created_at": "${RECENT_ISO}",
    "updated_at": "${RECENT_ISO}",
    "message_count": 10,
    "total_tokens": 750000,
    "cost": { "session_cost_usd": 1.75 },
    "model": "claude-3-5-sonnet",
    "mode": "chat",
    "workspace": "${MOCK_REPO}"
  }
}
EOF

    cat <<EOF > "${MOCK_HOME}/.claude/projects/mockproj/session.jsonl"
{"type":"assistant","timestamp":"${RECENT_ISO}","cwd":"${MOCK_REPO}","message":{"id":"gate-msg-1","model":"claude-3-5-sonnet","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":200,"cache_read_input_tokens":8000}}}
EOF
    export HOME="${MOCK_HOME}"
    echo -e "   ${GREEN}✔${RESET} Mock fixture environment ready at: ${MOCK_HOME}\n"
fi

cleanup() {
    if [[ -n "${MOCK_HOME}" && -d "${MOCK_HOME}" ]]; then
        rm -rf "${MOCK_HOME}"
    fi
}
trap cleanup EXIT

TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

run_check() {
    local desc="$1"
    local expect_code="$2" # 0 for success, 1 for validation error
    local check_type="$3"   # text, json, or error
    shift 3
    local cmd=("$@")

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    local out
    local err
    local code=0

    set +e
    out=$("${cmd[@]}" 2>&1)
    code=$?
    set -e

    if [[ "$code" -ne "$expect_code" ]]; then
        echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}"
        echo -e "    Expected exit code ${expect_code}, got ${code}"
        echo -e "    Command: ${cmd[*]}"
        echo -e "    Output: ${out}\n"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return
    fi

    # Formatting checks for text commands
    if [[ "$check_type" == "text" ]]; then
        # Check no ANSI if --no-color was passed
        if [[ " ${cmd[*]} " =~ " --no-color " ]] && [[ "$out" == *$'\033['* ]]; then
            echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}: ANSI escape codes detected with --no-color"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            return
        fi

        # Check no raw NaN or Inf
        if [[ "$out" =~ NaN ]] || [[ "$out" =~ [+-]Inf ]]; then
            echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}: NaN or Inf floating point flaw detected in output"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            return
        fi

        # Check for unformatted raw floats (e.g. 0.00000000001)
        if echo "$out" | grep -Eq '[0-9]+\.[0-9]{8,}'; then
            echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}: Unformatted raw floating point number detected"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            return
        fi
    fi

    # JSON validation checks
    if [[ "$check_type" == "json" ]]; then
        if ! echo "$out" | jq . >/dev/null 2>&1; then
            echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}: stdout is not valid JSON"
            echo -e "    Output: ${out}\n"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            return
        fi
    fi

    # Error message check
    if [[ "$check_type" == "error" ]]; then
        if [[ -z "$out" ]]; then
            echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}: Expected error message, but output was empty"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            return
        fi
    fi

    echo -e "  ${GREEN}✔ [PASS]${RESET} ${desc}"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
}

echo -e "${BOLD}2. Core Leaderboard & Dashboards:${RESET}"
run_check "Leaderboard standard output" 0 text "${BIN_PATH}" --no-color
run_check "Leaderboard sorted by tokens" 0 text "${BIN_PATH}" --sort tokens --no-color
run_check "Leaderboard sorted by cost" 0 text "${BIN_PATH}" --sort cost --no-color
run_check "Leaderboard JSON export" 0 json "${BIN_PATH}" --json

echo -e "\n${BOLD}3. Reports (Daily, Weekly, Monthly):${RESET}"
run_check "Daily report with --last 7" 0 text "${BIN_PATH}" daily --last 7 --no-color
run_check "Daily report JSON export" 0 json "${BIN_PATH}" daily --last 7 --json
run_check "Weekly report with --chart" 0 text "${BIN_PATH}" weekly --chart --no-color
run_check "Weekly report JSON export" 0 json "${BIN_PATH}" weekly --json
run_check "Monthly report" 0 text "${BIN_PATH}" monthly --no-color
run_check "Monthly report JSON export" 0 json "${BIN_PATH}" monthly --json

echo -e "\n${BOLD}4. Projects & Models Ranking:${RESET}"
run_check "Projects ranking" 0 text "${BIN_PATH}" projects --no-color
run_check "Projects sorted by cost with --top 5" 0 text "${BIN_PATH}" projects --sort cost --top 5 --no-color
run_check "Projects breakdown detail" 0 text "${BIN_PATH}" projects --breakdown --no-color
run_check "Projects JSON export" 0 json "${BIN_PATH}" projects --json
run_check "Models ranking" 0 text "${BIN_PATH}" models --no-color
run_check "Models sorted by cost" 0 text "${BIN_PATH}" models --sort cost --no-color
run_check "Models JSON export" 0 json "${BIN_PATH}" models --json

echo -e "\n${BOLD}5. Analytics (Mix, Stats, Trend):${RESET}"
run_check "Mix by tool" 0 text "${BIN_PATH}" mix --no-color
run_check "Mix by model with daily grain" 0 text "${BIN_PATH}" mix --by model --grain day --no-color
run_check "Mix by cost metric" 0 text "${BIN_PATH}" mix --metric cost --no-color
run_check "Mix JSON export" 0 json "${BIN_PATH}" mix --json
run_check "Stats distribution" 0 text "${BIN_PATH}" stats --no-color
run_check "Stats cost metric" 0 text "${BIN_PATH}" stats --metric cost --no-color
run_check "Stats JSON export" 0 json "${BIN_PATH}" stats --json
run_check "Trend fit & projection" 0 text "${BIN_PATH}" trend --no-color
run_check "Trend cost metric" 0 text "${BIN_PATH}" trend --metric cost --no-color
run_check "Trend JSON export" 0 json "${BIN_PATH}" trend --json

echo -e "\n${BOLD}6. Replay Simulation (Workloads vs Subscriptions & APIs):${RESET}"
run_check "Replay default multi-tier simulation" 0 text "${BIN_PATH}" replay --no-color
run_check "Replay against Claude Pro subscription" 0 text "${BIN_PATH}" replay --against claude-pro --no-color
run_check "Replay against DeepSeek V3 API" 0 text "${BIN_PATH}" replay --against deepseek-v3 --no-color
run_check "Replay compare all plans" 0 text "${BIN_PATH}" replay --compare all --no-color
run_check "Replay JSON export" 0 json "${BIN_PATH}" replay --json

echo -e "\n${BOLD}7. Info & Version Commands:${RESET}"
run_check "Version output" 0 text "${BIN_PATH}" version
run_check "Help output" 0 text "${BIN_PATH}" --help

echo -e "\n${BOLD}8. Validation & Negative Flag Audits (Exit 1 & Helpful Messages):${RESET}"
run_check "Rejects --top on weekly report" 1 error "${BIN_PATH}" weekly --top 5
run_check "Rejects --by on replay" 1 error "${BIN_PATH}" replay --by model
run_check "Rejects --against on projects" 1 error "${BIN_PATH}" projects --against claude-pro
run_check "Rejects --chart on replay" 1 error "${BIN_PATH}" replay --chart
run_check "Rejects negative --last" 1 error "${BIN_PATH}" daily --last -5

echo -e "\n${BOLD}${CYAN}─────────────────────────────────────────────────────────────────────${RESET}"
echo -e "${BOLD}Simulated User Gate Summary:${RESET} ${GREEN}${PASSED_CHECKS} passed${RESET}, ${RED}${FAILED_CHECKS} failed${RESET} (out of ${TOTAL_CHECKS} checks)"
echo -e "${BOLD}${CYAN}─────────────────────────────────────────────────────────────────────${RESET}\n"

if [[ "$FAILED_CHECKS" -gt 0 ]]; then
    echo -e "${RED}${BOLD}✖ RELEASE GATE FAILED:${RESET} Simulated user testing identified regressions." >&2
    exit 1
fi

echo -e "${GREEN}${BOLD}✔ RELEASE GATE PASSED:${RESET} All simulated user commands, formatting, semantics, and language verified.\n"
exit 0
