package render

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func sampleMixReport() thermal.MixReport {
	return thermal.MixReport{
		Type:   "mix",
		By:     "tool",
		Metric: "tokens",
		Series: []thermal.MixSeries{
			{Name: "OpenCode", Total: 600, Share: 0.6, Buckets: []float64{400, 200}},
			{Name: "Codex", Total: 400, Share: 0.4, Buckets: []float64{100, 300}},
		},
		Buckets: []thermal.MixBucket{
			{Period: "2026-08-30", Values: []float64{400, 100}, Total: 500},
			{Period: "2026-09-06", Values: []float64{200, 300}, Total: 500},
		},
		Dominant:        "OpenCode",
		Switches:        3,
		SwitchesByMonth: map[string]int{"2026-09": 3},
		Concentration:   0.52,
		ActiveSeries:    2,
		Total:           1000,
	}
}

func TestRenderMix_Table(t *testing.T) {
	out := RenderMix(sampleMixReport(), true)
	for _, want := range []string{
		"mix", "by tool", "2026-08-30", "2026-09-06", "OpenCode", "Codex",
		"80%", "20%", "Total", "Dominant", "Switches", "3 in 2026-09", "0.52",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
	if strings.Contains(out, "estimated") {
		t.Errorf("tokens metric should not mention estimates:\n%s", out)
	}
}

func TestRenderMix_EstimatedFootnoteAndOverflow(t *testing.T) {
	rep := sampleMixReport()
	rep.Metric = "cost"
	rep.Estimated = true
	rep.Series = append(rep.Series,
		thermal.MixSeries{Name: "Devin", Total: 50, Share: 0.05, Buckets: []float64{0, 0}},
		thermal.MixSeries{Name: "Grok", Total: 30, Share: 0.03, Buckets: []float64{0, 0}},
		thermal.MixSeries{Name: "Muse", Total: 20, Share: 0.02, Buckets: []float64{0, 0}},
	)
	out := RenderMix(rep, true)
	if !strings.Contains(out, "other (1)") {
		t.Errorf("expected an other bucket for the tail series:\n%s", out)
	}
	if !strings.Contains(out, "estimates from models.dev") {
		t.Errorf("expected the estimate footnote:\n%s", out)
	}
	if !strings.Contains(out, "$") {
		t.Errorf("cost metric should format values as currency:\n%s", out)
	}
}

func TestRenderMix_Empty(t *testing.T) {
	out := RenderMix(thermal.MixReport{Type: "mix", By: "model", Metric: "tokens"}, true)
	if !strings.Contains(out, "No activity") {
		t.Errorf("expected empty notice:\n%s", out)
	}
}

func TestRenderStats_SummaryAndHistogram(t *testing.T) {
	rep := thermal.StatsReport{
		Type:      "stats",
		Metric:    "tokens",
		Days:      5,
		Total:     1100,
		Mean:      220,
		Median:    30,
		P90:       40,
		Max:       1000,
		Weekday:   []thermal.WeekdayStat{{Weekday: "Monday", Mean: 25, Days: 2}},
		TopDays:   []thermal.DayValue{{Day: "2026-09-11", Value: 1000}},
		Outliers:  []thermal.DayValue{{Day: "2026-09-11", Value: 1000}},
		Histogram: []thermal.HistogramBin{{From: 0, To: 100, Count: 4}, {From: 900, To: 1000, Count: 1}},
	}
	rep.OutlierThreshold = 50
	out := RenderStats(rep, true)
	for _, want := range []string{"stats", "Active days", "Median", "p90", "Monday", "Distribution", "█", "Top days", "Outliers above", "2026-09-11"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestRenderStats_Empty(t *testing.T) {
	out := RenderStats(thermal.StatsReport{Type: "stats", Metric: "cost"}, true)
	if !strings.Contains(out, "No activity") {
		t.Errorf("expected empty notice:\n%s", out)
	}
}

func TestRenderTrend_SummaryAndProjection(t *testing.T) {
	rep := thermal.TrendReport{
		Type:      "trend",
		Metric:    "tokens",
		Mean:      25,
		Slope:     10,
		Intercept: 10,
		FirstDay:  "2026-09-01",
		LastDay:   "2026-09-04",
		Points:    []thermal.DayValue{{Day: "2026-09-01", Value: 10}, {Day: "2026-09-04", Value: 40}},
		Projection: &thermal.Projection{
			To: "2026-09-30", Days: 26, Expected: 300, Low: 250, High: 350,
		},
	}
	out := RenderTrend(rep, true)
	for _, want := range []string{"trend", "2026-09-01 to 2026-09-04", "+10 per day", "rising", "Projection to 2026-09-30", "300", "250", "350"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestRenderTrend_FallingAndEmpty(t *testing.T) {
	falling := RenderTrend(thermal.TrendReport{
		Type: "trend", Metric: "tokens", Mean: 100, Slope: -10,
		FirstDay: "2026-09-01", LastDay: "2026-09-02",
		Points: []thermal.DayValue{{Day: "2026-09-01", Value: 100}},
	}, true)
	if !strings.Contains(falling, "falling") || !strings.Contains(falling, "-10") {
		t.Errorf("expected a falling trend with a negative slope:\n%s", falling)
	}
	if out := RenderTrend(thermal.TrendReport{Type: "trend", Metric: "tokens"}, true); !strings.Contains(out, "No activity") {
		t.Errorf("expected empty notice:\n%s", out)
	}
}
