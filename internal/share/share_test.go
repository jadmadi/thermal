// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package share

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestShareToken_Roundtrip(t *testing.T) {
	days := []thermal.DailyRow{
		{Day: "2026-09-01", Tokens: 100_000, Turns: 10},
		{Day: "2026-09-02", Tokens: 150_000, Turns: 12},
		{Day: "2026-09-03", Tokens: 200_000, Turns: 15},
	}

	snap := BuildShareSnapshot("opencode", days, 4.50, false)

	url, token, err := ShareURL(snap)
	if err != nil {
		t.Fatalf("failed generating share URL: %v", err)
	}

	if !strings.HasPrefix(url, DefaultShareBase) {
		t.Errorf("expected URL to start with %s, got %s", DefaultShareBase, url)
	}

	if !strings.HasPrefix(token, "v1.") {
		t.Errorf("expected token to start with v1., got %s", token)
	}

	decoded, err := DecodeShareToken(token)
	if err != nil {
		t.Fatalf("failed decoding token: %v", err)
	}

	if decoded.Tool != "opencode" {
		t.Errorf("decoded Tool = %q, want %q", decoded.Tool, "opencode")
	}
	if decoded.CurrentStreak != 3 {
		t.Errorf("decoded CurrentStreak = %d, want 3", decoded.CurrentStreak)
	}
	if decoded.TotalTokens != 450_000 {
		t.Errorf("decoded TotalTokens = %d, want 450,000", decoded.TotalTokens)
	}
	if decoded.TotalCost != 4.50 {
		t.Errorf("decoded TotalCost = %f, want 4.50", decoded.TotalCost)
	}
}

func TestShareToken_TamperDefense(t *testing.T) {
	snap := TelemetryShareSnapshot{
		Version:       1,
		GeneratedAt:   time.Now().Unix(),
		Tool:          "devin",
		CurrentStreak: 10,
		TotalTokens:   1_000_000,
	}

	token, err := EncodeShareToken(snap)
	if err != nil {
		t.Fatalf("failed encoding: %v", err)
	}

	// 1. Corrupt the payload part (change a char in the base64 part)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}

	tamperedPayload := parts[2]
	if len(tamperedPayload) > 2 {
		tamperedPayload = "A" + tamperedPayload[1:]
	}
	tamperedToken := parts[0] + "." + parts[1] + "." + tamperedPayload

	_, err = DecodeShareToken(tamperedToken)
	if !errors.Is(err, ErrChecksumMismatch) && err == nil {
		t.Errorf("expected ErrChecksumMismatch, got %v", err)
	}

	// 2. Corrupt the checksum part
	tamperedChecksumToken := parts[0] + ".ffffffff." + parts[2]
	_, err = DecodeShareToken(tamperedChecksumToken)
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Errorf("expected ErrChecksumMismatch on corrupted checksum, got %v", err)
	}
}

func TestShareToken_Malformed(t *testing.T) {
	tests := []struct {
		token string
		err   error
	}{
		{"", ErrInvalidTokenFormat},
		{"v1.invalid", ErrInvalidTokenFormat},
		{"v2.12345678.abcd", ErrUnsupportedVersion},
	}

	for _, tt := range tests {
		_, err := DecodeShareToken(tt.token)
		if !errors.Is(err, tt.err) {
			t.Errorf("DecodeShareToken(%q) err = %v, want %v", tt.token, err, tt.err)
		}
	}
}
