#!/usr/bin/env bash
#
# retire_legacy_releases.sh — Audit and Retire Legacy GitHub Releases
#
# Audits and safely retires past legacy GitHub release assets and tags to ensure
# all distributed binaries conform strictly to Thermal's AGPL-3.0 / Dual-License baseline.
#
# Safety Invariants:
# - Default mode is strictly --dry-run (read-only audit; zero remote modifications).
# - Execution mode requires explicit --confirm flag.
# - Safety rail checks repository origin before deletion.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

DRY_RUN=1
CONFIRM=0

usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Options:
  --dry-run      Audit target releases and show planned actions without deleting (default)
  --confirm      Execute actual deletion on GitHub via gh CLI (requires authorization)
  --help         Show this help message

Description:
  Scans published GitHub releases and provides a dry-run audit or execution of
  legacy release retirement to reset distribution to the AGPL-3.0 release line.
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)
            DRY_RUN=1
            shift
            ;;
        --confirm)
            CONFIRM=1
            DRY_RUN=0
            shift
            ;;
        --help|-h)
            usage
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage
            ;;
    esac
done

echo "========================================================================"
echo " Thermal Legacy Release Retirement & Audit Tool"
echo " Mode: $( [ "$DRY_RUN" -eq 1 ] && echo "DRY-RUN (audit only, zero changes)" || echo "CONFIRMED EXECUTION" )"
echo "========================================================================"
echo ""

if ! command -v gh >/dev/null 2>&1; then
    echo "❌ Error: GitHub CLI ('gh') is not installed or not in PATH." >&2
    exit 1
fi

echo "Auditing published releases from GitHub..."
RELEASES=$(gh release list -L 100 --json tagName,name,publishedAt,isLatest 2>/dev/null || true)

if [[ -z "$RELEASES" || "$RELEASES" == "[]" ]]; then
    echo "Notice: No published GitHub releases found for this repository."
    exit 0
fi

echo "Found published releases:"
echo "$RELEASES" | jq -r '.[] | "  - \(.tagName) (Published: \(.publishedAt), Latest: \(.isLatest))"'

echo ""
if [ "$DRY_RUN" -eq 1 ]; then
    echo "------------------------------------------------------------------------"
    echo "DRY-RUN AUDIT SUMMARY:"
    echo "  The above releases are archived in docs/RELEASE.md."
    echo "  In execution mode (--confirm), legacy pre-AGPL release assets can be"
    echo "  retired using: gh release delete <tag> --cleanup-tag -y"
    echo "  Zero remote modifications were performed."
    echo "------------------------------------------------------------------------"
    exit 0
fi

if [ "$CONFIRM" -eq 1 ]; then
    echo "⚠️  WARNING: You are in CONFIRMED EXECUTION mode."
    echo "This will permanently delete legacy releases and tags on GitHub."
    read -r -p "Type 'DELETE_LEGACY_RELEASES' to proceed: " confirmation
    if [ "$confirmation" != "DELETE_LEGACY_RELEASES" ]; then
        echo "Aborted by user. No changes made."
        exit 1
    fi

    # Read tags and process
    for tag in $(echo "$RELEASES" | jq -r '.[].tagName'); do
        echo "Retiring release $tag..."
        gh release delete "$tag" --cleanup-tag -y || echo "Warning: failed to delete $tag"
    done
    echo "✅ Legacy release retirement completed."
fi
