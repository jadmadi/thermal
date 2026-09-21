package pricing

import (
	"testing"
)

func TestStandardPlans(t *testing.T) {
	plans := StandardPlans()
	if len(plans) == 0 {
		t.Fatalf("expected non-empty standard plans")
	}
	defaults := DefaultComparisonPlans()
	if len(defaults) == 0 {
		t.Fatalf("expected non-empty default comparison plans")
	}
	for _, id := range defaults {
		plan, ok := LookupPlan(id)
		if !ok {
			t.Errorf("default plan %q not found in LookupPlan", id)
		}
		if plan.ID != id {
			t.Errorf("expected plan ID %q, got %q", id, plan.ID)
		}
	}
}

func TestLookupPlan_PrefixAndCase(t *testing.T) {
	tests := []struct {
		query  string
		wantID string
		wantOK bool
	}{
		{"claude-pro", "claude-pro", true},
		{"CLAUDE-PRO", "claude-pro", true},
		{"claude-max", "claude-max", true},
		{"deepseek", "deepseek-api", true},
		{"deepseek-api", "deepseek-api", true},
		{"cursor", "cursor-pro", true},
		{"non-existent-plan", "", false},
	}

	for _, tt := range tests {
		plan, ok := LookupPlan(tt.query)
		if ok != tt.wantOK {
			t.Errorf("LookupPlan(%q) ok = %v, want %v", tt.query, ok, tt.wantOK)
		}
		if ok && plan.ID != tt.wantID {
			t.Errorf("LookupPlan(%q) ID = %q, want %q", tt.query, plan.ID, tt.wantID)
		}
	}
}
