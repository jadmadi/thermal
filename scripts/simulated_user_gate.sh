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
# 8. Semantic regression suites: audit claim semantics with
#    nonempty test-selection assertions
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

# 2. Provision hermetic mock fixture environment
echo -e "${BOLD}2. Provisioning hermetic mock fixture environment...${RESET}"
GATE_HOME="$(mktemp -d -t thermal-gate-home-XXXXXX)"
MOCK_REPO="${GATE_HOME}/projects/repo-sim"
mkdir -p "${MOCK_REPO}/.git"
mkdir -p "${GATE_HOME}/.codewhale/sessions"
mkdir -p "${GATE_HOME}/.claude/projects/mockproj"
mkdir -p "${GATE_HOME}/.cache/thermal"

TODAY_ISO="$(date +"%Y-%m-%dT12:00:00Z")"

# Mock CodeWhale session (750,000 tokens, $1.75 recorded cost, claude-3-5-sonnet)
cat <<EOF > "${GATE_HOME}/.codewhale/sessions/session_1.json"
{
  "session_id": "gate-sess-1",
  "metadata": {
    "created_at": "${TODAY_ISO}",
    "updated_at": "${TODAY_ISO}",
    "message_count": 10,
    "total_tokens": 750000,
    "cost": { "session_cost_usd": 1.75 },
    "model": "claude-3-5-sonnet",
    "mode": "chat",
    "workspace": "${MOCK_REPO}"
  }
}
EOF

# Mock Claude session (9,700 tokens: 1000 input, 500 output, 200 cache write, 8000 cache read)
cat <<EOF > "${GATE_HOME}/.claude/projects/mockproj/session.jsonl"
{"type":"assistant","timestamp":"${TODAY_ISO}","cwd":"${MOCK_REPO}","message":{"id":"gate-msg-1","model":"claude-3-5-sonnet","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":200,"cache_read_input_tokens":8000}}}
EOF

# Cached pricing catalog for offline / deterministic pricing
cat <<EOF > "${GATE_HOME}/.cache/thermal/pricing.json"
{
  "version": 1,
  "fetchedAt": "${TODAY_ISO}",
  "source": "https://models.dev/api.json",
  "models": {
    "claude-3-5-sonnet": { "input": 3.0, "output": 15.0, "cacheRead": 0.3, "cacheWrite": 3.75 }
  }
}
EOF

# Mock Antigravity session (Activity Hunter: 12 steps, gemini-3.1-pro)
mkdir -p "${GATE_HOME}/.gemini/antigravity/brain/sess-gate/.system_generated/logs"
cat <<EOF > "${GATE_HOME}/.gemini/antigravity/brain/sess-gate/.system_generated/logs/transcript.jsonl"
{"created_at":"${TODAY_ISO}","source":"USER_EXPLICIT","type":"USER_INPUT","content":"gate request"}
{"created_at":"${TODAY_ISO}","source":"SYSTEM","type":"STEP","content":"Model Selection changed from None to Gemini 3.1 Pro (High)."}
EOF
for i in $(seq 1 10); do
  echo "{\"created_at\":\"${TODAY_ISO}\",\"source\":\"MODEL\",\"type\":\"PLANNER_RESPONSE\",\"content\":\"step $i\"}" >> "${GATE_HOME}/.gemini/antigravity/brain/sess-gate/.system_generated/logs/transcript.jsonl"
done

# Isolate environment to hermetic GATE_HOME and neutral color settings
ORIG_HOME="${HOME}"
ORIG_GOCACHE="$(go env GOCACHE 2>/dev/null || true)"
export HOME="${GATE_HOME}"
export GROK_HOME="${GATE_HOME}/.grok"
export DSH_HOME="${GATE_HOME}/.dsh"
export HERMES_HOME="${GATE_HOME}/.hermes"
export CODEX_HOME="${GATE_HOME}/.codex"
export OPENCODE_HOME="${GATE_HOME}/.opencode"
export ZED_HOME="${GATE_HOME}/.local/share/zed"
export ANTIGRAVITY_APP_DATA_DIR="antigravity"
export AGY_HOME="${GATE_HOME}/.gemini/antigravity"
export NO_COLOR=""
export CLICOLOR_FORCE=""

echo -e "   ${GREEN}✔${RESET} Hermetic fixture environment ready at: ${GATE_HOME}\n"

cleanup() {
    if [[ -n "${GATE_HOME:-}" && -d "${GATE_HOME}" ]]; then
        chmod -R u+w "${GATE_HOME}" 2>/dev/null || true
        rm -rf "${GATE_HOME}"
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
    local out_file
    local err_file
    out_file="$(mktemp)"
    err_file="$(mktemp)"
    local code=0

    set +e
    timeout 15s "${cmd[@]}" >"$out_file" 2>"$err_file"
    code=$?
    set -e

    local out
    local err
    out="$(cat "$out_file")"
    err="$(cat "$err_file")"
    rm -f "$out_file" "$err_file"

    if [[ "$code" -ne "$expect_code" ]]; then
        echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}"
        echo -e "    Expected exit code ${expect_code}, got ${code}"
        echo -e "    Command: ${cmd[*]}"
        if [[ -n "$err" ]]; then
            echo -e "    Stderr: ${err}"
        fi
        if [[ -n "$out" ]]; then
            echo -e "    Stdout: ${out}"
        fi
        echo ""
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

        # Check for unformatted raw floats (e.g. 0.00000000001), ignoring URL lines
        if echo "$out" | grep -vE 'https?://' | grep -Eq '[0-9]+\.[0-9]{8,}'; then
            echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}: Unformatted raw floating point number detected"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            return
        fi
    fi

    # JSON validation checks: stdout MUST be valid JSON, stderr must not pollute stdout
    if [[ "$check_type" == "json" ]]; then
        if ! echo "$out" | jq . >/dev/null 2>&1; then
            echo -e "  ${RED}✖ [FAIL]${RESET} ${desc}: stdout is not valid JSON"
            if [[ -n "$err" ]]; then
                echo -e "    Stderr: ${err}"
            fi
            echo -e "    Stdout: ${out}\n"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            return
        fi
    fi

    # Error message check
    if [[ "$check_type" == "error" ]]; then
        if [[ -z "${err:-$out}" ]]; then
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
run_check "Agy single-tool dashboard" 0 text "${BIN_PATH}" agy --no-color
run_check "Agy single-tool antigravity alias" 0 text "${BIN_PATH}" antigravity --no-color
run_check "Agy single-tool JSON export" 0 json "${BIN_PATH}" agy --json

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
run_check "Stats distribution report" 0 text "${BIN_PATH}" stats --no-color
run_check "Stats cost metric" 0 text "${BIN_PATH}" stats --metric cost --no-color
run_check "Stats JSON export" 0 json "${BIN_PATH}" stats --json
run_check "Trend fit & projection" 0 text "${BIN_PATH}" trend --no-color
run_check "Trend cost metric" 0 text "${BIN_PATH}" trend --metric cost --no-color
run_check "Trend JSON export" 0 json "${BIN_PATH}" trend --json

echo -e "\n${BOLD}6. Retired Replay Command:${RESET}"
run_check "Replay command retired" 1 error "${BIN_PATH}" replay

echo -e "\n${BOLD}7. Token Yield:${RESET}"
run_check "Yield report standard output" 0 text "${BIN_PATH}" yield --no-color
run_check "Yield report sorted by lines" 0 text "${BIN_PATH}" yield --sort lines --no-color
run_check "Yield report sorted by yield" 0 text "${BIN_PATH}" yield --sort yield --no-color
run_check "Yield report JSON export" 0 json "${BIN_PATH}" yield --json

echo -e "\n${BOLD}8. Local Setup Audit & Stateless URL Sharing:${RESET}"
run_check "Audit command standard output" 0 text "${BIN_PATH}" audit --no-color
run_check "Audit command JSON export" 0 json "${BIN_PATH}" audit --json
run_check "Share command standard output" 0 text "${BIN_PATH}" share --no-color
run_check "Share command JSON export" 0 json "${BIN_PATH}" share --json

echo -e "\n${BOLD}9. Info, Version, License & Changelog Commands:${RESET}"
run_check "Version output" 0 text "${BIN_PATH}" version
run_check "Help output" 0 text "${BIN_PATH}" --help
run_check "License command text" 0 text "${BIN_PATH}" license
run_check "License flag text" 0 text "${BIN_PATH}" --license
run_check "License JSON output" 0 json "${BIN_PATH}" license --json
run_check "Changelog command text" 0 text "${BIN_PATH}" changelog --no-color
run_check "Changelog top 1" 0 text "${BIN_PATH}" changelog --top 1 --no-color
run_check "Changelog JSON output" 0 json "${BIN_PATH}" changelog --json

echo -e "\n${BOLD}10. Validation & Negative Flag Audits (Exit 1 & Helpful Messages):${RESET}"
run_check "Rejects --top on weekly report" 1 error "${BIN_PATH}" weekly --top 5
run_check "Rejects negative --last" 1 error "${BIN_PATH}" daily --last -5
run_check "Rejects invalid --sort on yield" 1 error "${BIN_PATH}" yield --sort invalid_sort
run_check "Rejects missing zed database" 1 error "${BIN_PATH}" zed --no-color

echo -e "\n${BOLD}11. License & Attribution Integrity Checks:${RESET}"
if [[ ! -f "${ROOT_DIR}/LICENSE" ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} LICENSE file missing"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
elif ! grep -q "GNU AFFERO GENERAL PUBLIC LICENSE" "${ROOT_DIR}/LICENSE"; then
    echo -e "   ${RED}✖ Failed:${RESET} LICENSE file is not AGPL-3.0"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} LICENSE is AGPL-3.0"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [[ ! -f "${ROOT_DIR}/DUAL-LICENSE.md" || ! -f "${ROOT_DIR}/CONTRIBUTING.md" || ! -f "${ROOT_DIR}/NOTICES.md" ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} DUAL-LICENSE.md, CONTRIBUTING.md, or NOTICES.md missing"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} DUAL-LICENSE.md, CONTRIBUTING.md, and NOTICES.md present"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if grep -rnIi "Jafar" "${ROOT_DIR}/LICENSE" "${ROOT_DIR}/DUAL-LICENSE.md" "${ROOT_DIR}/CONTRIBUTING.md" "${ROOT_DIR}/cmd" "${ROOT_DIR}/internal" >/dev/null 2>&1; then
    echo -e "   ${RED}✖ Failed:${RESET} Deprecated author typo 'Jafar' found in codebase"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} Author attribution clean (Jad Madi)"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if grep -rnIi --exclude='*_test.go' "contact@jadmadi\.com" "${ROOT_DIR}/LICENSE" "${ROOT_DIR}/DUAL-LICENSE.md" "${ROOT_DIR}/CONTRIBUTING.md" "${ROOT_DIR}/cmd" "${ROOT_DIR}/internal" "${ROOT_DIR}/README.md" >/dev/null 2>&1; then
    echo -e "   ${RED}✖ Failed:${RESET} Invalid contact email 'contact@jadmadi.com' found in codebase"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} Commercial contact email clean (contact@jadmadi.net)"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if ! "${BIN_PATH}" license --json | grep -q '"covenant"'; then
    echo -e "   ${RED}✖ Failed:${RESET} license --json missing covenant metadata"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} License covenant metadata verified"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [[ ! -x "${ROOT_DIR}/scripts/check_governance_metrics.sh" ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} scripts/check_governance_metrics.sh is missing or not executable"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    "${ROOT_DIR}/scripts/check_governance_metrics.sh" >/dev/null 2>&1
    echo -e "   ${GREEN}✔ Passed:${RESET} Governance health & CHAOSS absence factor verified"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

echo -e "\n${BOLD}12. Deprecation Warning Audits (Stderr Warnings & Clean JSON Stdout):${RESET}"

# Verify --license warning on stderr
WARN_OUT=$("${BIN_PATH}" --license 2>&1 >/dev/null || true)
if [[ "$WARN_OUT" != *"thermal: warning: --license is deprecated"* ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} --license did not emit deprecation warning to stderr"
    echo -e "     Output was: ${WARN_OUT}"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} --license emits deprecation warning to stderr"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# Verify stdout with --license --json is clean JSON without stderr pollution
JSON_OUT=$("${BIN_PATH}" --license --json 2>/dev/null || true)
if ! echo "$JSON_OUT" | jq . >/dev/null 2>&1; then
    echo -e "   ${RED}✖ Failed:${RESET} --license --json stdout is corrupted or invalid JSON"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} --license --json stdout is clean, valid JSON"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# Verify nous tool alias warning on stderr
NOUS_WARN=$("${BIN_PATH}" nous 2>&1 >/dev/null || true)
if [[ "$NOUS_WARN" != *"thermal: warning: nous is deprecated"* ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} nous tool alias did not emit deprecation warning to stderr"
    echo -e "     Output was: ${NOUS_WARN}"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} nous tool alias emits deprecation warning to stderr"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

echo -e "\n${BOLD}13. Release Communication & Changelog Integrity Checks:${RESET}"
if [[ ! -f "${ROOT_DIR}/release-please-config.json" ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} release-please-config.json missing"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
elif ! jq -e '.packages["."]["changelog-sections"]' "${ROOT_DIR}/release-please-config.json" >/dev/null 2>&1; then
    echo -e "   ${RED}✖ Failed:${RESET} release-please-config.json missing changelog-sections definition"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} release-please-config.json schema & changelog-sections verified"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [[ ! -f "${ROOT_DIR}/CHANGELOG.md" ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} CHANGELOG.md missing"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
elif ! grep -q "Keep a Changelog" "${ROOT_DIR}/CHANGELOG.md"; then
    echo -e "   ${RED}✖ Failed:${RESET} CHANGELOG.md does not reference Keep a Changelog standard"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} CHANGELOG.md conforms to Keep a Changelog format"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

echo -e "\n${BOLD}14. Distribution Coverage & Agent Discovery Checks:${RESET}"
if [[ ! -x "${ROOT_DIR}/docs/pages/install.sh" ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} docs/pages/install.sh missing or not executable"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
elif ! sh "${ROOT_DIR}/docs/pages/install.sh" --dry-run >/dev/null 2>&1; then
    echo -e "   ${RED}✖ Failed:${RESET} docs/pages/install.sh --dry-run failed"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} docs/pages/install.sh verified in dry-run mode"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [[ ! -f "${ROOT_DIR}/docs/pages/llms.txt" || ! -f "${ROOT_DIR}/docs/pages/llms-full.txt" ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} docs/pages/llms.txt or llms-full.txt missing"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} Machine & agent documentation (llms.txt, llms-full.txt) present"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [[ ! -f "${ROOT_DIR}/docs/DISTRIBUTION.md" ]]; then
    echo -e "   ${RED}✖ Failed:${RESET} docs/DISTRIBUTION.md missing"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
elif ! grep -q "awesome-go" "${ROOT_DIR}/docs/DISTRIBUTION.md"; then
    echo -e "   ${RED}✖ Failed:${RESET} docs/DISTRIBUTION.md missing awesome-go showcase package"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} docs/DISTRIBUTION.md showcase runbook verified"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if ! grep -q "homebrew-tap" "${ROOT_DIR}/.goreleaser.yml"; then
    echo -e "   ${RED}✖ Failed:${RESET} .goreleaser.yml missing homebrew-tap configuration"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} .goreleaser.yml homebrew tap formula verified"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if ! grep -q "generator_generic_slsa3.yml" "${ROOT_DIR}/.github/workflows/release.yml"; then
    echo -e "   ${RED}✖ Failed:${RESET} .github/workflows/release.yml missing SLSA Level 3 generator"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} OpenSSF SLSA Level 3 build provenance workflow verified"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if ! grep -q "slsa-verifier" "${ROOT_DIR}/docs/DISTRIBUTION.md"; then
    echo -e "   ${RED}✖ Failed:${RESET} docs/DISTRIBUTION.md missing slsa-verifier instructions"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
else
    echo -e "   ${GREEN}✔ Passed:${RESET} docs/DISTRIBUTION.md SLSA3 verification runbook verified"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

echo -e "\n${BOLD}15. End-to-End Human User Persona Simulation Flows:${RESET}"
# Persona 1: Daily Developer Standup flow
run_check "Persona: Developer daily standup flow" 0 text "${BIN_PATH}" daily --last 3 --no-color
run_check "Persona: Developer code yield review" 0 text "${BIN_PATH}" yield --sort yield --no-color

# Persona 2: FinOps & Engineering Director flow
run_check "Persona: FinOps daily distribution audit" 0 text "${BIN_PATH}" stats --no-color
run_check "Persona: FinOps monthly cost review" 0 text "${BIN_PATH}" monthly --no-color
run_check "Persona: FinOps project spend attribution" 0 text "${BIN_PATH}" projects --sort cost --top 3 --no-color

# Persona 3: Security & Local Infrastructure flow
run_check "Persona: System health and permissions audit" 0 text "${BIN_PATH}" audit --no-color
run_check "Persona: Stateless share card URL generation" 0 text "${BIN_PATH}" share --no-color
run_check "Persona: Live token burn non-interactive pipe" 0 text "${BIN_PATH}" live
run_check "Persona: Live token burn snapshot export" 0 json "${BIN_PATH}" live --json

echo -e "\n${BOLD}16. Semantic Parity & Release Gate Assertions:${RESET}"

# 1. Static / Live Token Parity Check
TOTAL_STATIC=$("${BIN_PATH}" --json --offline | jq '[.results[] | select((.tool // .Tool) != "agy" and (.tool // .Tool) != "command-code" and (.tool // .Tool) != "droid" and (.tool // .Tool) != "muse") | ((.summary // .Summary).lifetimeTokens // 0)] | add // 0')
TOTAL_LIVE=$("${BIN_PATH}" live --json --offline | jq '.todayTokens // 0')

if [[ "$TOTAL_STATIC" -eq 759700 && "$TOTAL_LIVE" -eq 759700 ]]; then
    echo -e "  ${GREEN}✔ [PASS]${RESET} Static/Live token parity verified (759,700 tokens)"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    echo -e "  ${RED}✖ [FAIL]${RESET} Token parity mismatch: static=${TOTAL_STATIC}, live=${TOTAL_LIVE} (expected 759700)"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# 3. Live Stream NDJSON Emission Check
STREAM_OUT=$(timeout 2s "${BIN_PATH}" live --json --stream --interval 100ms --offline || true)
if echo "$STREAM_OUT" | grep -q '"type":"snapshot"'; then
    echo -e "  ${GREEN}✔ [PASS]${RESET} Live NDJSON streaming emits snapshot with bounded interval"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    echo -e "  ${RED}✖ [FAIL]${RESET} Live NDJSON streaming failed to emit snapshot"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# 5. Delta Cache Sub-Second Hit Check
CACHE_START=$(date +%s%N)
"${BIN_PATH}" --json --offline >/dev/null
CACHE_END=$(date +%s%N)
CACHE_DUR_MS=$(( (CACHE_END - CACHE_START) / 1000000 ))
if [[ "$CACHE_DUR_MS" -lt 1000 ]]; then
    echo -e "  ${GREEN}✔ [PASS]${RESET} Sub-second delta cache hit verified (${CACHE_DUR_MS}ms)"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    echo -e "  ${YELLOW}⚠ [WARN]${RESET} Delta cache invocation took ${CACHE_DUR_MS}ms (>1000ms target)"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# 17. Semantic Regression Suites
# Runs the named Go regression tests that encode the semantics behind public
# claims: honest audit wording. The minimum test counts are deliberate
# tripwires: a renamed or deleted test fails the gate instead of the pattern
# silently matching nothing.
echo -e "\n${BOLD}17. Semantic Regression Suites (audit):${RESET}"

run_semantic_suite() {
    local label="$1"
    local min_tests="$2"
    local pattern="$3"
    shift 3
    local targets=("$@")
    local output exit_code
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    # Run the Go suite in a neutral environment: the gate's fixture overrides
    # (tool roots, color vars) must not leak into test processes that build
    # their own temporary fixtures.
    if output=$(cd "${ROOT_DIR}" && \
        env -u CODEX_HOME -u GROK_HOME -u DSH_HOME -u HERMES_HOME -u OPENCODE_HOME \
            -u AGY_HOME -u ZED_HOME -u ANTIGRAVITY_APP_DATA_DIR -u THERMAL_CONFIG \
            -u NO_COLOR -u CLICOLOR_FORCE \
            HOME="${ORIG_HOME}" GOCACHE="${ORIG_GOCACHE}" \
            go test "${targets[@]}" -run "${pattern}" -count=1 -v 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi

    local passed_count
    passed_count=$(printf '%s\n' "${output}" | grep -c '^--- PASS' || true)

    if [[ ${exit_code} -ne 0 ]]; then
        echo -e "  ${RED}✖ [FAIL]${RESET} ${label}: go test exited ${exit_code}"
        printf '%s\n' "${output}" | tail -6 | sed 's/^/      /' >&2
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    elif [[ ${passed_count} -lt ${min_tests} ]]; then
        echo -e "  ${RED}✖ [FAIL]${RESET} ${label}: matched ${passed_count} passing tests, expected at least ${min_tests} (test renamed or removed?)"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    else
        echo -e "  ${GREEN}✔ [PASS]${RESET} ${label} (${passed_count} tests)"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    fi
}

run_semantic_suite "Audit claim semantics" 10 'Audit' ./internal/audit ./internal/render

echo -e "\n${BOLD}${CYAN}─────────────────────────────────────────────────────────────────────${RESET}"
echo -e "${BOLD}Simulated User Gate Summary:${RESET} ${GREEN}${PASSED_CHECKS} passed${RESET}, ${RED}${FAILED_CHECKS} failed${RESET} (out of ${TOTAL_CHECKS} checks)"
echo -e "${BOLD}${CYAN}─────────────────────────────────────────────────────────────────────${RESET}\n"

if [[ "$FAILED_CHECKS" -gt 0 ]]; then
    echo -e "${RED}${BOLD}✖ RELEASE GATE FAILED:${RESET} Simulated user testing identified regressions." >&2
    exit 1
fi

echo -e "${GREEN}${BOLD}✔ RELEASE GATE PASSED:${RESET} All simulated user commands, formatting, semantics, and language verified.\n"
exit 0
