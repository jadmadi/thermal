#!/usr/bin/env bash
#
# benchmark.sh — End-to-End CLI Performance & Ingestion Baseline Runner
#
# Measures:
# 1. End-to-End CLI Ingestion Latency (Cold, Warm, Changed/Append)
# 2. Per-tool loader microbenchmarks & memory allocations
# 3. Sustained live monitor CPU utilization (100ms vs 1s)
# 4. Correctness validation against representative reproducible fixtures
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
echo -e "${BOLD}  THERMAL · Ingestion Performance & Latency Baseline${RESET}"
echo -e "${BOLD}${CYAN}═════════════════════════════════════════════════════════════════════${RESET}\n"

# 1. System & Hardware Environment
echo -e "${BOLD}1. Hardware & System Environment:${RESET}"
CPU_MODEL=$(grep -m1 "model name" /proc/cpuinfo 2>/dev/null | cut -d: -f2 | xargs || echo "Unknown CPU")
CPU_CORES=$(grep -c "^processor" /proc/cpuinfo 2>/dev/null || echo "1")
TOTAL_RAM=$(free -h 2>/dev/null | awk '/^Mem:/ {print $2}' || echo "Unknown")
KERNEL_VER=$(uname -sr)
GO_VER=$(go version 2>/dev/null || echo "unknown")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "   • CPU:      ${CPU_MODEL} (${CPU_CORES} cores)"
echo -e "   • Memory:   ${TOTAL_RAM}"
echo -e "   • OS:       ${KERNEL_VER}"
echo -e "   • Go:       ${GO_VER}"
echo -e "   • Git:      ${COMMIT}\n"

# 2. Build Fresh Thermal Binary
echo -e "${BOLD}2. Building Thermal binary...${RESET}"
(cd "${ROOT_DIR}" && go build -o "${BIN_PATH}" ./cmd/thermal)
echo -e "   ${GREEN}✔${RESET} Binary built at: ${BIN_PATH}\n"

# 3. Provision Representative Workload Fixtures
echo -e "${BOLD}3. Provisioning Representative Workload Fixtures...${RESET}"
BENCH_HOME="$(mktemp -d -t thermal-bench-home-XXXXXX)"
cleanup() {
    chmod -R u+w "${BENCH_HOME}" 2>/dev/null || true
    rm -rf "${BENCH_HOME}"
}
trap cleanup EXIT

ORIG_GOPATH="$(go env GOPATH 2>/dev/null || echo "${HOME}/go")"
ORIG_GOCACHE="$(go env GOCACHE 2>/dev/null || echo "${HOME}/.cache/go-build")"

MOCK_REPO="${BENCH_HOME}/projects/thermal-repo"
mkdir -p "${MOCK_REPO}/.git"
mkdir -p "${BENCH_HOME}/.codewhale/sessions"
mkdir -p "${BENCH_HOME}/.claude/projects/thermal"
mkdir -p "${BENCH_HOME}/.codex/sessions"
mkdir -p "${BENCH_HOME}/.local/share/devin/cli"
mkdir -p "${BENCH_HOME}/.local/share/opencode"
mkdir -p "${BENCH_HOME}/.local/share/mimocode"
mkdir -p "${BENCH_HOME}/.cache/thermal"

TODAY_ISO="$(date -u +"%Y-%m-%dT12:00:00Z")"
BASE_TS="$(date -u +"%s")"

# Fixture 1: Cached Pricing Catalog
cat <<EOF > "${BENCH_HOME}/.cache/thermal/pricing.json"
{
  "version": 1,
  "fetchedAt": "${TODAY_ISO}",
  "source": "https://models.dev/api.json",
  "models": {
    "claude-3-5-sonnet": { "input": 3.0, "output": 15.0, "cacheRead": 0.3, "cacheWrite": 3.75 },
    "gpt-4o": { "input": 2.5, "output": 10.0, "cacheRead": 1.25 }
  }
}
EOF

# Fixture 2: Devin SQLite Database (25 sessions, 500 message nodes)
DEVIN_DB="${BENCH_HOME}/.local/share/devin/cli/sessions.db"
sqlite3 "${DEVIN_DB}" <<EOF
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    created_at INTEGER,
    last_activity_at INTEGER,
    hidden INTEGER,
    working_directory TEXT,
    model TEXT
);
CREATE TABLE message_nodes (
    row_id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    created_at INTEGER,
    chat_message TEXT
);
CREATE TABLE prompt_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    prompt TEXT,
    created_at INTEGER,
    updated_at INTEGER
);
EOF

for i in $(seq 1 25); do
    S_TS=$((BASE_TS - (25 - i) * 3600))
    sqlite3 "${DEVIN_DB}" "INSERT INTO sessions VALUES ('devin-sess-${i}', ${S_TS}, ${S_TS}, 0, '${MOCK_REPO}', 'claude-3-5-sonnet');"
    for m in $(seq 1 20); do
        M_TS=$((S_TS + m * 60))
        MSG='{"role":"assistant","metadata":{"metrics":{"input_tokens":500,"output_tokens":100,"cache_creation_tokens":50,"cache_read_tokens":200}}}'
        sqlite3 "${DEVIN_DB}" "INSERT INTO message_nodes (session_id, created_at, chat_message) VALUES ('devin-sess-${i}', ${M_TS}, '${MSG}');"
    done
done

# Fixture 3: Codex SQLite Database & Rollout Logs (20 threads, 5 rollouts each)
CODEX_DB="${BENCH_HOME}/.codex/state_5.sqlite"
sqlite3 "${CODEX_DB}" <<EOF
CREATE TABLE threads (
    id TEXT PRIMARY KEY,
    tokens_used INTEGER,
    model TEXT,
    source TEXT,
    reasoning_effort TEXT,
    agent_role TEXT,
    created_at INTEGER,
    updated_at INTEGER,
    rollout_path TEXT,
    archived INTEGER,
    cwd TEXT
);
EOF

for i in $(seq 1 20); do
    T_TS=$((BASE_TS - (20 - i) * 3600))
    R_FILE="${BENCH_HOME}/.codex/sessions/thread_${i}.jsonl"
    for r in $(seq 1 5); do
        echo '{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":500,"cached_input_tokens":100,"output_tokens":200,"reasoning_output_tokens":50,"total_tokens":700}}}}' >> "${R_FILE}"
    done
    sqlite3 "${CODEX_DB}" "INSERT INTO threads VALUES ('codex-thread-${i}', 3500, 'gpt-4o', 'cli', 'medium', 'developer', ${T_TS}, ${T_TS}, '${R_FILE}', 0, '${MOCK_REPO}');"
done

# Fixture 4: OpenCode SQLite Database (50 sessions)
OPENCODE_DB="${BENCH_HOME}/.local/share/opencode/opencode.db"
sqlite3 "${OPENCODE_DB}" <<EOF
CREATE TABLE session_v2 (
    id TEXT PRIMARY KEY,
    tokens_input INTEGER,
    tokens_output INTEGER,
    tokens_reasoning INTEGER,
    tokens_cache_read INTEGER,
    tokens_cache_write INTEGER,
    cost REAL,
    summary_additions INTEGER,
    summary_deletions INTEGER,
    summary_files INTEGER,
    agent TEXT,
    time_created INTEGER,
    time_updated INTEGER,
    model TEXT
);
CREATE TABLE session (
    id TEXT,
    time_created INTEGER,
    time_updated INTEGER,
    tokens_input INTEGER,
    tokens_output INTEGER,
    tokens_reasoning INTEGER,
    tokens_cache_read INTEGER,
    tokens_cache_write INTEGER,
    cost REAL,
    summary_additions INTEGER,
    summary_deletions INTEGER,
    summary_files INTEGER,
    agent TEXT
);
EOF

for i in $(seq 1 50); do
    O_TS=$(( (BASE_TS - (50 - i) * 1800) * 1000 ))
    sqlite3 "${OPENCODE_DB}" "INSERT INTO session_v2 VALUES ('oc-sess-${i}', 1000, 200, 50, 400, 100, 0.05, 10, 2, 3, 'code', ${O_TS}, ${O_TS}, '{\"id\":\"claude-3-5-sonnet\"}');"
done

# Fixture 5: Claude Directory Sessions (30 files)
for i in $(seq 1 30); do
    C_TS=$(date -u -d "@$((BASE_TS - (30 - i) * 1800))" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "${TODAY_ISO}")
    echo "{\"type\":\"assistant\",\"timestamp\":\"${C_TS}\",\"cwd\":\"${MOCK_REPO}\",\"message\":{\"id\":\"claude-msg-${i}\",\"model\":\"claude-3-5-sonnet\",\"usage\":{\"input_tokens\":1000,\"output_tokens\":500,\"cache_creation_input_tokens\":200,\"cache_read_input_tokens\":8000}}}" > "${BENCH_HOME}/.claude/projects/thermal/sess_${i}.jsonl"
done

# Fixture 6: CodeWhale Sessions (20 files)
for i in $(seq 1 20); do
    W_TS=$(date -u -d "@$((BASE_TS - (20 - i) * 1800))" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "${TODAY_ISO}")
    cat <<EOF > "${BENCH_HOME}/.codewhale/sessions/session_${i}.json"
{
  "session_id": "whale-sess-${i}",
  "metadata": {
    "created_at": "${W_TS}",
    "updated_at": "${W_TS}",
    "message_count": 10,
    "total_tokens": 50000,
    "cost": { "session_cost_usd": 0.15 },
    "model": "claude-3-5-sonnet",
    "mode": "chat",
    "workspace": "${MOCK_REPO}"
  }
}
EOF
done

# Set isolated environment variables
export HOME="${BENCH_HOME}"
export GROK_HOME="${BENCH_HOME}/.grok"
export DSH_HOME="${BENCH_HOME}/.dsh"
export HERMES_HOME="${BENCH_HOME}/.hermes"
export CODEX_HOME="${BENCH_HOME}/.codex"
export OPENCODE_HOME="${BENCH_HOME}/.opencode"
export NO_COLOR="1"
export CLICOLOR_FORCE="0"
export GOPATH="${ORIG_GOPATH}"
export GOCACHE="${ORIG_GOCACHE}"

echo -e "   ${GREEN}✔${RESET} Workload fixture generated with:"
echo -e "     • Devin:     25 sessions, 500 message nodes (SQLite)"
echo -e "     • Codex:     20 threads, 100 rollout log entries (SQLite + JSONL)"
echo -e "     • OpenCode:  50 sessions (SQLite session_v2)"
echo -e "     • Claude:    30 sessions (Multi-file JSONL)"
echo -e "     • CodeWhale: 20 sessions (Multi-file JSON)\n"

# 4. Correctness Validation
echo -e "${BOLD}4. Validating Correctness on Fixtures...${RESET}"
VALID_OUT=$("${BIN_PATH}" --json --offline)
TOTAL_TOKENS=$(echo "${VALID_OUT}" | jq '[.results[].Summary.lifetimeTokens] | add')
TOOL_COUNT=$(echo "${VALID_OUT}" | jq '.results | length')
echo -e "   • Ingested Tools: ${TOOL_COUNT}"
echo -e "   • Total Tokens:   ${TOTAL_TOKENS}"

if [[ "${TOTAL_TOKENS}" -le 0 || "${TOOL_COUNT}" -lt 5 ]]; then
    echo -e "   ${RED}✖ Fixture validation failed (unexpected total tokens or tool count)${RESET}"
    exit 1
fi
echo -e "   ${GREEN}✔${RESET} Fixture data verified.\n"

# Helper for timing (returns elapsed ms)
time_cmd() {
    local start
    local end
    start=$(date +%s%N)
    "$@" >/dev/null 2>&1
    end=$(date +%s%N)
    echo $(( (end - start) / 1000000 ))
}

# 5. Measure CLI Ingestion Latency (Cold vs Warm vs Changed)
echo -e "${BOLD}5. Measuring CLI Ingestion Latency...${RESET}"

# A. Cold Ingestion (Empty delta cache)
COLD_RUNS=5
COLD_TIMES=()
for i in $(seq 1 ${COLD_RUNS}); do
    rm -rf "${BENCH_HOME}/.cache/thermal/devin"
    dur=$(time_cmd "${BIN_PATH}" --json --offline)
    COLD_TIMES+=("${dur}")
done

# B. Warm Ingestion (Populated delta cache)
# Run once to ensure cache is hot
"${BIN_PATH}" --json --offline >/dev/null 2>&1

WARM_RUNS=10
WARM_TIMES=()
for i in $(seq 1 ${WARM_RUNS}); do
    dur=$(time_cmd "${BIN_PATH}" --json --offline)
    WARM_TIMES+=("${dur}")
done

# C. Changed / Append Ingestion
# Add 1 message node to Devin, 1 session to OpenCode
CHANGED_RUNS=5
CHANGED_TIMES=()
for i in $(seq 1 ${CHANGED_RUNS}); do
    NOW_TS=$(date +%s)
    sqlite3 "${DEVIN_DB}" "INSERT INTO message_nodes (session_id, created_at, chat_message) VALUES ('devin-sess-1', ${NOW_TS}, '{\"metrics\":{\"input_tokens\":50,\"output_tokens\":20}}');"
    dur=$(time_cmd "${BIN_PATH}" --json --offline)
    CHANGED_TIMES+=("${dur}")
done

calc_stats() {
    local -a arr=("$@")
    local count=${#arr[@]}
    IFS=$'\n' sorted=($(sort -n <<<"${arr[*]}"))
    unset IFS

    local min=${sorted[0]}
    local max=${sorted[$((count - 1))]}
    local p50_idx=$((count * 50 / 100))
    local p95_idx=$((count * 95 / 100))
    local p50=${sorted[${p50_idx}]}
    local p95=${sorted[${p95_idx}]}

    local sum=0
    for v in "${arr[@]}"; do
        sum=$((sum + v))
    done
    local avg=$((sum / count))

    echo "min=${min}ms, avg=${avg}ms, p50=${p50}ms, p95=${p95}ms, max=${max}ms"
}

echo -e "   • ${BOLD}Cold Ingestion:${RESET}    $(calc_stats "${COLD_TIMES[@]}")"
echo -e "   • ${BOLD}Warm Ingestion:${RESET}    $(calc_stats "${WARM_TIMES[@]}")"
echo -e "   • ${BOLD}Changed/Append:${RESET}    $(calc_stats "${CHANGED_TIMES[@]}")\n"

# 6. Sustained Live Monitoring CPU Profile
echo -e "${BOLD}6. Measuring Sustained Live Monitoring CPU Utilization...${RESET}"

TIMEFORMAT="%R %U %S"

# 100ms interval (2s duration)
echo -e "   • Running live monitor at 100ms interval (2s sustained)..."
LIVE_100MS_TMP=$(mktemp)
{ time timeout 2s "${BIN_PATH}" live --json --stream --interval 100ms --offline >/dev/null; } 2>"${LIVE_100MS_TMP}" || true
read -r REAL_100 U_100 S_100 < "${LIVE_100MS_TMP}"
rm -f "${LIVE_100MS_TMP}"

CPU_100_SEC=$(awk "BEGIN {print ${U_100} + ${S_100}}")
CPU_100_PCT=$(awk "BEGIN {if (${REAL_100} > 0) printf \"%.1f%%\", ((${U_100} + ${S_100}) / ${REAL_100}) * 100; else print \"0%\"}")

echo -e "     → Wall time: ${REAL_100}s, CPU time: ${CPU_100_SEC}s (user: ${U_100}s, sys: ${S_100}s), CPU: ${CPU_100_PCT}"

# 1s interval (2s duration)
echo -e "   • Running live monitor at 1s interval (2s sustained)..."
LIVE_1S_TMP=$(mktemp)
{ time timeout 2s "${BIN_PATH}" live --json --stream --interval 1s --offline >/dev/null; } 2>"${LIVE_1S_TMP}" || true
read -r REAL_1S U_1S S_1S < "${LIVE_1S_TMP}"
rm -f "${LIVE_1S_TMP}"

CPU_1S_SEC=$(awk "BEGIN {print ${U_1S} + ${S_1S}}")
CPU_1S_PCT=$(awk "BEGIN {if (${REAL_1S} > 0) printf \"%.1f%%\", ((${U_1S} + ${S_1S}) / ${REAL_1S}) * 100; else print \"0%\"}")

echo -e "     → Wall time: ${REAL_1S}s, CPU time: ${CPU_1S_SEC}s (user: ${U_1S}s, sys: ${S_1S}s), CPU: ${CPU_1S_PCT}\n"

# 7. Go Ingestion Microbenchmarks
echo -e "${BOLD}7. Running Go Ingestion Microbenchmarks (internal/loaders)...${RESET}"
(cd "${ROOT_DIR}" && go test ./internal/loaders -run '^$' -bench 'Benchmark' -benchmem -benchtime 500ms)

echo -e "\n${BOLD}${CYAN}─────────────────────────────────────────────────────────────────────${RESET}"
echo -e "${BOLD}${GREEN}✔ Baseline Measurements Complete!${RESET}"
echo -e "${BOLD}${CYAN}─────────────────────────────────────────────────────────────────────${RESET}\n"
