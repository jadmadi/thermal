// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestIsNegativeNumber(t *testing.T) {
	for _, s := range []string{"-1", "-42", "-0"} {
		if !isNegativeNumber(s) {
			t.Errorf("isNegativeNumber(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"1", "", "-", "--json", "-1.5", "-a"} {
		if isNegativeNumber(s) {
			t.Errorf("isNegativeNumber(%q) = true, want false", s)
		}
	}
}

func TestValidateReportFlags(t *testing.T) {
	base := thermal.Options{Order: "desc", StartOfWeek: "sunday"}

	cases := []struct {
		name    string
		opts    thermal.Options
		wantErr bool
	}{
		{"plain dashboard", base, false},
		{"report ok", with(base, func(o *thermal.Options) { o.Report = "weekly" }), false},
		{"report flag without report", with(base, func(o *thermal.Options) { o.Since = "2026-01-01" }), true},
		{"breakdown without report", with(base, func(o *thermal.Options) { o.Breakdown = true }), true},
		{"non-default order without report", with(base, func(o *thermal.Options) { o.Order = "asc" }), true},
		{"last with since", with(base, func(o *thermal.Options) { o.Report = "daily"; o.Last = 2; o.Since = "2026-01-01" }), true},
		{"negative last", with(base, func(o *thermal.Options) { o.Report = "daily"; o.Last = -1 }), true},
		{"bad order", with(base, func(o *thermal.Options) { o.Report = "daily"; o.Order = "sideways" }), true},
		{"bad weekday", with(base, func(o *thermal.Options) { o.Report = "daily"; o.StartOfWeek = "someday" }), true},
		{"bad since", with(base, func(o *thermal.Options) { o.Report = "daily"; o.Since = "01/02/2026" }), true},
		{"compact since ok", with(base, func(o *thermal.Options) { o.Report = "monthly"; o.Since = "20260101"; o.Until = "2026-01-31" }), false},
		{"mix ok", with(base, func(o *thermal.Options) {
			o.Report = "mix"
			o.Metric = "cost"
			o.By = "model"
			o.Grain = "month"
		}), false},
		{"stats ok", with(base, func(o *thermal.Options) { o.Report = "stats"; o.Metric = "cost" }), false},
		{"trend ok", with(base, func(o *thermal.Options) { o.Report = "trend"; o.Last = 30 }), false},
		{"bad metric", with(base, func(o *thermal.Options) { o.Report = "stats"; o.Metric = "steps" }), true},
		{"bad by", with(base, func(o *thermal.Options) { o.Report = "mix"; o.By = "project" }), true},
		{"bad grain", with(base, func(o *thermal.Options) { o.Report = "mix"; o.Grain = "quarter" }), true},
		{"by on trend", with(base, func(o *thermal.Options) { o.Report = "trend"; o.By = "model" }), true},
		{"grain on stats", with(base, func(o *thermal.Options) { o.Report = "stats"; o.Grain = "day" }), true},
		{"metric on leaderboard", with(base, func(o *thermal.Options) { o.Metric = "cost" }), true},
		{"grain on weekly", with(base, func(o *thermal.Options) { o.Report = "weekly"; o.Grain = "month" }), true},
		{"replay ok", with(base, func(o *thermal.Options) { o.Report = "replay"; o.Against = "claude-sonnet" }), false},
		{"replay compare ok", with(base, func(o *thermal.Options) { o.Report = "replay"; o.Compare = "claude-pro,cursor-pro" }), false},
		{"replay bad sort", with(base, func(o *thermal.Options) { o.Report = "replay"; o.Sort = "tokens" }), true},
		{"replay bad chart", with(base, func(o *thermal.Options) { o.Report = "replay"; o.Chart = true }), true},
		{"against on weekly", with(base, func(o *thermal.Options) { o.Report = "weekly"; o.Against = "claude-sonnet" }), true},
		{"compare on leaderboard", with(base, func(o *thermal.Options) { o.Compare = "claude-pro" }), true},
	}

	for _, tc := range cases {
		err := validateReportFlags(tc.opts)
		if tc.wantErr && err == nil {
			t.Errorf("%s: expected an error, got none", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
		}
	}
}

func with(o thermal.Options, mutate func(*thermal.Options)) thermal.Options {
	mutate(&o)
	return o
}

func TestPrintToolWarnings(t *testing.T) {
	warnings := []string{
		"path/to/sess.jsonl: line exceeds 512 byte ceiling, rest of file skipped",
	}

	// When verbose is false, nothing is written.
	var nonVerboseBuf strings.Builder
	printToolWarningsTo(&nonVerboseBuf, "command-code", warnings, false)
	if nonVerboseBuf.Len() != 0 {
		t.Errorf("expected no output when verbose is false, got: %q", nonVerboseBuf.String())
	}

	// When verbose is true, formatted warning is written.
	var verboseBuf strings.Builder
	printToolWarningsTo(&verboseBuf, "command-code", warnings, true)
	expected := "thermal: warning: command-code: path/to/sess.jsonl: line exceeds 512 byte ceiling, rest of file skipped\n"
	if verboseBuf.String() != expected {
		t.Errorf("expected %q, got %q", expected, verboseBuf.String())
	}
}

func TestRunLicenseJSON(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	runLicense(thermal.Options{JSON: true})
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	var payload struct {
		Name     string `json:"name"`
		Author   string `json:"author"`
		License  string `json:"license"`
		Covenant struct {
			IrrevocableAGPL bool   `json:"irrevocable_agpl"`
			NonRelicensing  bool   `json:"non_relicensing_of_contributions"`
			Description     string `json:"description"`
		} `json:"covenant"`
	}
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	if payload.Name != "thermal" {
		t.Errorf("expected name thermal, got %q", payload.Name)
	}
	if !payload.Covenant.IrrevocableAGPL {
		t.Errorf("expected covenant irrevocable_agpl to be true")
	}
	if !payload.Covenant.NonRelicensing {
		t.Errorf("expected covenant non_relicensing_of_contributions to be true")
	}
}

func TestEmitDeprecationWarning(t *testing.T) {
	// 1. With noColor = true: plain text, no ANSI escape sequences
	var plainBuf bytes.Buffer
	emitDeprecationWarningTo(&plainBuf, "--license", "thermal license", "0.14.0", "deprecated-license-flag", true)
	expectedPlain := "thermal: warning: --license is deprecated and will be removed in v0.14.0; use thermal license (see docs/MIGRATION.md#deprecated-license-flag)\n"
	if plainBuf.String() != expectedPlain {
		t.Errorf("expected plain warning %q, got %q", expectedPlain, plainBuf.String())
	}

	// 2. With noColor = false and non-file writer: ANSI yellow escape sequences present
	var colorBuf bytes.Buffer
	emitDeprecationWarningTo(&colorBuf, "nous", "hermes", "0.14.0", "deprecated-tool-nous", false)
	expectedColor := "\033[33mthermal: warning:\033[0m nous is deprecated and will be removed in v0.14.0; use hermes (see docs/MIGRATION.md#deprecated-tool-nous)\n"
	if colorBuf.String() != expectedColor {
		t.Errorf("expected color warning %q, got %q", expectedColor, colorBuf.String())
	}
}

func TestDeprecationWarningStderrWithJSONStdout(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Simulate emitting deprecation warning to stderr while executing command outputting JSON to stdout
	emitDeprecationWarning("--license", "thermal license", "0.14.0", "deprecated-license-flag", true)
	runLicense(thermal.Options{JSON: true})

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, stdoutR)
	_, _ = io.Copy(&stderrBuf, stderrR)

	// Verify stderr got the warning
	if !strings.Contains(stderrBuf.String(), "thermal: warning: --license is deprecated") {
		t.Errorf("expected stderr to contain deprecation warning, got %q", stderrBuf.String())
	}

	// Verify stdout is 100% clean valid JSON without stderr contamination
	var payload map[string]any
	if err := json.Unmarshal(stdoutBuf.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed on stdout (stderr pollution?): %v; stdout was: %q", err, stdoutBuf.String())
	}
	if payload["name"] != "thermal" {
		t.Errorf("expected payload name thermal, got %v", payload["name"])
	}
}
