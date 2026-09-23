#!/usr/bin/env bash
#
# check_test.sh — Automated Test Suite for Tiered Gate Latency Architecture
#
# Verifies:
#   1. CLI flag parsing and help output
#   2. Fast-path pre-commit execution latency (<3s)
#   3. Scoped gofmt detection on unformatted modified files
#   4. Diff Sentry: Merge conflict marker detection
#   5. Diff Sentry: Leaked private key detection
#   6. Diff Sentry: Temporary debug print detection
#   7. Package isolation: Unit testing only touched Go packages
#   8. Docs-only bypass: Bypassing package unit tests for non-Go diffs
#   9. Git hook installation via --install-hooks
#
set -euo pipefail

BOLD=$'\033[1m'
GREEN=$'\033[32m'
RED=$'\033[31m'
CYAN=$'\033[36m'
RESET=$'\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHECK_SH="${ROOT_DIR}/scripts/check.sh"

PASSED=0
FAILED=0

assert_success() {
    local desc="$1"
    shift
    echo -n "• Testing: ${desc}... "
    if "$@" >/dev/null 2>&1; then
        echo -e "${GREEN}PASS${RESET}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}FAIL (expected exit 0)${RESET}"
        FAILED=$((FAILED + 1))
    fi
}

assert_failure() {
    local desc="$1"
    shift
    echo -n "• Testing: ${desc}... "
    if "$@" >/dev/null 2>&1; then
        echo -e "${RED}FAIL (expected non-zero exit)${RESET}"
        FAILED=$((FAILED + 1))
    else
        echo -e "${GREEN}PASS${RESET}"
        PASSED=$((PASSED + 1))
    fi
}

echo -e "\n${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  THERMAL · Check Suite & Tiered Gate Latency Verification${RESET}"
echo -e "${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}\n"

cd "${ROOT_DIR}"

# 1. Test CLI Help Output
echo -e "${BOLD}1. Verifying CLI Flags & Help${RESET}"
HELP_OUT=$("${CHECK_SH}" --help)
if echo "${HELP_OUT}" | grep -q -- "--pre-commit" && echo "${HELP_OUT}" | grep -q -- "--full"; then
    echo -e "• Testing: check.sh --help displays tiered modes... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: check.sh --help displays tiered modes... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

# 2. Test Fast-Path Pre-Commit Clean Latency (<3s)
echo -e "\n${BOLD}2. Verifying Fast-Path Latency on Clean State${RESET}"
START_TS=$(date +%s)
assert_success "check.sh --pre-commit exits 0" "${CHECK_SH}" --pre-commit
END_TS=$(date +%s)
DURATION=$(( END_TS - START_TS ))
if [ "${DURATION}" -le 3 ]; then
    echo -e "• Testing: Latency under 3 seconds (${DURATION}s)... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: Latency under 3 seconds (${DURATION}s)... ${RED}FAIL (${DURATION}s > 3s)${RESET}"
    FAILED=$((FAILED + 1))
fi

# 3. Test Scoped Formatting (Detects unformatted Go file)
echo -e "\n${BOLD}3. Verifying Scoped Code Formatting Detection${RESET}"
UNFORMATTED_FILE="${ROOT_DIR}/internal/pricing/temp_unformatted_test.go"
cat << 'EOF' > "${UNFORMATTED_FILE}"
package pricing

func UnformattedHelper( ) int {
return    42
}
EOF

UNF_OUT=$("${CHECK_SH}" --pre-commit 2>&1 || true)
rm -f "${UNFORMATTED_FILE}"
if echo "${UNF_OUT}" | grep -q "temp_unformatted_test.go" && echo "${UNF_OUT}" | grep -q "need formatting"; then
    echo -e "• Testing: Detects unformatted modified Go file... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: Detects unformatted modified Go file... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

# 4. Test Diff Sentry: Merge Conflict Markers
echo -e "\n${BOLD}4. Verifying Diff Sentry: Merge Conflict Marker Detection${RESET}"
TARGET_FILE="${ROOT_DIR}/internal/pricing/temp_conflict_test.go"
M_LT="<"
M_EQ="="
M_GT=">"
cat << EOF > "${TARGET_FILE}"
package pricing

${M_LT}${M_LT}${M_LT}${M_LT}${M_LT}${M_LT}${M_LT} HEAD
func ConflictHelper() int {
    return 1
}
${M_EQ}${M_EQ}${M_EQ}${M_EQ}${M_EQ}${M_EQ}${M_EQ}
func ConflictHelper() int {
    return 2
}
${M_GT}${M_GT}${M_GT}${M_GT}${M_GT}${M_GT}${M_GT} branch
EOF

CONFLICT_OUT=$("${CHECK_SH}" --pre-commit 2>&1 || true)
rm -f "${TARGET_FILE}"
if echo "${CONFLICT_OUT}" | grep -q "Detected unresolved Git merge conflict markers"; then
    echo -e "• Testing: Diff Sentry catches merge conflict markers... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: Diff Sentry catches merge conflict markers... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

# 5. Test Diff Sentry: Leaked Private Key
echo -e "\n${BOLD}5. Verifying Diff Sentry: Leaked Private Key Detection${RESET}"
KEY_FILE="${ROOT_DIR}/internal/pricing/temp_key_test.go"
K_BEG="-----BEGIN RSA "
K_END="PRIVATE KEY-----"
cat << EOF > "${KEY_FILE}"
package pricing

const LeakKey = \`${K_BEG}${K_END}
MIIEowIBAAKCAQEA...
-----END RSA PRIVATE KEY-----\`
EOF

KEY_OUT=$("${CHECK_SH}" --pre-commit 2>&1 || true)
rm -f "${KEY_FILE}"
if echo "${KEY_OUT}" | grep -q "Detected potential private key header"; then
    echo -e "• Testing: Diff Sentry catches leaked private key... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: Diff Sentry catches leaked private key... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

# 6. Test Diff Sentry: Temporary Debug Prints
echo -e "\n${BOLD}6. Verifying Diff Sentry: Debug Print Detection${RESET}"
DEBUG_FILE="${ROOT_DIR}/internal/pricing/temp_debug_test.go"
D_CALL="fmt.Println(\"DEBUG"
cat << EOF > "${DEBUG_FILE}"
package pricing

import "fmt"

func DebugHelper() {
	${D_CALL} temp logging")
}
EOF

DEBUG_OUT=$("${CHECK_SH}" --pre-commit 2>&1 || true)
rm -f "${DEBUG_FILE}"
if echo "${DEBUG_OUT}" | grep -q "Detected temporary debug print"; then
    echo -e "• Testing: Diff Sentry catches temporary debug prints... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: Diff Sentry catches temporary debug prints... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

# 7. Test Package Isolation: Only Tests Modified Package
echo -e "\n${BOLD}7. Verifying Package Isolation in Unit Testing${RESET}"
ISO_FILE="${ROOT_DIR}/internal/pricing/temp_isolated_test.go"
cat << 'EOF' > "${ISO_FILE}"
package pricing

import "testing"

func TestFastPathIsolated(t *testing.T) {
	if 1+1 != 2 {
		t.Fatal("math broken")
	}
}
EOF

ISO_OUT=$("${CHECK_SH}" --pre-commit 2>&1 || true)
rm -f "${ISO_FILE}"
if echo "${ISO_OUT}" | grep -q "Testing modified packages: ./internal/pricing" && ! echo "${ISO_OUT}" | grep -q "./cmd/thermal" && ! echo "${ISO_OUT}" | grep -q "./internal/tui"; then
    echo -e "• Testing: Isolates unit tests strictly to modified package... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: Isolates unit tests strictly to modified package... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

# 8. Test Docs-Only Bypass
echo -e "\n${BOLD}8. Verifying Docs/Assets-Only Bypass${RESET}"
DOC_FILE="${ROOT_DIR}/docs/temp_docs_test.md"
echo "# Temporary doc test" > "${DOC_FILE}"
DOCS_OUT=$("${CHECK_SH}" --pre-commit 2>&1 || true)
rm -f "${DOC_FILE}"
if echo "${DOCS_OUT}" | grep -q "No Go source packages modified (docs/assets diff); package tests bypassed"; then
    echo -e "• Testing: Bypasses package unit tests on docs-only diff... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: Bypasses package unit tests on docs-only diff... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

# 9. Test Git Hook Installation
echo -e "\n${BOLD}9. Verifying Git Hook Installation${RESET}"
INSTALL_OUT=$("${CHECK_SH}" --install-hooks 2>&1 || true)
CURRENT_HOOKS=$(git config --get core.hooksPath || echo "")
if [[ "${CURRENT_HOOKS}" == ".githooks" ]] && echo "${INSTALL_OUT}" | grep -q "Configured git core.hooksPath to '.githooks'"; then
    echo -e "• Testing: --install-hooks configures core.hooksPath... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: --install-hooks configures core.hooksPath... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

# 10. Test Hook Delegation (.githooks/pre-commit)
echo -e "\n${BOLD}10. Verifying .githooks/pre-commit Delegation${RESET}"
HOOK_OUT=$("${ROOT_DIR}/.githooks/pre-commit" 2>&1 || true)
if echo "${HOOK_OUT}" | grep -q "FAST-PATH PRE-COMMIT VERIFICATION CLEARED"; then
    echo -e "• Testing: .githooks/pre-commit executes fast-path... ${GREEN}PASS${RESET}"
    PASSED=$((PASSED + 1))
else
    echo -e "• Testing: .githooks/pre-commit executes fast-path... ${RED}FAIL${RESET}"
    FAILED=$((FAILED + 1))
fi

echo -e "\n${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  Verification Summary: ${GREEN}${PASSED} passed${RESET}, ${RED}${FAILED} failed${RESET}"
echo -e "${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}\n"

if [ "${FAILED}" -gt 0 ]; then
    exit 1
fi
exit 0
