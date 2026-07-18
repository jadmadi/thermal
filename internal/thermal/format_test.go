package thermal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompactNumber(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1540, "1.5K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{44000000, "44.0M"},
		{1000000000, "1.0B"},
		{1850000000, "1.9B"},
	}
	for _, tc := range tests {
		if got := CompactNumber(tc.input); got != tc.want {
			t.Errorf("CompactNumber(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestPadding(t *testing.T) {
	if got := PadRight("abc", 5); got != "abc  " {
		t.Errorf("PadRight = %q, want %q", got, "abc  ")
	}
	if got := PadRight("abcde", 3); got != "abcde" {
		t.Errorf("PadRight truncation = %q, want %q", got, "abcde")
	}
	if got := PadLeft("abc", 5); got != "  abc" {
		t.Errorf("PadLeft = %q, want %q", got, "  abc")
	}
}

func TestFormatPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	sub := filepath.Join(home, ".local", "share", "opencode")
	got := FormatPath(sub)
	if got != "~/.local/share/opencode" && got != "~\\.local\\share\\opencode" {
		t.Errorf("FormatPath(%q) = %q", sub, got)
	}
}
