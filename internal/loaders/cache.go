// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jadmadi/thermal/internal/thermal"
)

// devinCacheVersion bumps whenever the cached snapshot shape changes, so an
// old cache file is ignored instead of decoded into stale zero fields.
const devinCacheVersion = 6

// DevinCache is a disk-backed snapshot of the expensive message_nodes
// aggregation. Invalidation is keyed on canonical source database identity,
// file replacement detection, WAL freshness, and content signatures.
type DevinCache struct {
	Version           int                  `json:"version"`
	CanonicalPath     string               `json:"canonicalPath"`
	SourceID          string               `json:"sourceId"`
	FreshnessKey      string               `json:"freshnessKey"`
	MaxRowID          int64                `json:"maxRowId"`
	BaseRowCount      int64                `json:"baseRowCount"`
	BaseRowLength     int64                `json:"baseRowLength"`
	SessionCount      int                  `json:"sessionCount"`
	SessionsSignature string               `json:"sessionsSignature"`
	PromptSignature   string               `json:"promptSignature,omitempty"`
	Summary           thermal.Summary      `json:"summary"`
	Daily             []thermal.DailyRow   `json:"daily"`
	Projects          []thermal.ProjectDay `json:"projects,omitempty"`
}

// CanonicalDatabasePath resolves symlinks and returns the normalized path.
func CanonicalDatabasePath(p string) string {
	realPath, err := filepath.EvalSymlinks(p)
	if err != nil {
		return filepath.Clean(p)
	}
	return realPath
}

// devinSourceIdentity computes a replacement-sensitive identity for a database file.
func devinSourceIdentity(canonicalPath string) (string, error) {
	fi, err := os.Stat(canonicalPath)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	fmt.Fprintf(h, "path:%s;size:%d;mod:%d;", canonicalPath, fi.Size(), fi.ModTime().UnixNano())

	// Read SQLite header bytes (up to 96 bytes)
	f, err := os.Open(canonicalPath)
	if err == nil {
		hdr := make([]byte, 96)
		n, _ := f.Read(hdr)
		_ = f.Close()
		h.Write(hdr[:n])
	}

	// WAL file presence and size (only when non-empty, as 0-byte WAL and SHM are touched by read locks)
	if wfi, err := os.Stat(canonicalPath + "-wal"); err == nil && wfi.Size() > 0 {
		fmt.Fprintf(h, "wal:size:%d;mod:%d;", wfi.Size(), wfi.ModTime().UnixNano())
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func devinCachePath(canonicalPath string) (string, error) {
	home := thermal.HomeDir()
	dir := filepath.Join(home, ".cache", "thermal", "devin")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(canonicalPath))
	fileName := fmt.Sprintf("cache_%s.json", hex.EncodeToString(h[:12]))
	return filepath.Join(dir, fileName), nil
}

func loadDevinCache(canonicalPath string, expectedSourceID string) (DevinCache, bool) {
	p, err := devinCachePath(canonicalPath)
	if err != nil {
		return DevinCache{}, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return DevinCache{}, false
	}
	var c DevinCache
	if err := json.Unmarshal(b, &c); err != nil {
		return DevinCache{}, false
	}
	if c.Version != devinCacheVersion {
		return DevinCache{}, false
	}
	if c.CanonicalPath != canonicalPath {
		return DevinCache{}, false
	}
	if expectedSourceID != "" && c.SourceID != expectedSourceID {
		return DevinCache{}, false
	}
	return c, true
}

func saveDevinCache(canonicalPath string, c DevinCache) {
	p, err := devinCachePath(canonicalPath)
	if err != nil {
		return
	}
	c.Version = devinCacheVersion
	c.CanonicalPath = canonicalPath
	b, err := json.Marshal(c)
	if err != nil {
		return
	}

	dir := filepath.Dir(p)
	tmpFile, err := os.CreateTemp(dir, "devin_cache_*.tmp")
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
