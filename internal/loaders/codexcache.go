// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

const codexRolloutCacheVersion = 1

type rolloutResult struct {
	breakdown    *tokenBreakdown
	linesAdded   int64
	linesDeleted int64
	filesTouched int64
	warning      string
}

// CodexRolloutEntry represents the cached metrics for a single rollout JSONL file.
type CodexRolloutEntry struct {
	Size         int64  `json:"size"`
	ModTime      int64  `json:"modTime"`
	Input        int64  `json:"input"`
	Output       int64  `json:"output"`
	Reasoning    int64  `json:"reasoning"`
	Cache        int64  `json:"cache"`
	HasBreakdown bool   `json:"hasBreakdown"`
	LinesAdded   int64  `json:"linesAdded"`
	LinesDeleted int64  `json:"linesDeleted"`
	FilesTouched int64  `json:"filesTouched"`
	Warning      string `json:"warning,omitempty"`
	Offset       int64  `json:"offset,omitempty"`
}

// CodexRolloutCache is the persistent snapshot map for all rollouts associated with a state database.
type CodexRolloutCache struct {
	Version int                          `json:"version"`
	Entries map[string]CodexRolloutEntry `json:"entries"`
	mu      sync.RWMutex
}

func codexRolloutCachePath(dbPath string) (string, error) {
	home := thermal.HomeDir()
	dir := filepath.Join(home, ".cache", "thermal", "codex")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	canon := CanonicalDatabasePath(dbPath)
	h := sha256.Sum256([]byte(canon))
	fileName := fmt.Sprintf("rollouts_%s.json", hex.EncodeToString(h[:12]))
	return filepath.Join(dir, fileName), nil
}

func loadCodexRolloutCache(dbPath string) *CodexRolloutCache {
	c := &CodexRolloutCache{
		Version: codexRolloutCacheVersion,
		Entries: make(map[string]CodexRolloutEntry),
	}
	p, err := codexRolloutCachePath(dbPath)
	if err != nil {
		return c
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return c
	}
	var loaded struct {
		Version int                          `json:"version"`
		Entries map[string]CodexRolloutEntry `json:"entries"`
	}
	if err := json.Unmarshal(b, &loaded); err != nil || loaded.Version != codexRolloutCacheVersion || loaded.Entries == nil {
		return c
	}
	c.Entries = loaded.Entries
	return c
}

func saveCodexRolloutCache(dbPath string, cache *CodexRolloutCache) error {
	if cache == nil {
		return nil
	}
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	p, err := codexRolloutCachePath(dbPath)
	if err != nil {
		return err
	}

	payload := struct {
		Version int                          `json:"version"`
		Entries map[string]CodexRolloutEntry `json:"entries"`
	}{
		Version: codexRolloutCacheVersion,
		Entries: cache.Entries,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	tmp := fmt.Sprintf("%s.tmp.%d", p, time.Now().UnixNano())
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// scanOrReuseRollout evaluates whether a rollout file can be served from cache,
// appended incrementally, or fully re-scanned.
func scanOrReuseRollout(path string, cached *CodexRolloutEntry) (rolloutResult, CodexRolloutEntry, bool) {
	fi, err := os.Stat(path)
	if err != nil {
		return rolloutResult{}, CodexRolloutEntry{}, false
	}

	curSize := fi.Size()
	curMod := fi.ModTime().UnixNano()

	// 1. Exact Cache Hit: size and modTime match perfectly
	if cached != nil && cached.Size == curSize && cached.ModTime == curMod {
		var b *tokenBreakdown
		if cached.HasBreakdown {
			b = &tokenBreakdown{
				input:     cached.Input,
				output:    cached.Output,
				reasoning: cached.Reasoning,
				cache:     cached.Cache,
			}
		}
		return rolloutResult{
			breakdown:    b,
			linesAdded:   cached.LinesAdded,
			linesDeleted: cached.LinesDeleted,
			filesTouched: cached.FilesTouched,
			warning:      cached.Warning,
		}, *cached, false
	}

	// 2. Append Delta: file grew strictly forward and previous offset is valid
	if cached != nil && curSize > cached.Size && curMod >= cached.ModTime && cached.Offset > 0 && cached.Offset <= curSize {
		deltaRes, newEntry, ok := scanRolloutDelta(path, *cached, curSize, curMod)
		if ok {
			return deltaRes, newEntry, true
		}
		// Delta failed or encountered corruption; fall through to full scan
	}

	// 3. Full Cold Scan
	b, la, ld, ft, w, offset := scanRolloutFull(path)
	entry := CodexRolloutEntry{
		Size:         curSize,
		ModTime:      curMod,
		LinesAdded:   la,
		LinesDeleted: ld,
		FilesTouched: ft,
		Warning:      w,
		Offset:       offset,
	}
	if b != nil {
		entry.HasBreakdown = true
		entry.Input = b.input
		entry.Output = b.output
		entry.Reasoning = b.reasoning
		entry.Cache = b.cache
	}

	return rolloutResult{
		breakdown:    b,
		linesAdded:   la,
		linesDeleted: ld,
		filesTouched: ft,
		warning:      w,
	}, entry, true
}

func scanRolloutDelta(path string, prev CodexRolloutEntry, curSize, curMod int64) (rolloutResult, CodexRolloutEntry, bool) {
	f, err := os.Open(path)
	if err != nil {
		return rolloutResult{}, prev, false
	}
	defer f.Close()

	if _, err := f.Seek(prev.Offset, io.SeekStart); err != nil {
		return rolloutResult{}, prev, false
	}

	scanner := newJSONLScanner(f)
	var last *tokenBreakdown
	if prev.HasBreakdown {
		last = &tokenBreakdown{
			input:     prev.Input,
			output:    prev.Output,
			reasoning: prev.Reasoning,
			cache:     prev.Cache,
		}
	}

	linesAdded := prev.LinesAdded
	linesDeleted := prev.LinesDeleted
	filesTouched := prev.FilesTouched

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, `"FileChange"`) {
			parseFileChange(line, &linesAdded, &linesDeleted, &filesTouched)
			continue
		}

		if strings.Contains(line, `"token_count"`) {
			if tb := parseTokenCount(line); tb != nil {
				last = tb
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// On scan error during delta, trigger fallback to full scan
		return rolloutResult{}, prev, false
	}

	curOffset, _ := f.Seek(0, io.SeekCurrent)
	if curOffset < prev.Offset {
		curOffset = curSize
	}

	newEntry := CodexRolloutEntry{
		Size:         curSize,
		ModTime:      curMod,
		LinesAdded:   linesAdded,
		LinesDeleted: linesDeleted,
		FilesTouched: filesTouched,
		Offset:       curOffset,
	}
	if last != nil {
		newEntry.HasBreakdown = true
		newEntry.Input = last.input
		newEntry.Output = last.output
		newEntry.Reasoning = last.reasoning
		newEntry.Cache = last.cache
	}

	return rolloutResult{
		breakdown:    last,
		linesAdded:   linesAdded,
		linesDeleted: linesDeleted,
		filesTouched: filesTouched,
	}, newEntry, true
}

func scanRolloutFull(path string) (*tokenBreakdown, int64, int64, int64, string, int64) {
	if _, err := os.Stat(path); err != nil {
		return nil, 0, 0, 0, "", 0
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, 0, formatScanWarning(path, err), 0
	}
	defer f.Close()

	scanner := newJSONLScanner(f)
	var last *tokenBreakdown
	var linesAdded, linesDeleted, filesTouched int64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, `"FileChange"`) {
			parseFileChange(line, &linesAdded, &linesDeleted, &filesTouched)
			continue
		}

		if strings.Contains(line, `"token_count"`) {
			if tb := parseTokenCount(line); tb != nil {
				last = tb
			}
		}
	}

	var warning string
	if err := scanner.Err(); err != nil {
		warning = formatScanWarning(path, err)
	}

	curOffset, _ := f.Seek(0, io.SeekCurrent)
	fi, _ := f.Stat()
	if curOffset == 0 && fi != nil {
		curOffset = fi.Size()
	}

	return last, linesAdded, linesDeleted, filesTouched, warning, curOffset
}

func parseFileChange(line string, linesAdded, linesDeleted, filesTouched *int64) {
	var rec struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if json.Unmarshal([]byte(line), &rec) == nil && rec.Type == "event_msg" {
		var ev struct {
			Item *struct {
				Type    string `json:"type"`
				Changes map[string]struct {
					Type        string `json:"type"`
					UnifiedDiff string `json:"unified_diff"`
					Content     string `json:"content"`
				} `json:"changes"`
			} `json:"item"`
		}
		if json.Unmarshal(rec.Payload, &ev) == nil && ev.Item != nil && ev.Item.Type == "FileChange" && ev.Item.Changes != nil {
			*filesTouched += int64(len(ev.Item.Changes))
			for _, ch := range ev.Item.Changes {
				if ch.UnifiedDiff != "" {
					add, del := thermal.ParseDiffStats(ch.UnifiedDiff)
					*linesAdded += add
					*linesDeleted += del
				} else if ch.Content != "" && ch.Type == "add" {
					*linesAdded += thermal.CountLines(ch.Content)
				}
			}
		}
	}
}

func parseTokenCount(line string) *tokenBreakdown {
	var rec struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if json.Unmarshal([]byte(line), &rec) != nil || rec.Type != "event_msg" {
		return nil
	}

	var ev struct {
		Type string `json:"type"`
		Info *struct {
			TotalTokenUsage *struct {
				InputTokens           int64 `json:"input_tokens"`
				CachedInputTokens     int64 `json:"cached_input_tokens"`
				OutputTokens          int64 `json:"output_tokens"`
				ReasoningOutputTokens int64 `json:"reasoning_output_tokens"`
			} `json:"total_token_usage"`
		} `json:"info"`
	}
	if json.Unmarshal(rec.Payload, &ev) != nil || ev.Type != "token_count" || ev.Info == nil || ev.Info.TotalTokenUsage == nil {
		return nil
	}

	tu := ev.Info.TotalTokenUsage
	return &tokenBreakdown{
		input:     nonNegative(tu.InputTokens - tu.CachedInputTokens),
		output:    nonNegative(tu.OutputTokens - tu.ReasoningOutputTokens),
		reasoning: tu.ReasoningOutputTokens,
		cache:     tu.CachedInputTokens,
	}
}
