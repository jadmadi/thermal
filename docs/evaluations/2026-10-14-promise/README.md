# Reproduce the promise audit

Candidate: `302501b17f96a26a13b7b5cbaf4e7654199f7d5c`, 2026-10-14.
Application sources were unchanged. The working README contained a pre-existing 26-line addition.
The Go test sources here are one-off audit probes, not production tests or implementation fixes.
They intentionally fail on the evaluated candidate. Only synthetic data is used.

From `/mnt/Jad/github/lab/thermal-streak`:

```bash
export PATH=/home/jad/.local/share/mise/installs/go/1.27.2/bin:$PATH
export GOCACHE=/tmp/thermal-eval-gocache
unset GOROOT
mkdir -p /tmp/thermal-promise-probes
cat > /tmp/thermal-promise-probes/overlay.json <<'JSON'
{"Replace":{"/mnt/Jad/github/lab/thermal-streak/internal/thermal/promise_probe_test.go":"/mnt/Jad/github/lab/thermal-streak/docs/evaluations/2026-10-14-promise/thermal_test.go.txt","/mnt/Jad/github/lab/thermal-streak/internal/audit/promise_probe_test.go":"/mnt/Jad/github/lab/thermal-streak/docs/evaluations/2026-10-14-promise/audit_test.go.txt"}}
JSON
go test -overlay /tmp/thermal-promise-probes/overlay.json \
  ./internal/thermal ./internal/audit -run TestPromise -v
```

Expected exit on this candidate: 1. The log is `probes.txt`.
The receipt probe has three passing controls and three failures. Replay has three failing cases;
audit has one failing case. These establish defects, not their production frequency.

Existing scripts were copied to `/tmp/thermal-promise-probes/` with only `ROOT_DIR` fixed to this
checkout and `BIN_PATH` redirected to `/tmp/thermal-promise-{gate,bench}`. This avoids loose binaries
in the repository. All fixture generation, assertions and timing behavior remained unchanged.
Run the original scripts with the same substitutions to reproduce `benchmark.txt` and
`simulated-gate.txt`. The gate needs permission to bind a loopback socket.

The benchmark exited 0. Timing ran alongside verification work, so its 10ms p50 and 16ms p95 are
observations under contention, not an isolated performance regression claim. The fixture is small;
its two-second CPU samples do not establish sustained or large-history performance.

Terminal logs retain ANSI color output; trailing empty lines were removed for Git whitespace checks.
The empty `vet.txt` records a silent successful vet run; exit 0 is recorded in the evaluation.
