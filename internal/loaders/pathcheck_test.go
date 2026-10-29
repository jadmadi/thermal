// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func sqliteHeaderFixture(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "fixture.db")
	data := append([]byte("SQLite format 3\x00"), make([]byte, 256)...)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write sqlite fixture: %v", err)
	}
	return p
}

func TestToolPathShape(t *testing.T) {
	cases := map[thermal.Tool]PathShape{
		thermal.ToolOpenCode:    ShapeSQLiteFile,
		thermal.ToolMiMoCode:    ShapeSQLiteFile,
		thermal.ToolDevin:       ShapeSQLiteFile,
		thermal.ToolMuse:        ShapeEither,
		thermal.ToolClaude:      ShapeDirectory,
		thermal.ToolCommandCode: ShapeDirectory,
		thermal.ToolCodex:       ShapeDirectory,
	}
	for tool, want := range cases {
		if got := ToolPathShape(tool); got != want {
			t.Errorf("ToolPathShape(%s)=%v, want %v", tool, got, want)
		}
	}
}

func TestValidateDBPathMissing(t *testing.T) {
	_, err := ValidateDBPath(thermal.ToolOpenCode, filepath.Join(t.TempDir(), "missing.db"))
	if err == nil {
		t.Fatal("expected error")
	}
	var dbErr *DBPathError
	if !errors.As(err, &dbErr) {
		t.Fatalf("expected DBPathError, got %T", err)
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestValidateDBPathSQLiteFile(t *testing.T) {
	p := sqliteHeaderFixture(t)
	got, err := ValidateDBPath(thermal.ToolOpenCode, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != p {
		t.Fatalf("got %q, want %q", got, p)
	}
}

func TestValidateDBPathNonSQLite(t *testing.T) {
	p := filepath.Join(t.TempDir(), "not.db")
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ValidateDBPath(thermal.ToolOpenCode, p)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not a SQLite database") {
		t.Fatalf("unexpected message: %v", err)
	}
	if strings.Contains(err.Error(), "SQLite header") && !strings.Contains(err.Error(), "omit --db") {
		t.Fatalf("raw header detail leaked: %v", err)
	}
}

func TestValidateDBPathEmptyFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.db")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ValidateDBPath(thermal.ToolOpenCode, p)
	if err == nil || !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDBPathDirectoryForFileTool(t *testing.T) {
	dir := t.TempDir()
	_, err := ValidateDBPath(thermal.ToolOpenCode, dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expects opencode.db") {
		t.Fatalf("unexpected message: %v", err)
	}
	if !strings.Contains(err.Error(), "pass the database file directly") {
		t.Fatalf("missing fix: %v", err)
	}
}

func TestValidateDBPathDirectoryWithDB(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "opencode.db")
	if err := os.WriteFile(p, append([]byte("SQLite format 3\x00"), make([]byte, 64)...), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ValidateDBPath(thermal.ToolOpenCode, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != p {
		t.Fatalf("got %q, want %q", got, p)
	}
}

func TestValidateDBPathFileForDirectoryTool(t *testing.T) {
	p := sqliteHeaderFixture(t)
	_, err := ValidateDBPath(thermal.ToolClaude, p)
	if err == nil || !strings.Contains(err.Error(), "expects a directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDBPathDirectoryTool(t *testing.T) {
	dir := t.TempDir()
	got, err := ValidateDBPath(thermal.ToolClaude, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dir {
		t.Fatalf("got %q, want %q", got, dir)
	}
}

func TestValidateDBPathMuseEither(t *testing.T) {
	file := sqliteHeaderFixture(t)
	if _, err := ValidateDBPath(thermal.ToolMuse, file); err != nil {
		t.Fatalf("muse file: %v", err)
	}
	dir := t.TempDir()
	if _, err := ValidateDBPath(thermal.ToolMuse, dir); err != nil {
		t.Fatalf("muse dir: %v", err)
	}
}

func TestTranslateDBError(t *testing.T) {
	err := translateDBError(thermal.ToolOpenCode, "/tmp/db", errors.New("file is not a database (26)"))
	var dbErr *DBPathError
	if !errors.As(err, &dbErr) {
		t.Fatalf("expected DBPathError, got %T", err)
	}
	if !strings.Contains(err.Error(), "not a SQLite database") {
		t.Fatalf("unexpected message: %v", err)
	}
	if strings.Contains(err.Error(), "(26)") {
		t.Fatalf("driver code leaked: %v", err)
	}
	if dbErr.Err == nil || !strings.Contains(dbErr.Err.Error(), "(26)") {
		t.Fatalf("raw error not preserved: %v", dbErr.Err)
	}

	dir := t.TempDir()
	err = translateDBError(thermal.ToolOpenCode, dir, errors.New("disk I/O error (10)"))
	if !strings.Contains(err.Error(), "is a directory") || !strings.Contains(err.Error(), "opencode.db") {
		t.Fatalf("unexpected disk error message: %v", err)
	}
}

func TestLoadToolDataNoFallback(t *testing.T) {
	called := false
	info := ToolInfo{
		DBPath: sqliteHeaderFixture(t),
		Name:   "OpenCode",
		Loader: func(string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
			called = true
			return thermal.Summary{}, nil, nil, nil
		},
	}
	_, err := LoadToolData(thermal.ToolOpenCode, info, filepath.Join(t.TempDir(), "missing.db"))
	if err == nil {
		t.Fatal("expected error")
	}
	if called {
		t.Fatal("loader was called with the default path after a bad --db")
	}
	var dbErr *DBPathError
	if !errors.As(err, &dbErr) {
		t.Fatalf("expected DBPathError, got %T", err)
	}
}

func TestLoadToolDataTranslatesLoaderError(t *testing.T) {
	file := sqliteHeaderFixture(t)
	info := ToolInfo{
		Name: "OpenCode",
		Loader: func(string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
			return thermal.Summary{}, nil, nil, fmt.Errorf("cannot acquire sqlite connection: file is not a database (26)")
		},
	}
	_, err := LoadToolData(thermal.ToolOpenCode, info, file)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not a SQLite database") || strings.Contains(err.Error(), "(26)") {
		t.Fatalf("unexpected translated error: %v", err)
	}
}

func TestLoadToolDataDirectoryOverrideUsesExpectedFile(t *testing.T) {
	dir := t.TempDir()
	expected := filepath.Join(dir, "opencode.db")
	if err := os.WriteFile(expected, append([]byte("SQLite format 3\x00"), make([]byte, 64)...), 0o644); err != nil {
		t.Fatal(err)
	}
	var gotPath string
	info := ToolInfo{
		Name: "OpenCode",
		Loader: func(path string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
			gotPath = path
			return thermal.Summary{}, nil, nil, nil
		},
	}
	if _, err := LoadToolData(thermal.ToolOpenCode, info, dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != expected {
		t.Fatalf("loader path %q, want %q", gotPath, expected)
	}
}

func TestConfigOverrideMissingExpectedFileWarns(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("THERMAL_CONFIG", "")
	cfgDir := filepath.Join(home, ".config", "thermal")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	custom := filepath.Join(home, "custom-opencode")
	if err := os.MkdirAll(custom, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "[paths]\nopencode = " + custom + "\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "thermal.config"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	info := AllTools()[thermal.ToolOpenCode]
	if !strings.Contains(info.Warning, "has no opencode.db") {
		t.Fatalf("expected missing-file warning, got %q", info.Warning)
	}
}
