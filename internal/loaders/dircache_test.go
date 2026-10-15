// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestDirectoryScanSignature(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "a.jsonl")
	if err := os.WriteFile(file, []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	base, err := directoryScanSignature(root)
	if err != nil {
		t.Fatal(err)
	}
	again, err := directoryScanSignature(root)
	if err != nil {
		t.Fatal(err)
	}
	if base != again {
		t.Fatalf("signature not stable: %s vs %s", base, again)
	}

	// Append a file.
	if err := os.WriteFile(filepath.Join(root, "b.jsonl"), []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	added, _ := directoryScanSignature(root)
	if added == base {
		t.Fatal("adding a file did not change the signature")
	}

	// Rewrite with different size.
	if err := os.WriteFile(file, []byte("one changed and longer\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rewritten, _ := directoryScanSignature(root)
	if rewritten == added {
		t.Fatal("rewriting a file did not change the signature")
	}

	// Truncate a file.
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	truncated, _ := directoryScanSignature(root)
	if truncated == rewritten {
		t.Fatal("truncating a file did not change the signature")
	}

	// Delete a file.
	if err := os.Remove(filepath.Join(root, "b.jsonl")); err != nil {
		t.Fatal(err)
	}
	deleted, _ := directoryScanSignature(root)
	if deleted == truncated {
		t.Fatal("deleting a file did not change the signature")
	}

	// Documented boundary: a same-size rewrite that also restores the exact
	// modification time is outside what a stat-based signature can observe.
	boundaryFile := filepath.Join(root, "c.jsonl")
	if err := os.WriteFile(boundaryFile, []byte("AAAA"), 0o644); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(boundaryFile)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := directoryScanSignature(root)
	if err := os.WriteFile(boundaryFile, []byte("BBBB"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(boundaryFile, fi.ModTime(), fi.ModTime()); err != nil {
		t.Fatal(err)
	}
	after, _ := directoryScanSignature(root)
	if before != after {
		t.Fatalf("same-size same-mtime rewrite unexpectedly changed the signature")
	}
}

func TestDirScanCacheGenericWrapper(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("THERMAL_CONFIG", "")

	dataDir := filepath.Join(home, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "one.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	var (
		calls    int64
		failNext int64
	)
	info := ToolInfo{
		DataDir: dataDir,
		Name:    "TestTool",
		Loader: func(dir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
			atomic.AddInt64(&calls, 1)
			if atomic.LoadInt64(&failNext) == 1 {
				return thermal.Summary{}, nil, nil, os.ErrPermission
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				return thermal.Summary{}, nil, nil, err
			}
			files := int64(len(entries))
			return thermal.Summary{
					Sessions:       int(files),
					LifetimeTokens: files * 100,
					Warnings:       []string{"scan warning"},
				},
				[]thermal.DailyRow{{Day: "2026-10-15", Tokens: files * 100, Turns: int(files)}},
				[]thermal.ProjectDay{{Project: "p", Day: "2026-10-15", Tokens: files * 100}},
				nil
		},
	}
	tool := thermal.Tool("testtool")

	first, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("first load ran the loader %d times, want 1", got)
	}
	if first.Summary.LifetimeTokens != 100 {
		t.Fatalf("first load tokens = %d, want 100", first.Summary.LifetimeTokens)
	}

	// An unchanged tree must be served from the snapshot without running the
	// loader again, and warnings must survive the round trip.
	second, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("unchanged tree re-ran the loader: calls = %d, want 1", got)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("cached result differs:\n%+v\n%+v", first, second)
	}
	if len(second.Summary.Warnings) != 1 || second.Summary.Warnings[0] != "scan warning" {
		t.Fatalf("warnings lost through the cache: %+v", second.Summary.Warnings)
	}

	// A changed tree forces a full rescan and updates the snapshot.
	if err := os.WriteFile(filepath.Join(dataDir, "two.txt"), []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}
	third, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("third load: %v", err)
	}
	if got := atomic.LoadInt64(&calls); got != 2 {
		t.Fatalf("changed tree did not rescan: calls = %d, want 2", got)
	}
	if third.Summary.LifetimeTokens != 200 {
		t.Fatalf("third load tokens = %d, want 200", third.Summary.LifetimeTokens)
	}

	// A corrupt cache file falls back to a full scan instead of failing.
	cacheFiles, err := filepath.Glob(filepath.Join(home, ".cache", "thermal", "testtool", "dirscan_*.json"))
	if err != nil || len(cacheFiles) != 1 {
		t.Fatalf("expected exactly one cache file, got %v (err=%v)", cacheFiles, err)
	}
	if err := os.WriteFile(cacheFiles[0], []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	fourth, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("fourth load: %v", err)
	}
	if got := atomic.LoadInt64(&calls); got != 3 {
		t.Fatalf("corrupt cache did not rescan: calls = %d, want 3", got)
	}
	if fourth.Summary.LifetimeTokens != 200 {
		t.Fatalf("fourth load tokens = %d, want 200", fourth.Summary.LifetimeTokens)
	}

	// A loader error during a cache miss is surfaced and never cached: the next
	// call with the same tree rescans instead of serving a poisoned snapshot.
	if err := os.WriteFile(filepath.Join(dataDir, "three.txt"), []byte("third"), 0o644); err != nil {
		t.Fatal(err)
	}
	atomic.StoreInt64(&failNext, 1)
	if _, err := LoadToolData(tool, info, ""); err == nil {
		t.Fatal("loader error was swallowed")
	}
	atomic.StoreInt64(&failNext, 0)
	fifth, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("fifth load: %v", err)
	}
	if got := atomic.LoadInt64(&calls); got != 5 {
		t.Fatalf("retry ran the loader %d times total, want 5", got)
	}
	if fifth.Summary.LifetimeTokens != 300 {
		t.Fatalf("fifth load tokens = %d, want 300", fifth.Summary.LifetimeTokens)
	}
}

func writeClaudeSession(t *testing.T, dir, name string, tokens int64, ts string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	line := fmt.Sprintf(`{"type":"assistant","timestamp":%q,"cwd":"/tmp/dircache-repo","message":{"id":%q,"model":"claude-3-5-sonnet","usage":{"input_tokens":%d,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`+"\n",
		ts, name+"-m1", tokens)
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDirScanCacheClaudeIntegration(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("THERMAL_CONFIG", "")

	projectDir := filepath.Join(home, ".claude", "projects", "proj-a")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeClaudeSession(t, projectDir, "s1.jsonl", 1000, "2026-10-15T10:00:00Z")
	writeClaudeSession(t, projectDir, "s2.jsonl", 2000, "2026-10-15T11:00:00Z")

	info := AllTools()[thermal.ToolClaude]
	tool := thermal.ToolClaude

	full, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if full.Summary.LifetimeTokens != 3000 {
		t.Fatalf("lifetime tokens = %d, want 3000", full.Summary.LifetimeTokens)
	}

	// Cached result must be byte-identical to a forced full scan.
	cacheFiles, _ := filepath.Glob(filepath.Join(home, ".cache", "thermal", "claude", "dirscan_*.json"))
	if len(cacheFiles) != 1 {
		t.Fatalf("expected one claude cache file, got %v", cacheFiles)
	}
	if err := os.Remove(cacheFiles[0]); err != nil {
		t.Fatal(err)
	}
	forced, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("forced load: %v", err)
	}
	if !reflect.DeepEqual(full, forced) {
		t.Fatalf("cached result differs from a full scan:\n%+v\n%+v", full, forced)
	}

	// Append a session: totals grow and the cached snapshot is refreshed.
	writeClaudeSession(t, projectDir, "s3.jsonl", 1500, "2026-10-15T12:00:00Z")
	appended, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("append load: %v", err)
	}
	if appended.Summary.LifetimeTokens != 4500 {
		t.Fatalf("after append tokens = %d, want 4500", appended.Summary.LifetimeTokens)
	}

	// Rewrite a session with a different size: totals follow the rewrite.
	rewrittenLine := `{"type":"assistant","timestamp":"2026-10-15T11:00:00Z","cwd":"/tmp/dircache-repo","message":{"id":"s2-rewrite","model":"claude-3-5-sonnet","usage":{"input_tokens":5000,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}` + "\n"
	if err := os.WriteFile(filepath.Join(projectDir, "s2.jsonl"), []byte(rewrittenLine), 0o644); err != nil {
		t.Fatal(err)
	}
	rewritten, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("rewrite load: %v", err)
	}
	if rewritten.Summary.LifetimeTokens != 7500 {
		t.Fatalf("after rewrite tokens = %d, want 7500", rewritten.Summary.LifetimeTokens)
	}

	// Truncate a session to zero records: its tokens drop out.
	if err := os.WriteFile(filepath.Join(projectDir, "s3.jsonl"), []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	truncated, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("truncate load: %v", err)
	}
	if truncated.Summary.LifetimeTokens != 6000 {
		t.Fatalf("after truncate tokens = %d, want 6000", truncated.Summary.LifetimeTokens)
	}

	// Delete a session file: its tokens drop out.
	if err := os.Remove(filepath.Join(projectDir, "s1.jsonl")); err != nil {
		t.Fatal(err)
	}
	deleted, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("delete load: %v", err)
	}
	if deleted.Summary.LifetimeTokens != 5000 {
		t.Fatalf("after delete tokens = %d, want 5000", deleted.Summary.LifetimeTokens)
	}
	if deleted.Summary.Sessions != 1 {
		t.Fatalf("after delete sessions = %d, want 1", deleted.Summary.Sessions)
	}

	// Determinism: the cached view of that final state matches a forced scan.
	last, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("final cached load: %v", err)
	}
	cacheFiles, _ = filepath.Glob(filepath.Join(home, ".cache", "thermal", "claude", "dirscan_*.json"))
	if len(cacheFiles) != 1 {
		t.Fatalf("expected one claude cache file after mutations, got %v", cacheFiles)
	}
	if err := os.Remove(cacheFiles[0]); err != nil {
		t.Fatal(err)
	}
	lastForced, err := LoadToolData(tool, info, "")
	if err != nil {
		t.Fatalf("final forced load: %v", err)
	}
	if !reflect.DeepEqual(last, lastForced) {
		t.Fatalf("final cached view differs from a full scan:\n%+v\n%+v", last, lastForced)
	}

	// Warm reads must not regress into slow paths: a cached read is expected to
	// stay well under the full-scan cost on the same tree.
	start := time.Now()
	if _, err := LoadToolData(tool, info, ""); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Fatalf("cached read took %v", elapsed)
	}
}
