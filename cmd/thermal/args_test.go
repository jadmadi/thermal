package main

import (
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
