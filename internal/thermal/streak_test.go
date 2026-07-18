package thermal

import (
	"testing"
)

func TestComputeStreaks_Empty(t *testing.T) {
	cur, long := ComputeStreaks(map[string]bool{})
	if cur != 0 || long != 0 {
		t.Fatalf("expected (0,0), got (%d,%d)", cur, long)
	}
}

func TestComputeStreaks_SingleDay(t *testing.T) {
	cur, long := ComputeStreaks(map[string]bool{"2026-03-15": true})
	if cur != 1 || long != 1 {
		t.Fatalf("expected (1,1), got (%d,%d)", cur, long)
	}
}

func TestComputeStreaks_ConsecutiveAcrossDST(t *testing.T) {
	// Test days spanning US/Europe spring forward and fall back transitions
	days := map[string]bool{
		"2026-03-07": true,
		"2026-03-08": true, // Spring forward
		"2026-03-09": true,
		"2026-10-31": true,
		"2026-11-01": true, // Fall back
		"2026-11-02": true,
	}
	cur, long := ComputeStreaks(days)
	if long != 3 {
		t.Fatalf("expected longest streak 3 across DST transition, got %d", long)
	}
	_ = cur
}

func TestActivityThresholds_DistinctEnforcement(t *testing.T) {
	// 4 days with 10 tokens, 1 day with 1000 tokens
	act := map[string]DayActivity{
		"2026-01-01": {Tokens: 10},
		"2026-01-02": {Tokens: 10},
		"2026-01-03": {Tokens: 10},
		"2026-01-04": {Tokens: 10},
		"2026-01-05": {Tokens: 1000},
	}
	thresh := ActivityThresholds(act)
	if thresh[1] <= thresh[0] || thresh[2] <= thresh[1] {
		t.Fatalf("expected strictly ascending thresholds, got %v", thresh)
	}
}

func TestActivityLevel(t *testing.T) {
	thresh := [3]int64{100, 500, 1000}
	tests := []struct {
		tokens int64
		want   int
	}{
		{0, 1},
		{50, 1},
		{100, 1},
		{101, 2},
		{500, 2},
		{501, 3},
		{1000, 3},
		{1001, 4},
	}
	for _, tc := range tests {
		got := ActivityLevel(DayActivity{Tokens: tc.tokens}, thresh)
		if got != tc.want {
			t.Errorf("ActivityLevel(%d, %v) = %d, want %d", tc.tokens, thresh, got, tc.want)
		}
	}
}
