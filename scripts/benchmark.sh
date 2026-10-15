#!/usr/bin/env bash
#
# benchmark.sh — End-to-End CLI Performance & Ingestion Baseline Runner
#
# Provisions synthetic fixtures under a temporary directory, validates exact
# ingestion totals and delta behavior, then measures end-to-end CLI latency and
# sustained live-monitor behavior. The measurement logic lives in
# scripts/benchmark (Go, monotonic timing); this shell stays the public
# entrypoint. Correctness failures exit nonzero before any performance number
# is treated as evidence.
#
# Usage: ./scripts/benchmark.sh [--cold N] [--warm N] [--large-warm N]
#                              [--stream-seconds N] [--stream-warmup N]
#                              [--skip-stream] [--out DIR]
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

WORK_DIR="$(mktemp -d -t thermal-bench-XXXXXX)"
cleanup() {
    rm -rf "${WORK_DIR}"
}
trap cleanup EXIT

(cd "${ROOT_DIR}" && go build -o "${WORK_DIR}/thermal-benchmark" ./scripts/benchmark)
"${WORK_DIR}/thermal-benchmark" --repo "${ROOT_DIR}" "$@"
