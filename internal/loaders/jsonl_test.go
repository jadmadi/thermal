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
	if len(sum.Warnings) != 0 {
		t.Fatalf("expected 0 warnings on clean read, got %v", sum.Warnings)
	}
}

func TestLoadJsonlData_CeilingWarning(t *testing.T) {
	origCeiling := jsonlMaxLineBytes
	jsonlMaxLineBytes = 512
	defer func() { jsonlMaxLineBytes = origCeiling }()

	dir := t.TempDir()
	historyFile := filepath.Join(dir, "history.jsonl")

	msg1 := `{"ts":1785000000}` + "\n"
	msg2 := fmt.Sprintf(`{"ts":1785000100,"extra":"%s"}`, strings.Repeat("A", 600)) + "\n"
	msg3 := `{"ts":1785000200}` + "\n"

	if err := os.WriteFile(historyFile, []byte(msg1+msg2+msg3), 0644); err != nil {
		t.Fatalf("failed writing history: %v", err)
	}

	sum, _, _, err := loadJsonlData(dir, "ts", false)
	if err != nil {
		t.Fatalf("loadJsonlData error: %v", err)
	}

	if sum.LifetimeTokens != 1 {
		t.Fatalf("expected 1 command (pre-error data kept), got %d", sum.LifetimeTokens)
	}
	if len(sum.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(sum.Warnings), sum.Warnings)
	}
	w := sum.Warnings[0]
	if !strings.Contains(w, historyFile) || !strings.Contains(w, "512 byte ceiling") || !strings.Contains(w, "rest of file skipped") {
		t.Errorf("warning did not match expected format, got: %s", w)
	}
}

func TestLoadJsonlData_CleanFileNoWarnings(t *testing.T) {
	dir := t.TempDir()
	historyFile := filepath.Join(dir, "history.jsonl")

	msg1 := `{"ts":1785000000}` + "\n"
	msg2 := `{"ts":1785000100}` + "\n"

	if err := os.WriteFile(historyFile, []byte(msg1+msg2), 0644); err != nil {
		t.Fatalf("failed writing history: %v", err)
	}

	sum, _, _, err := loadJsonlData(dir, "ts", false)
	if err != nil {
		t.Fatalf("loadJsonlData error: %v", err)
	}

	if len(sum.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %v", sum.Warnings)
	}
}
