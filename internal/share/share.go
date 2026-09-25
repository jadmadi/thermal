// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package share

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

const (
	TokenPrefix      = "v1"
	MaxDecompSize    = 1024 * 1024 // 1MB decompression safety ceiling
	DefaultShareBase = "https://jadmadi.net/projects/thermal/share#"
)

var (
	ErrInvalidTokenFormat = errors.New("invalid share token format")
	ErrUnsupportedVersion = errors.New("unsupported share token version")
	ErrChecksumMismatch   = errors.New("share token checksum verification failed")
	ErrDecompressionBomb  = errors.New("decompressed payload exceeded safe limit")
)

// TelemetryShareSnapshot is the strictly sanitized, privacy-safe telemetry payload
// encoded into stateless share URLs. It excludes file paths, repository roots, prompts,
// and completions.
type TelemetryShareSnapshot struct {
	Version       int              `json:"v"`
	GeneratedAt   int64            `json:"t"`
	Tool          string           `json:"tool,omitempty"`
	CurrentStreak int              `json:"currentStreak"`
	LongestStreak int              `json:"longestStreak"`
	ActiveDays    int              `json:"activeDays"`
	TotalTokens   int64            `json:"totalTokens"`
	TotalCost     float64          `json:"totalCost,omitempty"`
	EstimatedCost bool             `json:"estimatedCost,omitempty"`
	Days          map[string]int64 `json:"days,omitempty"`   // YYYY-MM-DD -> tokens/steps
	Models        map[string]int64 `json:"models,omitempty"` // canonical model -> token count
}

// BuildShareSnapshot constructs a sanitized share payload from aggregated tool days.
func BuildShareSnapshot(tool string, days []thermal.DailyRow, cost float64, estimated bool) TelemetryShareSnapshot {
	dayMap := make(map[string]bool)
	dayActivity := make(map[string]int64)
	models := make(map[string]int64)
	var totalTokens int64

	for _, d := range days {
		if d.Day == "" {
			continue
		}
		if d.Tokens > 0 || d.Turns > 0 {
			dayMap[d.Day] = true
			dayActivity[d.Day] += d.Tokens
			totalTokens += d.Tokens
		}
		for m, counts := range d.Models {
			models[m] += counts.Total()
		}
	}

	currStreak, longestStreak := thermal.ComputeStreaks(dayMap)

	// Keep top 8 models to bound token payload size
	topModels := make(map[string]int64)
	count := 0
	for m, tok := range models {
		topModels[m] = tok
		count++
		if count >= 8 {
			break
		}
	}

	return TelemetryShareSnapshot{
		Version:       1,
		GeneratedAt:   time.Now().Unix(),
		Tool:          tool,
		CurrentStreak: currStreak,
		LongestStreak: longestStreak,
		ActiveDays:    len(dayMap),
		TotalTokens:   totalTokens,
		TotalCost:     cost,
		EstimatedCost: estimated,
		Days:          dayActivity,
		Models:        topModels,
	}
}

// EncodeShareToken serializes a snapshot to compressed, URL-safe v1.<checksum>.<deflate-b64>.
func EncodeShareToken(snap TelemetryShareSnapshot) (string, error) {
	snap.Version = 1
	if snap.GeneratedAt == 0 {
		snap.GeneratedAt = time.Now().Unix()
	}

	rawJSON, err := json.Marshal(snap)
	if err != nil {
		return "", fmt.Errorf("failed marshaling share payload: %w", err)
	}

	var buf bytes.Buffer
	fw, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return "", fmt.Errorf("failed initializing flate compressor: %w", err)
	}
	if _, err := fw.Write(rawJSON); err != nil {
		return "", fmt.Errorf("failed compressing payload: %w", err)
	}
	if err := fw.Close(); err != nil {
		return "", fmt.Errorf("failed finalizing flate compressor: %w", err)
	}

	compressed := buf.Bytes()
	sum := sha256.Sum256(compressed)
	checksum := hex.EncodeToString(sum[:4]) // 8-char hex checksum

	b64Payload := base64.RawURLEncoding.EncodeToString(compressed)

	return fmt.Sprintf("%s.%s.%s", TokenPrefix, checksum, b64Payload), nil
}

// DecodeShareToken parses, checks integrity, and decompresses a share token.
func DecodeShareToken(token string) (TelemetryShareSnapshot, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return TelemetryShareSnapshot{}, ErrInvalidTokenFormat
	}

	if parts[0] != TokenPrefix {
		return TelemetryShareSnapshot{}, ErrUnsupportedVersion
	}

	expectedChecksum := parts[1]
	compressed, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		// Fallback to standard URLEncoding if padded
		var err2 error
		compressed, err2 = base64.URLEncoding.DecodeString(parts[2])
		if err2 != nil {
			return TelemetryShareSnapshot{}, fmt.Errorf("failed base64 decoding: %w", err)
		}
	}

	sum := sha256.Sum256(compressed)
	actualChecksum := hex.EncodeToString(sum[:4])
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return TelemetryShareSnapshot{}, ErrChecksumMismatch
	}

	fr := flate.NewReader(bytes.NewReader(compressed))
	defer fr.Close()

	limited := io.LimitReader(fr, MaxDecompSize+1)
	decompressed, err := io.ReadAll(limited)
	if err != nil && err != io.EOF {
		return TelemetryShareSnapshot{}, fmt.Errorf("failed decompressing payload: %w", err)
	}
	if len(decompressed) > MaxDecompSize {
		return TelemetryShareSnapshot{}, ErrDecompressionBomb
	}

	var snap TelemetryShareSnapshot
	if err := json.Unmarshal(decompressed, &snap); err != nil {
		return TelemetryShareSnapshot{}, fmt.Errorf("failed unmarshaling snapshot JSON: %w", err)
	}

	return snap, nil
}

// ShareURL builds the full sharing URL from a snapshot.
func ShareURL(snap TelemetryShareSnapshot) (string, string, error) {
	token, err := EncodeShareToken(snap)
	if err != nil {
		return "", "", err
	}
	return DefaultShareBase + token, token, nil
}
