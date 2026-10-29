// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/jadmadi/thermal/internal/thermal"
)

// dirScanCacheVersion bumps whenever the cached snapshot shape changes, so an
// old cache file is ignored instead of decoded into stale zero fields.
const dirScanCacheVersion = 3

// DirScanCache is a disk-backed snapshot of a directory-backed loader scan.
// Invalidation is keyed on the canonical root plus a per-file tree signature
// (relative path, size, nanosecond modification time), so an untouched tree is
// reused without re-reading file bodies while any file addition, removal, or
// change forces a full rescan. As with the SQLite snapshot caches, a rewrite
// that preserves both byte size and the recorded modification time of a file
// is outside what a stat-based signature can observe.
type DirScanCache struct {
	Version  int                  `json:"version"`
	Tool     string               `json:"tool"`
	Root     string               `json:"root"`
	Sig      string               `json:"sig"`
	Warnings []string             `json:"warnings,omitempty"`
	Summary  thermal.Summary      `json:"summary"`
	Daily    []thermal.DailyRow   `json:"daily"`
	Projects []thermal.ProjectDay `json:"projects,omitempty"`
	Path     string               `json:"path"`
}

// directoryScanSignature hashes the stat of every regular file under root:
// relative path, byte size, and nanosecond modification time, sorted for a
// stable order. It never reads file bodies.
func directoryScanSignature(root string) (string, error) {
	var entries []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		entries = append(entries, fmt.Sprintf("%s\x00%d\x00%d", rel, info.Size(), info.ModTime().UnixNano()))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(entries)
	h := sha256.New()
	for _, e := range entries {
		h.Write([]byte(e))
		h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func dirScanCachePath(t thermal.Tool, canonicalRoot string) (string, error) {
	home := thermal.HomeDir()
	dir := filepath.Join(home, ".cache", "thermal", string(t))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(canonicalRoot))
	return filepath.Join(dir, fmt.Sprintf("dirscan_%s.json", hex.EncodeToString(h[:12]))), nil
}

// loadDirScanCache returns cached data only when the tool, canonical root, and
// tree signature all match. Any mismatch, decode failure, or version drift
// falls back to a full scan.
func loadDirScanCache(t thermal.Tool, canonicalRoot, sig, dataPath string) (ToolData, bool) {
	p, err := dirScanCachePath(t, canonicalRoot)
	if err != nil {
		return ToolData{}, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ToolData{}, false
	}
	var c DirScanCache
	if err := json.Unmarshal(b, &c); err != nil {
		return ToolData{}, false
	}
	if c.Version != dirScanCacheVersion || c.Tool != string(t) || c.Root != canonicalRoot || c.Sig != sig {
		return ToolData{}, false
	}
	summary := c.Summary
	summary.Warnings = c.Warnings
	return ToolData{Summary: summary, Daily: c.Daily, Projects: c.Projects, Path: dataPath}, true
}

func saveDirScanCache(t thermal.Tool, canonicalRoot, sig string, data ToolData) {
	p, err := dirScanCachePath(t, canonicalRoot)
	if err != nil {
		return
	}
	summary := data.Summary
	summary.Warnings = nil // persisted separately because Summary hides them from JSON
	c := DirScanCache{
		Version:  dirScanCacheVersion,
		Tool:     string(t),
		Root:     canonicalRoot,
		Sig:      sig,
		Warnings: data.Summary.Warnings,
		Summary:  summary,
		Daily:    data.Daily,
		Projects: data.Projects,
		Path:     data.Path,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return
	}

	dir := filepath.Dir(p)
	tmpFile, err := os.CreateTemp(dir, "dirscan_cache_*.tmp")
	if err != nil {
		return
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if err := tmpFile.Chmod(0o600); err != nil {
		_ = tmpFile.Close()
		return
	}
	if _, err := tmpFile.Write(b); err != nil {
		_ = tmpFile.Close()
		return
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return
	}
	if err := tmpFile.Close(); err != nil {
		return
	}

	_ = os.Rename(tmpName, p)
}
