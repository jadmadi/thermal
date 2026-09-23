// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package pricing

import (
	"testing"
)

func TestFindModelIdentity(t *testing.T) {
	tests := []struct {
		input      string
		wantID     string
		wantFamily string
		wantFound  bool
	}{
		{"claude-3.5-sonnet", "claude-3.5-sonnet", "claude", true},
		{"deepseek-chat", "deepseek-v3", "deepseek", true},
		{"deepseek-v3", "deepseek-v3", "deepseek", true},
		{"gpt-4o", "gpt-4o", "gpt", true},
		{"gemini-2.0-flash", "gemini-2.0-flash", "gemini", true},
		{"unknown-custom-model", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, ok := FindModelIdentity(tt.input)
			if ok != tt.wantFound {
				t.Fatalf("FindModelIdentity(%q) found = %v, want %v", tt.input, ok, tt.wantFound)
			}
			if ok {
				if id.CanonicalID != tt.wantID {
					t.Errorf("CanonicalID = %q, want %q", id.CanonicalID, tt.wantID)
				}
				if id.Family != tt.wantFamily {
					t.Errorf("Family = %q, want %q", id.Family, tt.wantFamily)
				}
			}
		})
	}
}
