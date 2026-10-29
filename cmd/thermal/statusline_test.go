// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/thermal"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = outW, errW
	defer func() {
		os.Stdout, os.Stderr = oldOut, oldErr
	}()

	fn()
	_ = outW.Close()
	_ = errW.Close()
	out, _ := io.ReadAll(outR)
	errOut, _ := io.ReadAll(errR)
	_ = outR.Close()
	_ = errR.Close()
	return string(out), string(errOut)
}

func isolateStatuslineHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	for _, key := range []string{"HOME", "USERPROFILE", "XDG_DATA_HOME", "HERMES_HOME", "DSH_HOME", "CODEX_HOME", "OPENCODE_HOME", "ZED_HOME", "AGY_HOME"} {
		t.Setenv(key, "")
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_DATA_HOME", home)
	return home
}

func TestStatuslineNoDataIsSilent(t *testing.T) {
	isolateStatuslineHome(t)
	out, errOut := captureOutput(t, func() {
		runStatusline(thermal.Options{Tool: "all"})
	})
	if out != "" || errOut != "" {
		t.Fatalf("statusline no-data stdout=%q stderr=%q, want both empty", out, errOut)
	}
}

func TestStatuslineNoDataJSON(t *testing.T) {
	isolateStatuslineHome(t)
	var out string
	var errOut string
	out, errOut = captureOutput(t, func() {
		runStatusline(thermal.Options{Tool: "all", JSON: true})
	})
	if errOut != "" {
		t.Fatalf("statusline --json stderr=%q", errOut)
	}
	var payload render.StatuslineData
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if payload.State != "no-data" || payload.TodayTokens != 0 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestStatuslineNoDataVerbose(t *testing.T) {
	isolateStatuslineHome(t)
	out, errOut := captureOutput(t, func() {
		runStatusline(thermal.Options{Tool: "all", Verbose: true})
	})
	if out != "" {
		t.Fatalf("statusline --verbose stdout=%q", out)
	}
	if !strings.Contains(errOut, "statusline: no data") {
		t.Fatalf("statusline --verbose missing diagnostic: %q", errOut)
	}
}

func TestStatuslineNoDataPlain(t *testing.T) {
	isolateStatuslineHome(t)
	out, errOut := captureOutput(t, func() {
		runStatusline(thermal.Options{Tool: "all", Plain: true})
	})
	if out != "" || errOut != "" {
		t.Fatalf("statusline --plain stdout=%q stderr=%q", out, errOut)
	}
}
