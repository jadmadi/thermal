package loaders

import (
	"testing"
	"github.com/jadmadi/thermal/internal/thermal"
)

func TestResolveTool_Aliases(t *testing.T) {
	tests := []struct {
		alias string
		want  thermal.Tool
		ok    bool
	}{
		{"mimo", thermal.ToolMiMoCode, true},
		{"oc", thermal.ToolOpenCode, true},
		{"cmd", thermal.ToolCommandCode, true},
		{"whale", thermal.ToolCodewhale, true},
		{"nonexistent", "", false},
	}
	for _, tc := range tests {
		got, ok := ResolveTool(tc.alias)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ResolveTool(%q) = (%q, %v), want (%q, %v)", tc.alias, got, ok, tc.want, tc.ok)
		}
	}
}

func TestAllTools_Config(t *testing.T) {
	tools := AllTools()
	if len(tools) == 0 {
		t.Fatalf("expected tools in registry")
	}
	for id, info := range tools {
		if info.Name == "" || info.Loader == nil {
			t.Errorf("tool %s has incomplete ToolInfo configuration", id)
		}
	}
}
