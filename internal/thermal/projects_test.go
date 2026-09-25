// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectKeyGitRoot(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	sub := filepath.Join(repo, "apps", "web")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := ProjectKey(sub); got != repo {
		t.Errorf("ProjectKey(%q) = %q, want the repository root %q", sub, got, repo)
	}
	if got := ProjectKey(repo); got != repo {
		t.Errorf("ProjectKey(%q) = %q, want %q", repo, got, repo)
	}

	// A stray empty .git directory is not a repository and must not swallow
	// every path below it.
	stray := filepath.Join(root, "stray")
	strayDeep := filepath.Join(stray, "deep")
	if err := os.MkdirAll(filepath.Join(stray, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(strayDeep, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := ProjectKey(strayDeep); got != strayDeep {
		t.Errorf("ProjectKey(%q) = %q, want the path itself", strayDeep, got)
	}

	// A worktree marks its root with a .git file rather than a directory.
	worktree := filepath.Join(root, "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: /elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ProjectKey(worktree); got != worktree {
		t.Errorf("ProjectKey(%q) = %q, want the worktree root", worktree, got)
	}
}

func TestProjectKeyResolvesSymlinks(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "real", "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "link")
	if err := os.Symlink(filepath.Join(root, "real"), link); err != nil {
		t.Fatal(err)
	}

	realKey := ProjectKey(repo)
	linkedKey := ProjectKey(filepath.Join(link, "repo"))
	if realKey != linkedKey {
		t.Errorf("symlinked path key %q != real path key %q", linkedKey, realKey)
	}
	if realKey != repo {
		t.Errorf("key = %q, want the repository root %q", realKey, repo)
	}
}

func TestProjectKeyFallback(t *testing.T) {
	root := t.TempDir()
	plain := filepath.Join(root, "plain", "nested")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := ProjectKey(plain); got != plain {
		t.Errorf("ProjectKey(%q) = %q, want the cleaned path", plain, got)
	}
	if got := ProjectKey("   "); got != "" {
		t.Errorf("ProjectKey(blank) = %q, want empty", got)
	}
}

func TestAggregateProjectsMergesToolsAndWindows(t *testing.T) {
	days := []ProjectDay{
		{Project: "/repo/a", Day: "2026-09-10", Tool: "OpenCode", Tokens: 100, Input: 100, Turns: 2, Cost: 1.5},
		{Project: "/repo/a", Day: "2026-09-11", Tool: "Claude", Tokens: 50, Input: 50, Turns: 1},
		{Project: "/repo/b", Day: "2026-09-11", Tool: "Codex", Tokens: 300, Input: 300, Turns: 3, Cost: 2.0},
		{Project: "/repo/c", Day: "2026-08-01", Tool: "Codex", Tokens: 900, Input: 900, Turns: 9},
	}

	all := AggregateProjects(days, ProjectOptions{}, nil)
	if len(all.Rows) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(all.Rows))
	}
	// Largest first by default.
	if all.Rows[0].Project != "/repo/c" || all.Rows[0].Tokens != 900 {
		t.Errorf("largest project first expected: %+v", all.Rows[0])
	}
	asc := AggregateProjects(days, ProjectOptions{Order: "asc"}, nil)
	if asc.Rows[0].Project != "/repo/a" {
		t.Errorf("ascending order wrong: %+v", asc.Rows[0])
	}

	windowed := AggregateProjects(days, ProjectOptions{Since: "2026-09-11", Until: "2026-09-11"}, nil)
	if len(windowed.Rows) != 2 {
		t.Fatalf("expected 2 projects in window, got %d", len(windowed.Rows))
	}
	// Sorted largest first: /repo/b then /repo/a.
	if windowed.Rows[0].Project != "/repo/b" || windowed.Rows[1].Project != "/repo/a" {
		t.Errorf("token ordering wrong: %+v", windowed.Rows)
	}
	if got := windowed.Rows[1].Tools; len(got) != 1 || got[0] != "Claude" {
		t.Errorf("tools in window = %v, want [Claude]", got)
	}
	if windowed.Totals.Tokens != 350 || windowed.Totals.StoredCost != 2.0 {
		t.Errorf("totals = %+v", windowed.Totals)
	}

	// The merged project keeps both tools and both days across the full set.
	merged := all.Rows[2]
	if merged.Project != "/repo/a" || merged.Tokens != 150 || merged.ActiveDays != 2 {
		t.Errorf("merged project = %+v", merged)
	}
	// Contributing tools rank by the tokens they contributed.
	if len(merged.Tools) != 2 || merged.Tools[0] != "OpenCode" || merged.Tools[1] != "Claude" {
		t.Errorf("merged tools = %v, want OpenCode first with 100 tokens", merged.Tools)
	}
	if merged.ToolTokens["OpenCode"] != 100 || merged.ToolTokens["Claude"] != 50 {
		t.Errorf("tool token weights = %v", merged.ToolTokens)
	}
	if merged.FirstDay != "2026-09-10" || merged.LastDay != "2026-09-11" {
		t.Errorf("date range = %s..%s", merged.FirstDay, merged.LastDay)
	}
	if merged.Cost != 1.5 {
		t.Errorf("stored cost = %v, want 1.5", merged.Cost)
	}
}

func TestAggregateProjectsSortKeys(t *testing.T) {
	days := []ProjectDay{
		{Project: "/repo/big", Day: "2026-09-01", Tokens: 5000, Turns: 1, Cost: 1.0},
		{Project: "/repo/expensive", Day: "2026-09-16", Tokens: 100, Turns: 1, Cost: 90.0},
		{Project: "/repo/steady", Day: "2026-09-10", Tokens: 900, Turns: 1, Cost: 5.0},
	}
	// Give the steady project more active days than the others.
	days = append(days,
		ProjectDay{Project: "/repo/steady", Day: "2026-09-11", Tokens: 100, Turns: 1},
		ProjectDay{Project: "/repo/steady", Day: "2026-09-12", Tokens: 100, Turns: 1},
	)

	cost := AggregateProjects(days, ProjectOptions{Sort: "cost"}, nil)
	if cost.Rows[0].Project != "/repo/expensive" {
		t.Errorf("cost sort first = %s, want /repo/expensive", cost.Rows[0].Project)
	}

	days2 := AggregateProjects(days, ProjectOptions{Sort: "days"}, nil)
	if days2.Rows[0].Project != "/repo/steady" {
		t.Errorf("days sort first = %s, want /repo/steady", days2.Rows[0].Project)
	}

	recent := AggregateProjects(days, ProjectOptions{Sort: "recent"}, nil)
	if recent.Rows[0].Project != "/repo/expensive" {
		t.Errorf("recent sort first = %s, want the newest day", recent.Rows[0].Project)
	}

	tokens := AggregateProjects(days, ProjectOptions{}, nil)
	if tokens.Rows[0].Project != "/repo/big" {
		t.Errorf("default sort first = %s, want the largest token total", tokens.Rows[0].Project)
	}

	asc := AggregateProjects(days, ProjectOptions{Sort: "cost", Order: "asc"}, nil)
	if asc.Rows[0].Project != "/repo/big" {
		t.Errorf("ascending cost first = %s, want the cheapest", asc.Rows[0].Project)
	}
}

func TestAggregateModelsRankingAndWindow(t *testing.T) {
	batches := []ToolDays{
		{Tool: "OpenCode", Days: []DailyRow{
			{Day: "2026-09-10", Models: map[string]ModelTokens{
				"deepseek-flash": {Input: 1000},
				"glm-5.3":        {Input: 100},
			}},
			{Day: "2026-09-11", Models: map[string]ModelTokens{
				"deepseek-flash": {Input: 500},
			}},
		}},
		{Tool: "Codex", Days: []DailyRow{
			{Day: "2026-09-11", Models: map[string]ModelTokens{
				"gpt-6-astra": {Input: 300},
			}},
			{Day: "2026-08-01", Models: map[string]ModelTokens{
				"ancient-model": {Input: 99_999},
			}},
		}},
	}

	all := AggregateModels(batches, ModelOptions{}, nil)
	if len(all.Rows) != 4 {
		t.Fatalf("expected 4 models, got %d", len(all.Rows))
	}
	if all.Rows[0].Model != "ancient-model" {
		t.Errorf("default order should lead with the largest token total, got %s", all.Rows[0].Model)
	}

	windowed := AggregateModels(batches, ModelOptions{Since: "2026-09-10", Until: "2026-09-11"}, nil)
	if len(windowed.Rows) != 3 {
		t.Fatalf("expected 3 models in the window, got %d", len(windowed.Rows))
	}
	top := windowed.Rows[0]
	if top.Model != "deepseek-flash" || top.Tokens != 1500 {
		t.Errorf("top model = %+v, want deepseek-flash with 1500 tokens", top)
	}
	if top.Days != 2 {
		t.Errorf("active days = %d, want 2", top.Days)
	}
	if len(top.Tools) != 1 || top.Tools[0] != "OpenCode" {
		t.Errorf("tools = %v, want [OpenCode]", top.Tools)
	}
	if windowed.Totals.Tokens != 1900 {
		t.Errorf("totals = %+v", windowed.Totals)
	}

	last := AggregateModels(batches, ModelOptions{Last: 2, Now: time.Date(2026, 9, 11, 12, 0, 0, 0, time.Local)}, nil)
	if len(last.Rows) != 3 {
		t.Errorf("--last 2 should drop the August day, got %d rows", len(last.Rows))
	}
}

// ratePricer prices each model at its own per-million rate and reports the
// rest as unpriced, which is how the real catalog behaves.
type ratePricer struct{ rates map[string]float64 }

func (r ratePricer) PriceDay(day DailyRow) (float64, []string) {
	var cost float64
	var missing []string
	for model, counts := range day.Models {
		rate, ok := r.rates[model]
		if !ok {
			missing = append(missing, model)
			continue
		}
		cost += float64(counts.Total()) * rate / 1_000_000
	}
	return cost, missing
}

func TestAggregateModelsPricing(t *testing.T) {
	batches := []ToolDays{
		{Tool: "Codex", Days: []DailyRow{
			{Day: "2026-09-11", Models: map[string]ModelTokens{
				"gpt-5.4": {Input: 1_000_000},
				"mystery": {Input: 10},
			}},
		}},
	}
	pricer := ratePricer{rates: map[string]float64{"gpt-5.4": 2.5}}
	rep := AggregateModels(batches, ModelOptions{Sort: "cost"}, pricer)

	if rep.Rows[0].Model != "gpt-5.4" || rep.Rows[0].Cost != 2.5 {
		t.Errorf("cost sort should lead with the priced model, got %+v", rep.Rows[0])
	}
	if rep.Totals.Cost != 2.5 {
		t.Errorf("total cost = %v, want 2.5", rep.Totals.Cost)
	}
	if len(rep.Totals.MissingPricing) != 1 || rep.Totals.MissingPricing[0] != "mystery" {
		t.Errorf("missing pricing = %v", rep.Totals.MissingPricing)
	}
}

func TestAggregateProjectsLastAndPricing(t *testing.T) {
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.Local)
	days := []ProjectDay{
		{Project: "/repo/old", Day: "2026-08-01", Tool: "OpenCode", Tokens: 10, Turns: 1},
		{Project: "/repo/new", Day: "2026-09-15", Tokens: 20, Turns: 1, Models: map[string]ModelTokens{
			"gpt-5.4": {Input: 20},
		}},
		{Project: "/repo/stored", Day: "2026-09-15", Tokens: 30, Turns: 1, Cost: 4.0},
	}

	pricer := &stubPricer{cost: 0.5, missing: []string{"mystery"}}
	rep := AggregateProjects(days, ProjectOptions{Last: 7, Now: now}, pricer)

	if len(rep.Rows) != 2 {
		t.Fatalf("expected 2 projects in the last 7 days, got %d", len(rep.Rows))
	}
	// The old project is outside the window.
	for _, row := range rep.Rows {
		if row.Project == "/repo/old" {
			t.Errorf("old project should have been filtered: %+v", row)
		}
	}
	if rep.Totals.EstimatedCost != 0.5 {
		t.Errorf("estimated cost = %v, want 0.5", rep.Totals.EstimatedCost)
	}
	if rep.Totals.StoredCost != 4.0 {
		t.Errorf("stored cost = %v, want 4.0", rep.Totals.StoredCost)
	}
	if rep.Totals.Cost != 4.5 {
		t.Errorf("total cost = %v, want 4.5", rep.Totals.Cost)
	}
	if len(rep.Totals.MissingPricing) != 1 || rep.Totals.MissingPricing[0] != "mystery" {
		t.Errorf("missing pricing = %v", rep.Totals.MissingPricing)
	}
}

func TestProjectSlug(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"/mnt/Jad/github/lab/thermal-streak", "thermal-streak"},
		{"/mnt/Jad/github/lab/thermal-streak/", "thermal-streak"},
		{"/home/user/projects/waqftech/tree.waqf.app", "tree.waqf.app"},
		{"/home/user/My Project (v1)", "My-Project-v1"},
		{"   ", ""},
		{"/", "project"},
		{".", "project"},
		{"project-name", "project-name"},
		{"/var/log/--special--/", "special"},
	}

	for _, tc := range cases {
		got := ProjectSlug(tc.input)
		if got != tc.want {
			t.Errorf("ProjectSlug(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
