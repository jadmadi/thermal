package loaders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadJsonlData_LineOver64KB(t *testing.T) {
	dir := t.TempDir()
	historyFile := filepath.Join(dir, "history.jsonl")

	// Message 1: standard
	msg1 := `{"ts":1785000000}` + "\n"
	// Message 2: large line > 64KB (80KB)
	largeField := strings.Repeat("X", 80*1024)
	msg2 := fmt.Sprintf(`{"ts":1785000100,"extra":"%s"}`, largeField) + "\n"
	// Message 3: after large line
	msg3 := `{"ts":1785000200}` + "\n"

	if err := os.WriteFile(historyFile, []byte(msg1+msg2+msg3), 0644); err != nil {
		t.Fatalf("failed writing history: %v", err)
	}

	sum, _, _, err := loadJsonlData(dir, "ts", false)
	if err != nil {
		t.Fatalf("loadJsonlData error: %v", err)
	}

	if sum.LifetimeTokens != 3 {
		t.Fatalf("expected 3 commands, got %d", sum.LifetimeTokens)
	}
}
