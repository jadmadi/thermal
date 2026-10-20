#!/usr/bin/env bash
#
# output_gallery.sh — regenerate docs/OUTPUT-GALLERY.md from one hermetic
# fixture. The generated document is a consistency artifact, not a product
# surface; the gallery gate compares its normalized values.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${TMPDIR:-/tmp}/thermal-gallery-bin-$$"
OUT="${1:-${ROOT_DIR}/docs/OUTPUT-GALLERY.md}"
TMP="$(mktemp -d -t thermal-gallery-XXXXXX)"
FIXTURE_HOME="$( "${ROOT_DIR}/scripts/output_gallery_fixture.sh" )"
WORK="${FIXTURE_HOME}/work"

cleanup() {
  chmod -R u+w "${FIXTURE_HOME}" 2>/dev/null || true
  rm -rf "${FIXTURE_HOME}" "${TMP}" "${BIN}"
}
trap cleanup EXIT

(cd "${ROOT_DIR}" && go build -o "${BIN}" ./cmd/thermal)

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

capture_text() {
  local name="$1"
  shift
  (cd "${WORK}" && "${BIN}" "$@" --no-color --offline 2>&1) >"${TMP}/${name}.txt"
}

capture_json() {
  local name="$1"
  shift
  (cd "${WORK}" && "${BIN}" "$@" --json --offline 2>"${TMP}/${name}.err") >"${TMP}/${name}.json"
}

capture_exit() {
  local name="$1"
  shift
  set +e
  (cd "${WORK}" && "${BIN}" "$@" --offline 2>&1) >"${TMP}/${name}.txt"
  echo "$?" >"${TMP}/${name}.code"
  set -e
}

capture_width() {
  local name="$1" cols="$2"
  shift 2
  local cmd=""
  local arg
  for arg in "$@"; do
    cmd+=" $(printf '%q' "${arg}")"
  done
  script -qec "stty cols ${cols}; cd $(printf '%q' "${WORK}"); export HOME=$(printf '%q' "${FIXTURE_HOME}") GROK_HOME=$(printf '%q' "${FIXTURE_HOME}/.grok") DSH_HOME=$(printf '%q' "${FIXTURE_HOME}/.dsh") HERMES_HOME=$(printf '%q' "${FIXTURE_HOME}/.hermes") CODEX_HOME=$(printf '%q' "${FIXTURE_HOME}/.codex") OPENCODE_HOME=$(printf '%q' "${FIXTURE_HOME}/.opencode") ZED_HOME=$(printf '%q' "${FIXTURE_HOME}/.local/share/zed") AGY_HOME=$(printf '%q' "${FIXTURE_HOME}/.gemini/antigravity") NO_COLOR=1 CLICOLOR_FORCE=0; $(printf '%q' "${BIN}")${cmd} --no-color --offline 2>&1" /dev/null |
    tr -d '\r' >"${TMP}/${name}.txt"
}

FIXTURE_HASH="$(cd "${FIXTURE_HOME}" && find . -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum | cut -d' ' -f1)"
CAPTURED_AT="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
VERSION_LINE="$("${BIN}" version | head -1)"

# Text surfaces (fixed fixture; non-TTY width 100).
capture_text help --no-color --help
capture_text leaderboard
capture_text single-tool-codewhale codewhale
capture_text single-tool-agy agy
capture_text daily daily
capture_text weekly weekly
capture_text monthly monthly
capture_text projects projects
capture_text models models
capture_text mix mix
capture_text stats-token stats
capture_text stats-cost stats --metric cost
capture_text trend-token trend
capture_text trend-cost trend --metric cost
capture_text projects-chart projects --top 2 --order asc --chart
capture_text models-chart models --top 2 --order asc --chart
capture_text share share
capture_text statusline statusline
capture_text audit audit
capture_text changelog changelog --top 5
capture_text license license
capture_text live-nontty live
capture_text version version

# Machine-readable surfaces.
capture_json leaderboard
capture_json daily daily
capture_json weekly weekly
capture_json monthly monthly
capture_json projects projects
capture_json models models
capture_json mix mix
capture_json mix-cost mix --metric cost
capture_json stats-token stats
capture_json stats-cost stats --metric cost
capture_json trend-token trend
capture_json trend-cost trend --metric cost
capture_json share share
capture_json audit audit
capture_json changelog changelog --top 5
capture_json live live
capture_json version version

# Validation cases.
capture_exit validation-report-flag --since 2026-01-01
capture_exit validation-mix-chart mix --chart
capture_exit validation-daily-top daily --top 2
capture_exit validation-chart-daily daily --chart
capture_width width80-help 80 --help
capture_width width120-help 120 --help
capture_width width80-share 80 share
capture_width width120-share 120 share
capture_width width80-stats 80 stats
capture_width width120-stats 120 stats

jqr() { jq -r "$1" "${TMP}/$2"; }

LB_TOKENS="$(jqr '.totals.totalTokens' leaderboard.json)"
LB_RECORDED="$(jqr '.totals.storedCost' leaderboard.json)"
LB_ESTIMATED="$(jqr '.totals.estimatedCost' leaderboard.json)"
LB_COST="$(jqr '.totals.cost' leaderboard.json)"
LB_ACTIVE="$(jqr '.totals.activeDays' leaderboard.json)"
DAILY_TOKENS="$(jqr '.totals.totalTokens' daily.json)"
DAILY_RECORDED="$(jqr '.totals.storedCost' daily.json)"
DAILY_ESTIMATED="$(jqr '.totals.estimatedCost' daily.json)"
DAILY_COST="$(jqr '.totals.cost' daily.json)"
DAILY_WINDOW="$(jqr '.window' daily.json)"
DAILY_ACTIVE="$(jqr '.data | length' daily.json)"
WEEKLY_TOKENS="$(jqr '.totals.totalTokens' weekly.json)"
WEEKLY_RECORDED="$(jqr '.totals.storedCost' weekly.json)"
WEEKLY_ESTIMATED="$(jqr '.totals.estimatedCost' weekly.json)"
WEEKLY_COST="$(jqr '.totals.cost' weekly.json)"
WEEKLY_WINDOW="$(jqr '.window' weekly.json)"
MONTHLY_TOKENS="$(jqr '.totals.totalTokens' monthly.json)"
MONTHLY_RECORDED="$(jqr '.totals.storedCost' monthly.json)"
MONTHLY_ESTIMATED="$(jqr '.totals.estimatedCost' monthly.json)"
MONTHLY_COST="$(jqr '.totals.cost' monthly.json)"
MONTHLY_WINDOW="$(jqr '.window' monthly.json)"
PROJ_TOKENS="$(jqr '.totals.totalTokens' projects.json)"
PROJ_RECORDED="$(jqr '.totals.storedCost' projects.json)"
PROJ_ESTIMATED="$(jqr '.totals.estimatedCost' projects.json)"
PROJ_COST="$(jqr '.totals.cost' projects.json)"
PROJ_ACTIVE="$(jqr '.totals.activeDays' projects.json)"
MODEL_TOKENS="$(jqr '.totals.totalTokens' models.json)"
MODEL_UNATTR="$(jqr '.totals.unattributedTokens // 0' models.json)"
MODEL_ESTIMATED="$(jqr '.totals.estimatedCost' models.json)"
MODEL_DAYS="$(jqr '.totals.activeDays' models.json)"
MIX_COST="$(jqr '.total' mix-cost.json)"
MIX_BASIS="$(jqr '.costBasis' mix-cost.json)"
STATS_TOKENS="$(jqr '.total' stats-token.json)"
STATS_ACTIVE="$(jqr '.activeDays' stats-token.json)"
STATS_COST="$(jqr '.total' stats-cost.json)"
STATS_COST_ACTIVE="$(jqr '.activeDays' stats-cost.json)"
STATS_COSTED="$(jqr '.costedDays // 0' stats-cost.json)"
TREND_TOKEN_MEAN="$(jqr '.mean' trend-token.json)"
TREND_COST_MEAN="$(jqr '.mean' trend-cost.json)"
SHARE_TOKENS="$(jqr '.snapshot.totalTokens' share.json)"
SHARE_ACTIVE="$(jqr '.snapshot.activeDays' share.json)"
SHARE_RECORDED="$(jqr '.snapshot.recordedCost' share.json)"
SHARE_ESTIMATED="$(jqr '.snapshot.estimatedCost' share.json)"
SHARE_COST="$(jqr '.snapshot.totalCost' share.json)"
SHARE_SCOPE="$(jqr '.snapshot.scope' share.json)"
SHARE_WINDOW="$(jqr '.snapshot.window' share.json)"
SHARE_BASIS="$(jqr '.snapshot.coverage' share.json)"
LIVE_TYPE="$(jqr '.type' live.json)"
LIVE_TS="$(jqr '.timestamp' live.json)"

cat >"${OUT}" <<EOF
# Thermal Output Gallery

> Generated from a fixed fixture by \`scripts/output_gallery.sh\`. Capture
> \`${CAPTURED_AT}\` · fixture \`${FIXTURE_HASH}\` · \`${VERSION_LINE}\`.
> Fixed data date: 2026-01-15. Text is \`NO_COLOR=1\`; non-TTY default width is
> 100. Width sections use a PTY at 80/120. The live TUI block is synthetic and
> labeled; the other text blocks are captured from the same fixture run.

## Contents

- [Consistency matrix](#consistency-matrix)
- [Text surfaces](#text-surfaces)
  - [Global help](#global-help)
  - [Leaderboard](#leaderboard)
  - [Daily](#daily) · [Weekly](#weekly) · [Monthly](#monthly)
  - [Projects](#projects) · [Models](#models)
  - [Mix](#mix) · [Stats](#stats) · [Trend](#trend)
  - [Share](#share) · [Statusline](#statusline) · [Audit](#audit)
  - [Changelog](#changelog) · [License](#license) · [Live](#live)
- [Width and color contracts](#width-and-color-contracts)
- [Row-level JSON schemas](#row-level-json-schemas)
- [Validation cases](#validation-cases)
- [Appendices](#appendices)

## Consistency matrix

| Surface | Tokens | Recorded | Estimated | Total | Active/costed days | Window/scope | Basis |
|---|---:|---:|---:|---:|---|---|---|
| leaderboard | ${LB_TOKENS} | ${LB_RECORDED} | ${LB_ESTIMATED} | ${LB_COST} | ${LB_ACTIVE}/— | all tools | total |
| daily | ${DAILY_TOKENS} | ${DAILY_RECORDED} | ${DAILY_ESTIMATED} | ${DAILY_COST} | ${DAILY_ACTIVE}/— | ${DAILY_WINDOW} | total |
| weekly | ${WEEKLY_TOKENS} | ${WEEKLY_RECORDED} | ${WEEKLY_ESTIMATED} | ${WEEKLY_COST} | — | ${WEEKLY_WINDOW} | total |
| monthly | ${MONTHLY_TOKENS} | ${MONTHLY_RECORDED} | ${MONTHLY_ESTIMATED} | ${MONTHLY_COST} | — | ${MONTHLY_WINDOW} | total |
| projects | ${PROJ_TOKENS} | ${PROJ_RECORDED} | ${PROJ_ESTIMATED} | ${PROJ_COST} | ${PROJ_ACTIVE} project-days | all tools/projects | total |
| models | ${MODEL_TOKENS} classified + ${MODEL_UNATTR} unclassified | 0 | ${MODEL_ESTIMATED} | ${MODEL_ESTIMATED} | ${MODEL_DAYS} model-days | all models | estimated |
| mix (tool cost) | — | — | — | ${MIX_COST} | — | all tools | ${MIX_BASIS} |
| stats (tokens) | ${STATS_TOKENS} | — | — | — | ${STATS_ACTIVE}/— | all time | n/a (tokens) |
| stats (cost) | — | — | — | ${STATS_COST} | ${STATS_COST_ACTIVE}/${STATS_COSTED} | all time | total |
| trend (tokens) | mean ${TREND_TOKEN_MEAN} | — | — | — | — | calendar | n/a (tokens) |
| trend (cost) | mean ${TREND_COST_MEAN} | — | — | — | — | calendar | total |
| share | ${SHARE_TOKENS} | ${SHARE_RECORDED} | ${SHARE_ESTIMATED} | ${SHARE_COST} | ${SHARE_ACTIVE}/— | ${SHARE_SCOPE} · ${SHARE_WINDOW} | ${SHARE_BASIS} |
| live | — | — | — | — | — | today snapshot (${LIVE_TYPE}, ${LIVE_TS}) | recorded+estimated |

Allowed equalities are checked by \`scripts/output_gallery_gate.sh\`: leaderboard
estimated equals daily estimated, projects total equals daily total, models
classified + unclassified equals daily total, tool-mix cost equals daily total,
and share token/cost fields use one window.

## Text surfaces

### Global help

\`\`\`text
$(cat "${TMP}/help.txt")
\`\`\`

### Leaderboard

\`\`\`text
$(cat "${TMP}/leaderboard.txt")
\`\`\`

Single tool: [codewhale](#single-tool-codewhale) · [Agy](#single-tool-agy)

### Single tool: codewhale

\`\`\`text
$(cat "${TMP}/single-tool-codewhale.txt")
\`\`\`

### Single tool: Agy

\`\`\`text
$(cat "${TMP}/single-tool-agy.txt")
\`\`\`

### Daily

\`\`\`text
$(cat "${TMP}/daily.txt")
\`\`\`

### Weekly

\`\`\`text
$(cat "${TMP}/weekly.txt")
\`\`\`

### Monthly

\`\`\`text
$(cat "${TMP}/monthly.txt")
\`\`\`

### Projects

\`\`\`text
$(cat "${TMP}/projects.txt")
\`\`\`

Chart variant (\`projects --top 2 --order asc --chart\`):

\`\`\`text
$(cat "${TMP}/projects-chart.txt")
\`\`\`

### Models

\`\`\`text
$(cat "${TMP}/models.txt")
\`\`\`

Chart variant (\`models --top 2 --order asc --chart\`):

\`\`\`text
$(cat "${TMP}/models-chart.txt")
\`\`\`

### Mix

\`\`\`text
$(cat "${TMP}/mix.txt")
\`\`\`

### Stats

Token stats:

\`\`\`text
$(cat "${TMP}/stats-token.txt")
\`\`\`

Cost stats:

\`\`\`text
$(cat "${TMP}/stats-cost.txt")
\`\`\`

### Trend

Token trend:

\`\`\`text
$(cat "${TMP}/trend-token.txt")
\`\`\`

Cost trend:

\`\`\`text
$(cat "${TMP}/trend-cost.txt")
\`\`\`

### Share

\`\`\`text
$(cat "${TMP}/share.txt")
\`\`\`

The full URL is in [the share JSON appendix](#full-share-json).

### Statusline

\`\`\`text
$(cat "${TMP}/statusline.txt")
\`\`\`

### Audit

\`\`\`text
$(cat "${TMP}/audit.txt")
\`\`\`

### Changelog

\`\`\`text
$(cat "${TMP}/changelog.txt")
\`\`\`

### License

\`\`\`text
$(cat "${TMP}/license.txt")
\`\`\`

### Live

Non-TTY pointer:

\`\`\`text
$(cat "${TMP}/live-nontty.txt")
\`\`\`

Representative synthetic TUI frame (80x24; units and labels are the contract,
values are illustrative):

\`\`\`text
╭─ 🔥 THERMAL LIVE ─────────────────────────────── ● LIVE (1.0s) · 12:00:00 ─╮
│ BURN VELOCITY           TODAY'S USAGE           LIVE SESSION              │
│ Rate:     3.4M tok/m    Volume:   1.7M tok      Since launch: 1.6M tok   │
│ Speed:    56.7K tok/s   Cost:     ~\$1.20        Spend:    ~\$1.10          │
│                         Burned: +60 step (session)                     │
│ Burn/hr now: \$0.00/hr   Cache hit: 96.1%        State:    IDLE 💤         │
│   Steps: 23 step/m · 0.383 step/s                                         │
│   Top model (est. cost): deepseek-v4.1-flash ~\$3.95                       │
│   Today by project: thermal ~\$4.43 · scripts ~\$0.46                       │
│   Flame: [····················]   0% of 9.0M tok/m  IDLE 💤              │
│   Spark: [···▂▃▂···············]  Peak: 1.0K tok/s (1s bucket)            │
│   Last event: 0s ago · * >60s · ! >3x                                     │
│   CMD×3  1.7M TOK  60 STEP  MODEL · PROJECT  ~\$0.30                       │
╰──────────────────────────────────────────────────────────────────────────╯
 • Controls: [Space] Pause · [c] Reset · [s] Sort: time · [?] Help · [q] Quit
\`\`\`

At 120x40 the same frame keeps the full hybrid rate/speed units and controls;
the 80x24 frame uses the compact \`Steps:\` line when a KPI cell cannot fit the
secondary value. No unit is truncated.

## Width and color contracts

All captures below use \`NO_COLOR=1\`; no ANSI escapes are present. The body is
compact; the full fixture tables are small enough to keep inline.

The fixed report grid is 93 cells wide; at 80 it is intentionally not shrunk
because the columns are numeric values, not prose. Responsive surfaces (help,
stats, share, statusline) are captured at both widths.

### 80 columns: help

\`\`\`text
$(cat "${TMP}/width80-help.txt")
\`\`\`

### 120 columns: help

\`\`\`text
$(cat "${TMP}/width120-help.txt")
\`\`\`

### 80 columns: stats

\`\`\`text
$(cat "${TMP}/width80-stats.txt")
\`\`\`

### 120 columns: stats

\`\`\`text
$(cat "${TMP}/width120-stats.txt")
\`\`\`

### 80 columns: share

\`\`\`text
$(cat "${TMP}/width80-share.txt")
\`\`\`

### 120 columns: share

\`\`\`text
$(cat "${TMP}/width120-share.txt")
\`\`\`

## Row-level JSON schemas

Canonical lower-camelCase keys; no uppercase aliases. All floats are serialized
at four decimal places by \`cmd/thermal/json.go\`.

### Leaderboard result

\`\`\`text
$(jq -c '.results[0] | keys' "${TMP}/leaderboard.json")
$(jq -c '.results[0].summary | keys' "${TMP}/leaderboard.json")
\`\`\`

### Daily row

\`\`\`text
$(jq -c '.data[0] | keys' "${TMP}/daily.json")
\`\`\`

### Project row

\`\`\`text
$(jq -c '.data[0] | keys' "${TMP}/projects.json")
\`\`\`

### Model row

\`\`\`text
$(jq -c '.data[0] | keys' "${TMP}/models.json")
\`\`\`

### Live event / rollup

\`\`\`text
$(jq -c '.recentEvents[0] // {} | keys' "${TMP}/live.json")
$(jq -c '.todayModels[0] // {} | keys' "${TMP}/live.json")
\`\`\`

### Share snapshot

\`\`\`text
$(jq -c '.snapshot | keys' "${TMP}/share.json")
\`\`\`

### Envelope types and timestamps

\`\`\`text
$(for f in leaderboard daily weekly monthly projects models mix stats-token stats-cost trend-token trend-cost share audit changelog live version; do
  printf '%-16s %-14s %s\n' "${f}" "$(jq -r '.type // "—"' "${TMP}/${f}.json")" "$(jq -r '.generatedAt // .timestamp // "—"' "${TMP}/${f}.json")"
done)
\`\`\`

## Validation cases

Actual command and exit status; stderr/stdout are combined for readability.

### Report flag without report

\`\`\`text
$ thermal --since 2026-01-01
exit $(cat "${TMP}/validation-report-flag.code")
$(cat "${TMP}/validation-report-flag.txt")
\`\`\`

### Analytics --chart is rejected

\`\`\`text
$ thermal mix --chart
exit $(cat "${TMP}/validation-mix-chart.code")
$(cat "${TMP}/validation-mix-chart.txt")
\`\`\`

### --top on daily is rejected

\`\`\`text
$ thermal daily --top 2
exit $(cat "${TMP}/validation-daily-top.code")
$(cat "${TMP}/validation-daily-top.txt")
\`\`\`

### --chart on daily succeeds

\`\`\`text
$ thermal daily --chart
exit $(cat "${TMP}/validation-chart-daily.code")
$(cat "${TMP}/validation-chart-daily.txt")
\`\`\`

## Appendices

### Full share JSON

\`\`\`json
$(cat "${TMP}/share.json")
\`\`\`

### Fixture

- Home: fixed temp fixture (removed after generation)
- Date: 2026-01-15
- Hash: \`${FIXTURE_HASH}\`
- Reports: codewhale 750,000 tokens/\$1.75 recorded; Claude 9,700 tokens;
  Agy 12 steps; two projects; two models.

### Normalization

The gallery gate compares normalized totals and schema keys, not capture
timestamps or randomized fixture paths. It fails on retired commands, mixed
totals, aliases, >4-decimal floats, missing basis/unclassified fields, unit
truncation, overlapping heatmap month labels, chart scale/order drift, and empty validation cases.
EOF

echo "wrote ${OUT}"
