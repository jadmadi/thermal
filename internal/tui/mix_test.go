// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/thermal"
)

func renderMixFixture(t *testing.T, width, height int, tab int, by string) string {
	t.Helper()
	m := New(projectsFixture(), false)
	m.tab = tab
	if by != "" {
		m.mixBy = by
	}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	view := updated.(Model).View()
	return ansi.Strip(view.Content)
}

// TestMixParityAgainstStatic proves the view reports the same concentration and
// switch count as the aggregation behind `thermal mix`. Parity is asserted
// against the function, so a change in either layer breaks the test.
func TestMixParityAgainstStatic(t *testing.T) {
	a := projectsFixture()
	mv := a.BuildMix(Range30d, MetricTokens)

	batches := make([]thermal.ToolDays, 0, len(a.Tools))
	for _, tool := range a.Tools {
		batches = append(batches, thermal.ToolDays{Tool: tool.Name, Days: tool.Days})
	}
	rep := thermal.AggregateToolMix(batches, thermal.MixOptions{
		Since:  mv.Start,
		Until:  mv.End,
		Grain:  mv.Grain,
		By:     "tool",
		Metric: string(mv.Metric),
	}, a.Pricer)

	if mv.Dominant != rep.Dominant {
		t.Errorf("dominant = %q, static = %q", mv.Dominant, rep.Dominant)
	}
	if mv.Switches != rep.Switches {
		t.Errorf("switches = %d, static = %d", mv.Switches, rep.Switches)
	}
	if diff := mv.Conc - rep.Concentration; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("concentration = %.6f, static = %.6f", mv.Conc, rep.Concentration)
	}
	if mv.Active != rep.ActiveSeries {
		t.Errorf("active series = %d, static = %d", mv.Active, rep.ActiveSeries)
	}
}

// stubPricer charges a flat rate per million tokens, so the cost path has
// numbers to divide without pulling in the real catalog.
type stubPricer struct{}

func (stubPricer) PriceDay(day thermal.DailyRow) (float64, []string) {
	return float64(day.Tokens) / 1e6 * 3, nil
}

// TestMixSharesSumTo100 is the invariant the goal names, including the other
// bucket. A share table that does not reconcile is worse than no share table.
func TestMixSharesSumTo100(t *testing.T) {
	a := manySeriesFixture()
	a.Pricer = stubPricer{}
	for _, m := range []Metric{MetricTokens, MetricCost} {
		mv := a.BuildMix(Range30d, m)
		var sum float64
		for _, s := range mv.Series {
			sum += s.Share
		}
		if diff := sum - 100; diff > 0.01 || diff < -0.01 {
			t.Errorf("metric %s: shares sum to %.4f, want 100", m, sum)
		}
		if len(mv.Series) > MixSeriesCap+1 {
			t.Errorf("metric %s: %d series on screen, cap is %d plus other",
				m, len(mv.Series), MixSeriesCap)
		}
		if len(mv.Series) == MixSeriesCap+1 && mv.Series[MixSeriesCap].Name != "other" {
			t.Errorf("metric %s: the folded bucket is named %q, want other",
				m, mv.Series[MixSeriesCap].Name)
		}
	}
}

// TestMixBucketValuesMatchSeries verifies the chart and the legend agree: the
// per-bucket cells must add up to the same totals the legend shows.
func TestMixBucketValuesMatchSeries(t *testing.T) {
	a := projectsFixture()
	mv := a.BuildMix(Range30d, MetricTokens)

	perSeries := make([]float64, len(mv.Series))
	perBucket := make([]float64, len(mv.Buckets))
	for bi, bucket := range mv.Buckets {
		var bucketSum float64
		for si, v := range bucket.Values {
			if si < len(perSeries) {
				perSeries[si] += v
			}
			bucketSum += v
		}
		perBucket[bi] = bucketSum
		if diff := bucketSum - bucket.Total; diff > 0.5 || diff < -0.5 {
			t.Errorf("bucket %s: cells sum to %.2f, total says %.2f", bucket.Period, bucketSum, bucket.Total)
		}
	}
	for si, s := range mv.Series {
		if si >= len(perSeries) {
			break
		}
		// The other bucket is a sum by construction, so it is checked above.
		if s.Name == "other" {
			continue
		}
		if diff := perSeries[si] - s.Total; diff > 0.5 || diff < -0.5 {
			t.Errorf("series %s: buckets sum to %.2f, series total says %.2f", s.Name, perSeries[si], s.Total)
		}
	}
}

// TestModeslParityAgainstStatic proves the Models view reuses the static
// aggregation rather than deriving its own numbers.
func TestModelsParityAgainstStatic(t *testing.T) {
	a := projectsFixture()
	mv := a.BuildModels(Range30d, MetricTokens, ModelSortTokens)

	batches := make([]thermal.ToolDays, 0, len(a.Tools))
	for _, tool := range a.Tools {
		batches = append(batches, thermal.ToolDays{Tool: tool.Name, Days: tool.Days})
	}
	rep := thermal.AggregateModels(batches, thermal.ModelOptions{Since: mv.Start, Until: mv.End}, a.Pricer)

	if mv.Tokens != rep.Totals.Tokens {
		t.Errorf("tokens = %d, static = %d", mv.Tokens, rep.Totals.Tokens)
	}
	if diff := mv.Cost - rep.Totals.Cost; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("cost = %.4f, static = %.4f", mv.Cost, rep.Totals.Cost)
	}
	if len(mv.Rows) != len(rep.Rows) {
		t.Fatalf("rows = %d, static = %d", len(mv.Rows), len(rep.Rows))
	}
	byName := make(map[string]thermal.ModelRow, len(rep.Rows))
	for _, r := range rep.Rows {
		byName[r.Model] = r
	}
	for _, row := range mv.Rows {
		want, ok := byName[row.Name]
		if !ok {
			t.Errorf("view has model %s, static does not", row.Name)
			continue
		}
		if row.Tokens != want.Tokens {
			t.Errorf("%s tokens: view %d, static %d", row.Name, row.Tokens, want.Tokens)
		}
		if diff := row.Cost - want.Cost; diff > 0.0001 || diff < -0.0001 {
			t.Errorf("%s cost: view %.4f, static %.4f", row.Name, row.Cost, want.Cost)
		}
	}
}

// TestModelsSortOrders proves each key reorders.
func TestModelsSortOrders(t *testing.T) {
	a := projectsFixture()
	byTokens := a.BuildModels(Range30d, MetricTokens, ModelSortTokens)
	if len(byTokens.Rows) < 2 {
		t.Skip("need at least two models")
	}
	byName := a.BuildModels(Range30d, MetricTokens, ModelSortName)
	for i := 1; i < len(byName.Rows); i++ {
		if byName.Rows[i-1].Name > byName.Rows[i].Name {
			t.Errorf("name sort ascending broken: %s before %s", byName.Rows[i-1].Name, byName.Rows[i].Name)
		}
	}
	byCost := a.BuildModels(Range30d, MetricTokens, ModelSortCost)
	for i := 1; i < len(byCost.Rows); i++ {
		if byCost.Rows[i-1].Cost < byCost.Rows[i].Cost {
			t.Errorf("cost sort descending broken: %.4f before %.4f", byCost.Rows[i-1].Cost, byCost.Rows[i].Cost)
		}
	}
}

// TestMixAndModelsGolden pins both frames at two sizes.
func TestMixAndModelsGolden(t *testing.T) {
	for _, tc := range []struct {
		name string
		w, h int
		tab  int
		by   string
		file string
	}{
		{"mix tool 80x24", 80, 24, tabMix, "tool", "mix_80x24.golden"},
		{"mix tool 120x40", 120, 40, tabMix, "tool", "mix_120x40.golden"},
		{"mix model 120x40", 120, 40, tabMix, "model", "mix_models_120x40.golden"},
		{"models 80x24", 80, 24, tabModels, "", "models_80x24.golden"},
		{"models 120x40", 120, 40, tabModels, "", "models_120x40.golden"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderMixFixture(t, tc.w, tc.h, tc.tab, tc.by)
			path := filepath.Join("testdata", tc.file)
			if *update {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("golden missing: %v (run: go test ./internal/tui -update)", err)
			}
			if got != string(want) {
				t.Errorf("frame changed at %s\n--- got ---\n%s\n--- want ---\n%s", tc.name, got, want)
			}
		})
	}
}

// TestMixKeyFlow covers the keys the Mix and Models views own.
func TestMixKeyFlow(t *testing.T) {
	t.Run("v switches the mix subject", func(t *testing.T) {
		m := New(projectsFixture(), false)
		m.tab = tabMix
		if m.mixBy != "tool" {
			t.Fatalf("default subject = %q, want tool", m.mixBy)
		}
		m = press(t, m, "v")
		if m.mixBy != "model" {
			t.Fatalf("after v: subject = %q, want model", m.mixBy)
		}
		m = press(t, m, "v")
		if m.mixBy != "tool" {
			t.Errorf("after two v: subject = %q, want tool", m.mixBy)
		}
	})

	t.Run("s cycles the model sort", func(t *testing.T) {
		m := New(projectsFixture(), false)
		m.tab = tabModels
		seen := map[ModelSort]bool{m.modelsSort: true}
		for i := 0; i < len(ModelSortOrder); i++ {
			m = press(t, m, "s")
			seen[m.modelsSort] = true
		}
		for _, want := range ModelSortOrder {
			if !seen[want] {
				t.Errorf("cycling never reached %s", want)
			}
		}
	})

	t.Run("j and k move the model selection", func(t *testing.T) {
		m := New(projectsFixture(), false)
		m.tab = tabModels
		m = press(t, m, "j")
		if m.modelsSel != 1 {
			t.Errorf("j: selection = %d, want 1", m.modelsSel)
		}
		m = press(t, m, "k")
		if m.modelsSel != 0 {
			t.Errorf("k: selection = %d, want 0", m.modelsSel)
		}
	})

	t.Run("the subject survives a tab round trip", func(t *testing.T) {
		m := New(projectsFixture(), false)
		m.tab = tabMix
		m = press(t, m, "v")
		m = press(t, m, "1")
		m = press(t, m, "3")
		if m.mixBy != "model" {
			t.Errorf("subject after a tab round trip = %q, want model", m.mixBy)
		}
	})
}

// TestMixGrainFollowsRange pins the bucket size, because a daily grain across a
// year would print hundreds of columns.
func TestMixGrainFollowsRange(t *testing.T) {
	for _, tc := range []struct {
		r    Range
		want thermal.Grain
	}{
		{Range30d, thermal.GrainDay},
		{Range90d, thermal.GrainWeek},
		{Range1y, thermal.GrainMonth},
		{RangeAll, thermal.GrainMonth},
	} {
		if got := mixGrain(tc.r); got != tc.want {
			t.Errorf("grain for %s = %s, want %s", tc.r, got, tc.want)
		}
	}
}

// manySeriesFixture has more tools than the series cap, so the other bucket and
// the share invariant are exercised.
func manySeriesFixture() Adapter {
	names := []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot"}
	a := Adapter{Latest: "2026-09-18", Earliest: "2025-09-20"}
	for i, name := range names {
		// Each day names its model, because the cost path only prices days the
		// source attributes to a model.
		a.Tools = append(a.Tools, ToolDay{
			Name: name,
			Days: []thermal.DailyRow{
				{
					Day: "2026-09-16", Tokens: int64((i + 1) * 1_000_000), Turns: 1,
					Models: map[string]thermal.ModelTokens{
						"stub-model": {Input: int64((i + 1) * 1_000_000)},
					},
				},
				{
					Day: "2026-09-17", Tokens: int64((i + 1) * 2_000_000), Turns: 1,
					Models: map[string]thermal.ModelTokens{
						"stub-model": {Input: int64((i + 1) * 2_000_000)},
					},
				},
			},
		})
	}
	return a
}

var _ = strings.Contains
