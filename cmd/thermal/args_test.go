package main

import (
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
