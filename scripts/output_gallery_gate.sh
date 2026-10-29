#!/usr/bin/env bash
#
# output_gallery_gate.sh — CI-friendly consistency gate for the generated
# output gallery. Regenerates from the hermetic fixture and checks normalized
# values, schemas, units, validation exits, and retired surfaces.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d -t thermal-gallery-gate-XXXXXX)"
BIN="${TMP}/thermal"
GALLERY="${TMP}/OUTPUT-GALLERY.md"
DOC_GALLERY="${ROOT_DIR}/docs/OUTPUT-GALLERY.md"

if [[ -n "${THERMAL_GALLERY_BIN:-}" && -x "${THERMAL_GALLERY_BIN}" ]]; then
  BIN="${THERMAL_GALLERY_BIN}"
else
  (cd "${ROOT_DIR}" && go build -o "${BIN}" ./cmd/thermal)
fi

FIXTURE_HOME="$( "${ROOT_DIR}/scripts/output_gallery_fixture.sh" )"
WORK="${FIXTURE_HOME}/work"

cleanup() {
  chmod -R u+w "${FIXTURE_HOME}" 2>/dev/null || true
  rm -rf "${TMP}" "${FIXTURE_HOME}"
}
trap cleanup EXIT

export HOME="${FIXTURE_HOME}"
export GROK_HOME="${FIXTURE_HOME}/.grok"
export DSH_HOME="${FIXTURE_HOME}/.dsh"
export HERMES_HOME="${FIXTURE_HOME}/.hermes"
export CODEX_HOME="${FIXTURE_HOME}/.codex"
export OPENCODE_HOME="${FIXTURE_HOME}/.opencode"
export ZED_HOME="${FIXTURE_HOME}/.local/share/zed"
export ANTIGRAVITY_APP_DATA_DIR="antigravity"
export AGY_HOME="${FIXTURE_HOME}/.gemini/antigravity"
export NO_COLOR=1
export CLICOLOR_FORCE=0

THERMAL_GALLERY_BIN="${BIN}" "${ROOT_DIR}/scripts/output_gallery.sh" "${GALLERY}" >/dev/null

FAILED=0
check() {
  local name="$1"
  shift
  if "$@"; then
    printf '  ✔ %s\n' "${name}"
  else
    printf '  ✘ %s\n' "${name}"
    FAILED=1
  fi
}

check_contains() {
  local name="$1" needle="$2" file="${3:-${GALLERY}}"
  if grep -Fq -- "${needle}" "${file}"; then
    printf '  ✔ %s\n' "${name}"
  else
    printf '  ✘ %s (missing %q)\n' "${name}" "${needle}"
    FAILED=1
  fi
}

check_absent() {
  local name="$1" pattern="$2"
  if grep -Eq -- "${pattern}" "${GALLERY}"; then
    printf '  ✘ %s (found %q)\n' "${name}" "${pattern}"
    FAILED=1
  else
    printf '  ✔ %s\n' "${name}"
  fi
}

json() {
  local name="$1"
  shift
  (cd "${WORK}" && "${BIN}" "$@" --json --offline) >"${TMP}/${name}.json"
}

section() {
  local heading="$1"
  awk -v h="${heading}" '$0 == h {on=1; next} on && /^### / {exit} on {print}' "${GALLERY}"
}

echo "Output gallery gate"

check_absent "retired surfaces and unit fragments" 'thermal (dashboard|web|serve|replay|yield|receipt)|9-box|ONov|step…|tok…'
if LC_ALL=C grep -q $'\033\[' "${GALLERY}"; then
  printf '  ✘ gallery contains ANSI escapes\n'; FAILED=1
else
  printf '  ✔ no ANSI escapes\n'
fi

check_contains "help includes Zed" "Zed"
check_contains "help includes MiMoCode alias" "MiMoCode (mimo)"
check_contains "help includes DSH alias" "DeepSeek DSH (dsh)"
check_contains "validation no-report command" "thermal --since 2026-01-01"
check_contains "validation no-report exit 2" "exit 2"
check_contains "chart scale names largest charted row" "largest charted row"
check_contains "chart keeps table order" "repo-b"
check_contains "live synthetic has token units" "tok/m"
check_contains "live synthetic has step units" "step/m"
check_contains "live synthetic has speed units" "tok/s"
check_contains "live synthetic has step speed" "step/s"
check_contains "live synthetic has burned steps" "+60 step"
check_contains "share URL is shortened" "full URL"
check_contains "statusline scope" "d all tools ·"
check_contains "statusline today label" "tok today"
check_contains "statusline top label" "top:"

if grep -q '^### Daily$' "${GALLERY}"; then
  count=$(grep -c '^### Daily$' "${GALLERY}")
  if [[ "${count}" -eq 1 ]]; then
    printf '  ✔ one daily section (no duplicate validation capture)\n'
  else
    printf '  ✘ duplicate daily sections: %s\n' "${count}"; FAILED=1
  fi
else
  printf '  ✘ daily section missing\n'; FAILED=1
fi

for spec in "80 help" "120 help" "80 stats" "120 stats" "80 share" "120 share"; do
  cols="${spec%% *}"
  name="${spec#* }"
  max=$(section "### ${cols} columns: ${name}" | awk '{ if (length($0) > m) m = length($0) } END { print m+0 }')
  if [[ "${max}" -le "${cols}" ]]; then
    printf '  ✔ %s-column %s fits (%s)\n' "${cols}" "${name}" "${max}"
  else
    printf '  ✘ %s-column %s exceeds width (%s)\n' "${cols}" "${name}" "${max}"; FAILED=1
  fi
done

json leaderboard
json daily daily
json projects projects
json models models
json mix-cost mix --metric cost
json stats-token stats
json stats-cost stats --metric cost
json share share

check "leaderboard has no uppercase aliases" bash -c 'test "$(jq "[.results[0] | keys[] | select(test(\"^[A-Z]\"))] | length" "$1")" -eq 0' _ "${TMP}/leaderboard.json"
check "leaderboard estimated == daily estimated" bash -c 'jq -e --slurpfile d "$2" "(.totals.estimatedCost - \$d[0].totals.estimatedCost) | fabs < 0.0001" "$1" >/dev/null' _ "${TMP}/leaderboard.json" "${TMP}/daily.json"
check "projects total == daily total" bash -c 'jq -e --slurpfile d "$2" "(.totals.cost - \$d[0].totals.cost) | fabs < 0.0001" "$1" >/dev/null' _ "${TMP}/projects.json" "${TMP}/daily.json"
check "models classified+unclassified == daily tokens" bash -c 'jq -e --slurpfile d "$2" "(.totals.totalTokens + .totals.unattributedTokens) == \$d[0].totals.totalTokens" "$1" >/dev/null' _ "${TMP}/models.json" "${TMP}/daily.json"
check "models exposes unclassified" jq -e 'has("totals") and (.totals | has("unattributedTokens"))' "${TMP}/models.json"
check "stats exposes unclassified" jq -e '.composition | has("unclassified")' "${TMP}/stats-token.json"
check "stats cost has active/costed" jq -e '.activeDays > 0 and (.costedDays // 0) > 0' "${TMP}/stats-cost.json"
check "tool mix cost == daily cost" bash -c 'jq -e --slurpfile d "$2" "(.total - \$d[0].totals.cost) | fabs < 0.0001" "$1" >/dev/null' _ "${TMP}/mix-cost.json" "${TMP}/daily.json"
check "share total == daily total" bash -c 'jq -e --slurpfile d "$2" "(.snapshot.totalCost - \$d[0].totals.cost) | fabs < 0.0001" "$1" >/dev/null' _ "${TMP}/share.json" "${TMP}/daily.json"
check "share active days == daily rows" bash -c 'jq -e --slurpfile d "$2" ".snapshot.activeDays == (\$d[0].data | length)" "$1" >/dev/null' _ "${TMP}/share.json" "${TMP}/daily.json"
check "share fields are non-null" jq -e '.snapshot | [.scope,.window,.streakBasis,.activeDays,.totalTokens,.recordedCost,.estimatedCost,.totalCost,.coverage] | all(. != null)' "${TMP}/share.json"
check "share URL full in JSON" bash -c 'test "$(jq -r ".url | length" "$1")" -gt 96' _ "${TMP}/share.json"

json_precision_bad=0
for json_file in "${TMP}"/*.json; do
  if jq -r '.. | numbers' "${json_file}" | grep -Eq '\.[0-9]{5,}'; then
    json_precision_bad=1
  fi
done
if [[ "${json_precision_bad}" -eq 1 ]]; then
  printf '  ✘ JSON contains >4 decimal places\n'; FAILED=1
else
  printf '  ✔ JSON floats use <=4 decimals\n'
fi

chart=$(cd "${WORK}" && "${BIN}" projects --top 2 --order asc --chart --no-color --offline)
if grep -q 'largest charted row, 750.0K tok' <<<"${chart}"; then
  printf '  ✔ chart scales to the largest charted row\n'
else
  printf '  ✘ chart scale drifted\n'; FAILED=1
fi
if grep -q 'ONov' <<<"$(cd "${WORK}" && "${BIN}" codewhale --no-color --offline)"; then
  printf '  ✘ overlapping heatmap labels\n'; FAILED=1
else
  printf '  ✔ heatmap labels do not overlap\n'
fi

set +e
(cd "${WORK}" && "${BIN}" --since 2026-01-01 --offline >/dev/null 2>&1)
root_code=$?
(cd "${WORK}" && "${BIN}" mix --chart --offline >/dev/null 2>&1)
mix_code=$?
(cd "${WORK}" && "${BIN}" stats --top 2 --offline >/dev/null 2>&1)
top_code=$?
(cd "${WORK}" && "${BIN}" daily --chart --offline >/dev/null 2>&1)
daily_code=$?
set -e
if [[ "${root_code}" -eq 2 && "${mix_code}" -eq 2 && "${top_code}" -eq 2 && "${daily_code}" -eq 0 ]]; then
  printf '  ✔ validation exits are stable\n'
else
  printf '  ✘ validation exit drift: %s/%s/%s/%s\n' "${root_code}" "${mix_code}" "${top_code}" "${daily_code}"
  FAILED=1
fi

matrix() {
  awk '/^## Consistency matrix$/{on=1;next} on && /^## /{exit} on && /^\|/{print}' "$1" |
    sed -E 's/\(live, [^)]*\)/(live, TIMESTAMP)/'
}
if [[ -f "${DOC_GALLERY}" ]] && diff -q <(matrix "${DOC_GALLERY}") <(matrix "${GALLERY}") >/dev/null; then
  printf '  ✔ checked-in gallery matrix matches generated fixture\n'
else
  printf '  ✘ checked-in gallery matrix is stale; regenerate docs/OUTPUT-GALLERY.md\n'
  FAILED=1
fi

if [[ "${FAILED}" -eq 0 ]]; then
  printf '\n✔ OUTPUT GALLERY GATE PASSED\n'
  exit 0
fi
printf '\n✘ OUTPUT GALLERY GATE FAILED\n'
exit 1
