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
	if len(merged.Tools) != 2 || merged.Tools[0] != "Claude" || merged.Tools[1] != "OpenCode" {
		t.Errorf("merged tools = %v", merged.Tools)
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
