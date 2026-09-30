// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jadmadi/thermal/internal/thermal"
)

// SourceSig captures file size, count, and modification timestamp to detect local tool changes.
type SourceSig struct {
	FileCount int
	TotalSize int64
	MaxNano   int64
	Missing   bool
}

// GetToolSourceSig calculates a lightweight file system signature for a tool's data sources.
func GetToolSourceSig(t thermal.Tool, info ToolInfo, dbPathOverride string) SourceSig {
	switch t {
	case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin, thermal.ToolZCode, thermal.ToolMuse, thermal.ToolHermes:
		p := dbPathOverride
		if p == "" {
			p = info.DBPath
		}
		if p == "" {
			return SourceSig{Missing: true}
		}
		fi, err := os.Stat(p)
		if err != nil {
			return SourceSig{Missing: true}
		}
		sig := SourceSig{
			FileCount: 1,
			TotalSize: fi.Size(),
			MaxNano:   fi.ModTime().UnixNano(),
		}
		if wfi, err := os.Stat(p + "-wal"); err == nil {
			sig.FileCount++
			sig.TotalSize += wfi.Size()
			sig.MaxNano ^= wfi.ModTime().UnixNano()
		}
		if sfi, err := os.Stat(p + "-shm"); err == nil {
			sig.FileCount++
			sig.TotalSize += sfi.Size()
			sig.MaxNano ^= sfi.ModTime().UnixNano()
		}
		return sig

	default:
		dir := info.DataDir
		if dbPathOverride != "" {
			dir = dbPathOverride
		}
		if dir == "" {
			return SourceSig{Missing: true}
		}
		if t == thermal.ToolAgy {
			scanDir := ResolveAgyBrainDir(dir)
			if st, err := os.Stat(scanDir); err != nil || !st.IsDir() {
				return SourceSig{Missing: true}
			}
			var sig SourceSig
			if bst, err := os.Stat(scanDir); err == nil {
				sig.FileCount++
				sig.TotalSize += bst.Size()
				sig.MaxNano = bst.ModTime().UnixNano()
			}
			parentDir := filepath.Dir(scanDir)
			csDB := filepath.Join(parentDir, "conversation_summaries.db")
			if cst, err := os.Stat(csDB); err == nil {
				sig.FileCount++
				sig.TotalSize += cst.Size()
				if nano := cst.ModTime().UnixNano(); nano > sig.MaxNano {
					sig.MaxNano = nano
				}
			}
			if walt, err := os.Stat(csDB + "-wal"); err == nil {
				sig.FileCount++
				sig.TotalSize += walt.Size()
				if nano := walt.ModTime().UnixNano(); nano > sig.MaxNano {
					sig.MaxNano = nano
				}
			}
			entries, err := os.ReadDir(scanDir)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						continue
					}
					logsDir := filepath.Join(scanDir, entry.Name(), ".system_generated", "logs")
					transPath := filepath.Join(logsDir, "transcript.jsonl")
					fi, err := os.Stat(transPath)
					if err != nil {
						transPath = filepath.Join(logsDir, "overview.txt")
						fi, err = os.Stat(transPath)
					}
					if err == nil && !fi.IsDir() {
						sig.FileCount++
						sig.TotalSize += fi.Size()
						if nano := fi.ModTime().UnixNano(); nano > sig.MaxNano {
							sig.MaxNano = nano
						}
					}
				}
			}
			return sig
		}

		scanDir := dir
		if info.DataSubdir != "" {
			sub := filepath.Join(dir, info.DataSubdir)
			if _, err := os.Stat(sub); err == nil {
				scanDir = sub
			}
		}
		if _, err := os.Stat(scanDir); err != nil {
			return SourceSig{Missing: true}
		}

		var sig SourceSig
		if t == thermal.ToolCodex {
			stateDB := filepath.Join(dir, "state_5.sqlite")
			if sfi, err := os.Stat(stateDB); err == nil {
				sig.FileCount++
				sig.TotalSize += sfi.Size()
				sig.MaxNano = sfi.ModTime().UnixNano()
			}
		}

		_ = filepath.WalkDir(scanDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			sig.FileCount++
			sig.TotalSize += info.Size()
			if nano := info.ModTime().UnixNano(); nano > sig.MaxNano {
				sig.MaxNano = nano
			}
			return nil
		})
		return sig
	}
}
