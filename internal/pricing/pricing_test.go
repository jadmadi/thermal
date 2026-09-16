package pricing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func testCatalog() *Catalog {
	return NewCatalog(map[string]Price{
		"claude-sonnet-4-5": {Input: 3, Output: 15, CacheRead: 0.3, CacheWrite: 3.75},
		"gpt-5.4":           {Input: 2.5, Output: 15},
		"free-model":        {},
	})
}

func TestLookupCandidates(t *testing.T) {
	cat := testCatalog()
	cases := []string{
		"claude-sonnet-4-5",
		"Claude-Sonnet-4-5",
		"anthropic/claude-sonnet-4-5",
		"accounts/fireworks/models/claude-sonnet-4-5",
		"claude-sonnet-4-5-20250929",
		"claude-sonnet-4-5:free",
		"claude-sonnet-4-5-latest",
	}
	for _, name := range cases {
		if _, ok := cat.Lookup(name); !ok {
			t.Errorf("Lookup(%q) failed", name)
		}
	}
	if _, ok := cat.Lookup("unknown-model"); ok {
		t.Error("expected Lookup to reject an unknown model")
	}
	if _, ok := cat.Lookup(""); ok {
		t.Error("expected Lookup to reject an empty name")
	}
}

func TestLookupPrefersOriginalName(t *testing.T) {
	cat := NewCatalog(map[string]Price{
		"vendor/model": {Input: 1},
		"model":        {Input: 2},
	})
	p, ok := cat.Lookup("vendor/model")
	if !ok || p.Input != 1 {
		t.Errorf("Lookup(vendor/model) = %v, %v; want the full-name price", p, ok)
	}
}

func TestPriceFallbacks(t *testing.T) {
	cat := testCatalog()
	p, _ := cat.Lookup("gpt-5.4")
	// One million input tokens at 2.5 and one million cache-read tokens at
	// the 10 percent fallback rate.
	day := thermal.DailyRow{
		Tokens: 2_000_000,
		Models: map[string]thermal.ModelTokens{
			"gpt-5.4": {Input: 1_000_000, CacheRead: 1_000_000},
		},
	}
	cost, missing := cat.PriceDay(day)
	if len(missing) != 0 {
		t.Fatalf("unexpected missing models: %v", missing)
	}
	if cost != 2.5+0.25 {
		t.Errorf("cost = %v, want 2.75", cost)
	}
	if p.CacheWrite != 2.5 {
		t.Errorf("cache write fallback = %v, want the input rate", p.CacheWrite)
	}
}

func TestPriceDayTypesAndMissing(t *testing.T) {
	cat := testCatalog()
	day := thermal.DailyRow{
		Tokens: 3_000_000,
		Models: map[string]thermal.ModelTokens{
			"claude-sonnet-4-5": {
				Input:      1_000_000,
				Output:     500_000,
				Reasoning:  500_000,
				CacheRead:  500_000,
				CacheWrite: 500_000,
			},
			"mystery-model": {Input: 1000},
		},
	}
	cost, missing := cat.PriceDay(day)
	want := 3*1 + 15*1 + 0.3*0.5 + 3.75*0.5
	if cost != want {
		t.Errorf("cost = %v, want %v", cost, want)
	}
	if len(missing) != 1 || missing[0] != "mystery-model" {
		t.Errorf("missing = %v, want [mystery-model]", missing)
	}
}

func TestPriceDaySkipsUnclassifiedAndEmpty(t *testing.T) {
	cat := testCatalog()
	day := thermal.DailyRow{
		Tokens: 2_000_000,
		Models: map[string]thermal.ModelTokens{
			"codewhale-session": {Unclassified: 1_000_000},
			"gpt-5.4":           {},
		},
	}
	cost, missing := cat.PriceDay(day)
	if cost != 0 {
		t.Errorf("cost = %v, want 0", cost)
	}
	if len(missing) != 1 || missing[0] != "codewhale-session" {
		t.Errorf("missing = %v, want [codewhale-session]", missing)
	}
}

func TestReadCacheRejectsOldVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pricing.json")
	if err := os.WriteFile(path, []byte(`{"version":0,"models":{"a":{"input":1}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if cat := readCache(path); cat != nil {
		t.Errorf("expected old cache versions to be ignored, got %+v", cat)
	}
}

func TestOverridesWin(t *testing.T) {
	dir := t.TempDir()
	overrides := filepath.Join(dir, "pricing.json")
	if err := os.WriteFile(overrides, []byte(`{"gpt-5.4":{"input":9,"output":9}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cat := testCatalog()
	applyOverrides(cat, overrides)
	p, _ := cat.Lookup("gpt-5.4")
	if p.Input != 9 || p.Output != 9 {
		t.Errorf("override not applied: %+v", p)
	}
	// Unrelated models keep catalog values.
	if p, _ := cat.Lookup("claude-sonnet-4-5"); p.Input != 3 {
		t.Errorf("catalog entry changed: %+v", p)
	}
}

func TestBuildCatalogProviderPreference(t *testing.T) {
	providers := map[string]modelsDevProvider{
		"some-coding-plan": {Models: map[string]modelsDevModel{
			"glm-5.3": {Cost: &modelsDevCost{Input: 0, Output: 0}},
		}},
		"acme-reseller": {Models: map[string]modelsDevModel{
			"glm-5.3": {Cost: &modelsDevCost{Input: 9, Output: 9}},
		}},
		"zhipuai": {Models: map[string]modelsDevModel{
			"glm-5.3": {Cost: &modelsDevCost{Input: 1.4, Output: 4.4}},
		}},
		"no-cost-entry": {Models: map[string]modelsDevModel{
			"ghost": {},
		}},
	}
	cat := buildCatalog(providers)
	p, ok := cat["glm-5.3"]
	if !ok || p.Input != 1.4 {
		t.Errorf("glm-5.3 = %+v, %v; want the first-party price", p, ok)
	}
	if _, ok := cat["ghost"]; ok {
		t.Error("models without a cost block must be skipped")
	}
	if diff := p.CacheRead - 0.14; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cache read fallback = %v, want 0.14", p.CacheRead)
	}
	if p.CacheWrite != 1.4 {
		t.Errorf("cache write fallback = %v, want the input rate", p.CacheWrite)
	}
}

func TestBuildCatalogAlias(t *testing.T) {
	providers := map[string]modelsDevProvider{
		"google": {Models: map[string]modelsDevModel{
			"gemini-3-pro-preview": {Cost: &modelsDevCost{Input: 1}},
		}},
	}
	cat := buildCatalog(providers)
	if _, ok := cat["gemini-3-pro-high"]; !ok {
		t.Errorf("expected the variant alias to resolve, got %v", cat)
	}
}
