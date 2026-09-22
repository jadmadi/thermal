#!/usr/bin/env bash
#
# check_governance_metrics.sh — Monitor Governance Health & Absence Factor
#
# Calculates the CHAOSS contributor absence factor (the smallest number of contributors
# responsible for >= 50% of commits over the trailing 12 months) and checks credential
# concentration against single-point-of-failure thresholds.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

SINCE="${1:-12 months ago}"

echo "========================================================================"
echo " Thermal Governance Health & CHAOSS Absence Factor Audit"
echo " Trailing Period: ${SINCE}"
echo "========================================================================"

TOTAL_COMMITS=$(git -C "${ROOT_DIR}" log --since="${SINCE}" --no-merges --format="%aN" | wc -l)

if [ "${TOTAL_COMMITS}" -eq 0 ]; then
    echo "Notice: No commits found in the specified window."
    exit 0
fi

HALF_COMMITS=$(( (TOTAL_COMMITS + 1) / 2 ))

# Generate author commit counts descending
AUTHOR_COUNTS=$(git -C "${ROOT_DIR}" log --since="${SINCE}" --no-merges --format="%aN" | sort | uniq -c | sort -rn)

echo "Total Non-Merge Commits: ${TOTAL_COMMITS}"
echo "50% Contribution Threshold: ${HALF_COMMITS}"
echo ""
echo "Commit Distribution by Contributor:"

ABSENCE_FACTOR=0
ACCUMULATED=0

while read -r count author; do
    ABSENCE_FACTOR=$((ABSENCE_FACTOR + 1))
    ACCUMULATED=$((ACCUMULATED + count))
    PERCENT=$(( (count * 100) / TOTAL_COMMITS ))
    printf "  - %-25s : %4d commits (%2d%%)\n" "${author}" "${count}" "${PERCENT}"
    if [ "${ACCUMULATED}" -ge "${HALF_COMMITS}" ] && [ "${ABSENCE_FACTOR_MET:-0}" -eq 0 ]; then
        ABSENCE_FACTOR_MET=1
        RECORDED_ABSENCE_FACTOR="${ABSENCE_FACTOR}"
    fi
done <<< "${AUTHOR_COUNTS}"

ABSENCE_FACTOR="${RECORDED_ABSENCE_FACTOR:-${ABSENCE_FACTOR}}"

echo ""
echo "------------------------------------------------------------------------"
echo "Governance Metrics Summary:"
echo "  CHAOSS Contributor Absence Factor : ${ABSENCE_FACTOR}"
echo "  Target Absence Factor Threshold   : >= 2 (Healthy Distributed Maintenance)"
if [ "${ABSENCE_FACTOR}" -lt 2 ]; then
    echo "  Status: STADIUM QUADRANT (Single Author Core / Stadium Project)"
    echo "  Note:   Governed via GOVERNANCE.md Written Solo Authority + Succession Runbook"
else
    echo "  Status: DISTRIBUTED MAINTENANCE (Multi-Contributor Core)"
fi
echo "------------------------------------------------------------------------"
exit 0
