// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestGetToolSourceSig_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbFile := filepath.Join(tmpDir, "test.sqlite")
	if err := os.WriteFile(dbFile, []byte("fake sqlite database"), 0644); err != nil {
		t.Fatal(err)
	}

	sig1 := GetToolSourceSig(thermal.ToolDevin, ToolInfo{DBPath: dbFile}, "")
	if sig1.Missing {
		t.Errorf("expected sig1 not missing")
	}
	if sig1.FileCount != 1 || sig1.TotalSize != int64(len("fake sqlite database")) {
		t.Errorf("unexpected sig1: %+v", sig1)
	}

	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(dbFile, []byte("fake sqlite database updated"), 0644); err != nil {
		t.Fatal(err)
	}

	sig2 := GetToolSourceSig(thermal.ToolDevin, ToolInfo{DBPath: dbFile}, "")
	if sig2.TotalSize == sig1.TotalSize || sig2.MaxNano == sig1.MaxNano {
		t.Errorf("expected signature to change upon file modification: sig1=%+v, sig2=%+v", sig1, sig2)
	}
}

func TestSourceSigStability(t *testing.T) {
	allTools := AllTools()
	for toolName, info := range allTools {
		sig1 := GetToolSourceSig(toolName, info, "")
		sig2 := GetToolSourceSig(toolName, info, "")
		if sig1 != sig2 {
			t.Errorf("unstable sig for %s: sig1=%+v, sig2=%+v", toolName, sig1, sig2)
		}
	}
}
