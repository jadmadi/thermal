// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

// StatuslineData carries lightweight state for terminal prompt and statusline integration.
type StatuslineData struct {
	Streak        int    `json:"streak"`
	LongestStreak int    `json:"longestStreak"`
	TodayTokens   int64  `json:"todayTokens"`
	TodayTurns    int    `json:"todayTurns"`
	ActiveTool    string `json:"activeTool"`
	Day           string `json:"day"`
	UpdatedAt     int64  `json:"updatedAt"` // Unix timestamp in seconds
}

// StatuslineOptions configures statusline formatting.
type StatuslineOptions struct {
	NoColor bool
	Nerd    bool // use Nerd Font flame symbol (󰈸)
	Plain   bool // omit emoji/glyphs completely
	JSON    bool
}

// StatuslineCachePath returns the canonical path to the statusline snapshot cache.
func StatuslineCachePath(homeDir string) string {
	if homeDir == "" {
		homeDir = thermal.HomeDir()
	}
	return filepath.Join(homeDir, ".cache", "thermal", "statusline.json")
}

// LoadStatuslineCache loads the cached statusline snapshot if it is fresh (<120s old and matching today).
func LoadStatuslineCache(homeDir string) (StatuslineData, bool) {
	cachePath := StatuslineCachePath(homeDir)
	fi, err := os.Stat(cachePath)
	if err != nil {
		return StatuslineData{}, false
	}

	// Stale if modified more than 120 seconds ago
	if time.Since(fi.ModTime()) > 120*time.Second {
		return StatuslineData{}, false
	}

	dataBytes, err := os.ReadFile(cachePath)
	if err != nil {
		return StatuslineData{}, false
	}

	var data StatuslineData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return StatuslineData{}, false
	}

	today := thermal.LocalDay(time.Now())
	if data.Day != today {
		return StatuslineData{}, false
	}

	return data, true
}

// SaveStatuslineCache writes the statusline snapshot to disk for sub-3ms prompt rendering.
func SaveStatuslineCache(homeDir string, data StatuslineData) error {
	cachePath := StatuslineCachePath(homeDir)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	tmpPath := fmt.Sprintf("%s.tmp.%d", cachePath, time.Now().UnixNano())
	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, cachePath)
}

// RenderStatusline formats StatuslineData into a concise, one-line string for shell prompts.
func RenderStatusline(data StatuslineData, opts StatuslineOptions) string {
	if opts.JSON {
		bytes, _ := json.Marshal(data)
		return string(bytes) + "\n"
	}

	colors := !opts.NoColor

	// 1. Flame / Inactivity indicator
	var icon string
	if !opts.Plain {
		if data.Streak > 0 {
			if opts.Nerd {
				icon = "󰈸 "
			} else {
				icon = "🔥 "
			}
		} else {
			if opts.Nerd {
				icon = "󰒲 "
			} else {
				icon = "💤 "
			}
		}
	}

	streakText := fmt.Sprintf("%dd", data.Streak)
	if colors {
		if data.Streak > 0 {
			streakText = theme.Primary.Sprint(true, icon+streakText)
		} else {
			streakText = theme.TextMuted.Sprint(true, icon+streakText)
		}
	} else {
		streakText = icon + streakText
	}

	// 2. Today's Token Burn or Turns
	var tokenText string
	if data.TodayTokens > 0 {
		tokenText = fmt.Sprintf("%s tok", thermal.CompactNumber(data.TodayTokens))
	} else if data.TodayTurns > 0 {
		tokenText = fmt.Sprintf("%d turns", data.TodayTurns)
	} else {
		tokenText = "0 tok"
	}

	if colors {
		if data.TodayTokens >= 1_000_000 {
			tokenText = theme.Primary.SprintBold(true, tokenText)
		} else if data.TodayTokens > 0 {
			tokenText = theme.Secondary.Sprint(true, tokenText)
		} else {
			tokenText = theme.TextMuted.Sprint(true, tokenText)
		}
	}

	// 3. Active Tool
	var toolText string
	if data.ActiveTool != "" {
		if colors {
			toolText = theme.Text.Sprint(true, data.ActiveTool)
		} else {
			toolText = data.ActiveTool
		}
	}

	// 4. Join parts
	sep := " · "
	if colors {
		sep = theme.TextMuted.Sprint(true, " · ")
	}

	parts := []string{streakText, tokenText}
	if toolText != "" {
		parts = append(parts, toolText)
	}

	return strings.Join(parts, sep) + "\n"
}
