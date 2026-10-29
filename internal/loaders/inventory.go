// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// InventoryOrder is the stable display order shared with the leaderboard.
var InventoryOrder = []thermal.Tool{
	thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolCodex, thermal.ToolDevin,
	thermal.ToolAgy, thermal.ToolCommandCode, thermal.ToolCodewhale, thermal.ToolZCode,
	thermal.ToolGrok, thermal.ToolMuse, thermal.ToolClaude, thermal.ToolDroid,
	thermal.ToolDsh, thermal.ToolHermes, thermal.ToolZed,
}

var inventoryAliases = map[thermal.Tool]string{
	thermal.ToolMiMoCode:    "mimo",
	thermal.ToolOpenCode:    "oc",
	thermal.ToolCommandCode: "cmd",
	thermal.ToolCodewhale:   "whale",
	thermal.ToolZCode:       "zc",
	thermal.ToolDsh:         "dsh",
	thermal.ToolHermes:      "hermes",
	thermal.ToolClaude:      "ccode",
	thermal.ToolAgy:         "agy",
	thermal.ToolMuse:        "muse",
	thermal.ToolDroid:       "droid",
}

// Inventory returns the stat-only source inventory for a home directory.
// It never opens a database, reads a transcript, or parses a session.
func Inventory(home string) []thermal.SourceInfo {
	return InventoryFromTools(AllToolsForHome(home), home)
}

// InventoryFromTools builds the inventory from an already resolved registry.
func InventoryFromTools(tools map[thermal.Tool]ToolInfo, home string) []thermal.SourceInfo {
	items := make([]thermal.SourceInfo, 0, len(InventoryOrder))
	for _, tool := range InventoryOrder {
		info, ok := tools[tool]
		if !ok {
			items = append(items, thermal.SourceInfo{
				Tool:   string(tool),
				Name:   toolDisplayName(tool),
				Alias:  inventoryAliases[tool],
				Shape:  shapeLabel(ToolPathShape(tool)),
				State:  thermal.SourceUnknown,
				Detail: "no configured path",
			})
			continue
		}
		items = append(items, sourceInfoFor(tool, info, home))
	}
	return items
}

func sourceInfoFor(tool thermal.Tool, info ToolInfo, home string) thermal.SourceInfo {
	shape := ToolPathShape(tool)
	path := info.DBPath
	if shape == ShapeDirectory {
		path = info.DataDir
	} else if shape == ShapeEither && path == "" {
		path = info.DataDir
	}

	item := thermal.SourceInfo{
		Tool:        string(tool),
		Name:        info.Name,
		Alias:       inventoryAliases[tool],
		Path:        path,
		DisplayPath: displayHomePath(path, home),
		Shape:       shapeLabel(shape),
	}
	if path == "" {
		item.State = thermal.SourceUnknown
		item.Detail = "no configured path"
		return item
	}

	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			item.State = thermal.SourceMissing
			if shape == ShapeSQLiteFile {
				item.Detail = "expected file"
			} else {
				item.Detail = "expected directory"
			}
		} else {
			item.State = thermal.SourceUnreadable
			item.Detail = "check permissions"
		}
		return item
	}
	if !fi.ModTime().IsZero() {
		item.Modified = fi.ModTime().UTC().Format(time.RFC3339)
	}

	isDir := fi.IsDir()
	switch shape {
	case ShapeSQLiteFile:
		if isDir {
			item.State = thermal.SourceWrongShape
			item.Detail = "expected file"
			return item
		}
		item.State, item.Detail = fileState(fi)
	case ShapeDirectory:
		if !isDir {
			item.State = thermal.SourceWrongShape
			item.Detail = "expected directory"
			return item
		}
		item.State, item.Detail = directoryState(path)
	case ShapeEither:
		if isDir {
			item.State, item.Detail = directoryState(path)
		} else {
			item.State, item.Detail = fileState(fi)
		}
	}
	return item
}

func fileState(fi os.FileInfo) (thermal.SourceState, string) {
	if fi.Size() == 0 {
		return thermal.SourceEmpty, "empty file"
	}
	return thermal.SourceFound, formatSourceBytes(fi.Size())
}

func directoryState(path string) (thermal.SourceState, string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return thermal.SourceUnreadable, "check permissions"
	}
	if len(entries) == 0 {
		return thermal.SourceEmpty, "no sessions yet"
	}
	return thermal.SourceFound, fmt.Sprintf("%d entries", len(entries))
}

func shapeLabel(shape PathShape) string {
	switch shape {
	case ShapeSQLiteFile:
		return "file"
	case ShapeDirectory:
		return "directory"
	default:
		return "either"
	}
}

func displayHomePath(path, home string) string {
	if path == "" || home == "" {
		return path
	}
	cleanHome := filepath.Clean(home)
	cleanPath := filepath.Clean(path)
	if cleanPath == cleanHome {
		return "~"
	}
	if strings.HasPrefix(cleanPath, cleanHome+string(filepath.Separator)) {
		return "~" + strings.TrimPrefix(cleanPath, cleanHome)
	}
	return path
}

func formatSourceBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// SummarizeInventory counts each source state.
func SummarizeInventory(items []thermal.SourceInfo) thermal.SourceSummary {
	return thermal.SummarizeSources(items)
}
