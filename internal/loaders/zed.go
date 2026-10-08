// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
	"github.com/klauspost/compress/zstd"

	_ "modernc.org/sqlite"
)

// zedThreadRow holds columns read from Zed's threads table.
type zedThreadRow struct {
	id          string
	summary     string
	updatedAt   string
	dataType    string
	data        []byte
	parentID    sql.NullString
	branch      sql.NullString
	folderPaths sql.NullString
	folderOrder sql.NullString
	createdAt   sql.NullString
}

type zedThreadPayload struct {
	Title                  string                    `json:"title"`
	UpdatedAt              string                    `json:"updated_at"`
	Version                string                    `json:"version"`
	Profile                string                    `json:"profile"`
	Model                  zedModelPayload           `json:"model"`
	CumulativeTokenUsage   zedTokenUsagePayload      `json:"cumulative_token_usage"`
	InitialProjectSnapshot zedInitialProjectSnapshot `json:"initial_project_snapshot"`
	Messages               []json.RawMessage         `json:"messages"`
}

type zedModelPayload struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type zedTokenUsagePayload struct {
	InputTokens          int64 `json:"input_tokens"`
	OutputTokens         int64 `json:"output_tokens"`
	CacheReadInputTokens int64 `json:"cache_read_input_tokens"`
}

type zedInitialProjectSnapshot struct {
	Timestamp         string `json:"timestamp"`
	WorktreeSnapshots []struct {
		WorktreePath string `json:"worktree_path"`
	} `json:"worktree_snapshots"`
}

func parseZedTimestamp(val string) (time.Time, bool) {
	val = strings.TrimSpace(val)
	if val == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339Nano, val); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", val, time.Local); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02", val, time.Local); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// LoadZedData reads Zed's assistant conversations from threads.db
// (~/.local/share/zed/threads/threads.db).
func LoadZedData(dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	if dbPath == "" {
		dbPath, _ = probePlatformDataPath(thermal.HomeDir(), filepath.Join("zed", "threads"), "threads.db")
	}

	if _, err := os.Stat(dbPath); err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("zed: database not found: %s", dbPath)
	}

	canonicalPath := CanonicalDatabasePath(dbPath)
	srcID, err := zedSourceIdentity(canonicalPath)
	if err == nil {
		if cached, ok := loadZedCache(canonicalPath, srcID); ok {
			return cached.Summary, cached.Daily, cached.Projects, nil
		}
	}

	db, err := sql.Open("sqlite", sqliteReadOnlyDSN(canonicalPath, 30000000000))
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	var hasThreads bool
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='threads'`).Scan(&hasThreads); err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	if !hasThreads {
		return thermal.Summary{}, nil, nil, fmt.Errorf("zed: threads table not found in %s", canonicalPath)
	}

	rows, err := db.Query(`
		SELECT id, summary, updated_at, data_type, data, parent_id, worktree_branch, folder_paths, folder_paths_order, created_at
		FROM threads
	`)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer rows.Close()

	dec, err := zstd.NewReader(nil)
	if err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("zed: failed to initialize zstd decoder: %w", err)
	}
	defer dec.Close()

	var summary thermal.Summary
	summary.ModelBreakdown = make(map[string]int64)

	byDay := make(map[string]*thermal.DailyRow)
	byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)

	for rows.Next() {
		var tr zedThreadRow
		if err := rows.Scan(&tr.id, &tr.summary, &tr.updatedAt, &tr.dataType, &tr.data, &tr.parentID, &tr.branch, &tr.folderPaths, &tr.folderOrder, &tr.createdAt); err != nil {
			summary.Warnings = append(summary.Warnings, fmt.Sprintf("zed: failed to scan thread: %v", err))
			continue
		}

		var rawJSON []byte
		if tr.dataType == "zstd" {
			decompressed, err := dec.DecodeAll(tr.data, nil)
			if err != nil {
				summary.Warnings = append(summary.Warnings, fmt.Sprintf("thread %s: zstd decompression failed: %v", tr.id, err))
				continue
			}
			rawJSON = decompressed
		} else {
			rawJSON = tr.data
		}

		var payload zedThreadPayload
		if err := json.Unmarshal(rawJSON, &payload); err != nil {
			summary.Warnings = append(summary.Warnings, fmt.Sprintf("thread %s: json unmarshal failed: %v", tr.id, err))
			continue
		}

		tsStr := tr.createdAt.String
		if tsStr == "" {
			tsStr = payload.InitialProjectSnapshot.Timestamp
		}
		if tsStr == "" {
			tsStr = tr.updatedAt
		}
		if tsStr == "" {
			tsStr = payload.UpdatedAt
		}

		startT, ok := parseZedTimestamp(tsStr)
		if !ok || startT.IsZero() {
			continue
		}
		day := thermal.LocalDay(startT.Local())

		if tr.updatedAt != "" {
			if endT, ok := parseZedTimestamp(tr.updatedAt); ok && endT.After(startT) {
				durMs := endT.Sub(startT).Milliseconds()
				if durMs > summary.LongestSessionMs {
					summary.LongestSessionMs = durMs
				}
			}
		}

		rawModel := payload.Model.Model
		if rawModel == "" {
			rawModel = payload.Model.Provider
		}
		canonModel := modelName(rawModel)

		inTok := payload.CumulativeTokenUsage.InputTokens
		outTok := payload.CumulativeTokenUsage.OutputTokens
		cacheTok := payload.CumulativeTokenUsage.CacheReadInputTokens
		totTok := inTok + outTok + cacheTok

		turns := len(payload.Messages)
		if turns <= 0 {
			turns = 1
		}

		folderPath := tr.folderPaths.String
		if folderPath == "" && len(payload.InitialProjectSnapshot.WorktreeSnapshots) > 0 {
			folderPath = payload.InitialProjectSnapshot.WorktreeSnapshots[0].WorktreePath
		}
		var projectKey string
		if folderPath != "" {
			projectKey = thermal.ProjectKey(folderPath)
		}

		summary.Sessions++
		summary.LifetimeTokens += totTok
		summary.InputTokens += inTok
		summary.OutputTokens += outTok
		summary.CacheTokens += cacheTok
		if canonModel != "" {
			summary.ModelBreakdown[canonModel]++
		}

		row := byDay[day]
		if row == nil {
			row = &thermal.DailyRow{
				Day:    day,
				Models: make(map[string]thermal.ModelTokens),
			}
			byDay[day] = row
		}
		row.Tokens += totTok
		row.Input += inTok
		row.Output += outTok
		row.Cache += cacheTok
		row.Turns += turns
		if canonModel != "" {
			row.Models[canonModel] = row.Models[canonModel].Add(thermal.ModelTokens{
				Input:     inTok,
				Output:    outTok,
				CacheRead: cacheTok,
			})
		}

		if projectKey != "" {
			key := projectDayKey{day: day, project: projectKey}
			pd := byProjectDay[key]
			if pd == nil {
				pd = &thermal.ProjectDay{
					Project: projectKey,
					Day:     day,
					Models:  make(map[string]thermal.ModelTokens),
				}
				byProjectDay[key] = pd
			}
			pd.Tokens += totTok
			pd.Input += inTok
			pd.Output += outTok
			pd.CacheRead += cacheTok
			pd.Turns += turns
			if canonModel != "" {
				pd.Models[canonModel] = pd.Models[canonModel].Add(thermal.ModelTokens{
					Input:     inTok,
					Output:    outTok,
					CacheRead: cacheTok,
				})
			}
		}
	}

	if err := rows.Err(); err != nil {
		return thermal.Summary{}, nil, nil, err
	}

	daily := make([]thermal.DailyRow, 0, len(byDay))
	for _, row := range byDay {
		daily = append(daily, *row)
	}
	sort.Slice(daily, func(i, j int) bool {
		return daily[i].Day < daily[j].Day
	})

	projects := make([]thermal.ProjectDay, 0, len(byProjectDay))
	for _, pd := range byProjectDay {
		projects = append(projects, *pd)
	}
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Day != projects[j].Day {
			return projects[i].Day < projects[j].Day
		}
		return projects[i].Project < projects[j].Project
	})

	if srcID != "" {
		saveZedCache(canonicalPath, ZedCache{
			Version:       zedCacheVersion,
			CanonicalPath: canonicalPath,
			SourceID:      srcID,
			Summary:       summary,
			Daily:         daily,
			Projects:      projects,
		})
	}

	return summary, daily, projects, nil
}
