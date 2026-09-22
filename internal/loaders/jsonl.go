// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// maxJSONLLineBytes is the upper bound for a single line in JSONL log files.
// Real assistant transcripts contain multi-megabyte payloads (e.g. tool results,
// base64 images, rollout input_text chunks). A 32 MiB ceiling ensures large lines
// are read in full without truncating the session or dropping subsequent messages.
const maxJSONLLineBytes = 32 * 1024 * 1024

// jsonlMaxLineBytes is the ceiling used by newJSONLScanner. It is a package variable
// so unit tests can lower it to test over-ceiling behavior without generating 32 MiB files.
var jsonlMaxLineBytes = maxJSONLLineBytes

// newJSONLScanner creates a bufio.Scanner configured with an initial buffer and
// the shared 32 MiB max line ceiling.
func newJSONLScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	initBuf := 64 * 1024
	if jsonlMaxLineBytes < initBuf {
		initBuf = jsonlMaxLineBytes
	}
	scanner.Buffer(make([]byte, 0, initBuf), jsonlMaxLineBytes)
	return scanner
}

// formatScanWarning formats a non-fatal warning when a JSONL scan encounters an error
// or exceeds the line ceiling.
func formatScanWarning(path string, err error) string {
	if errors.Is(err, bufio.ErrTooLong) {
		return fmt.Sprintf("%s: line exceeds %d byte ceiling, rest of file skipped", path, jsonlMaxLineBytes)
	}
	return fmt.Sprintf("%s: %v", path, err)
}

func loadJsonlData(dataDir string, fieldTimestamp string, useMillis bool) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	historyPath := filepath.Join(dataDir, "history.jsonl")
	f, err := os.Open(historyPath)
	if err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("cannot open %s: %w", historyPath, err)
	}
	defer f.Close()

	dayCounts := make(map[string]int)
	var totalCommands int64
	var warnings []string

	scanner := newJSONLScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		var ts int64
		if v, ok := raw[fieldTimestamp].(float64); ok {
			ts = int64(v)
		}
		if ts == 0 {
			continue
		}

		var t time.Time
		if useMillis {
			t = time.UnixMilli(ts)
		} else {
			t = time.Unix(ts, 0)
		}
		day := thermal.LocalDay(t)
		dayCounts[day]++
		totalCommands++
	}
	if err := scanner.Err(); err != nil {
		warnings = append(warnings, formatScanWarning(historyPath, err))
	}

	var daily []thermal.DailyRow
	for day, count := range dayCounts {
		daily = append(daily, thermal.DailyRow{
			Day:    day,
			Tokens: int64(count),
			Turns:  count,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	summary := thermal.Summary{
		Sessions:       0,
		LifetimeTokens: totalCommands,
		Warnings:       warnings,
	}

	return summary, daily, nil, nil
}
