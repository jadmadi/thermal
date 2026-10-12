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

func TestGetToolSourceSig_SQLite_SHMIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	dbFile := filepath.Join(tmpDir, "test.sqlite")
	walFile := dbFile + "-wal"
	shmFile := dbFile + "-shm"

	if err := os.WriteFile(dbFile, []byte("fake sqlite database"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(walFile, []byte("fake wal log data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shmFile, []byte("fake shm shared memory headers"), 0644); err != nil {
		t.Fatal(err)
	}

	info := ToolInfo{DBPath: dbFile}
	sig1 := GetToolSourceSig(thermal.ToolDevin, info, "")
	if sig1.Missing || sig1.FileCount != 2 {
		t.Fatalf("expected 2 files (db + wal), got: %+v", sig1)
	}

	// 1. Modifying -shm should NOT change the signature (read lock touch immunity)
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(shmFile, []byte("fake shm updated read locks"), 0644); err != nil {
		t.Fatal(err)
	}
	sig2 := GetToolSourceSig(thermal.ToolDevin, info, "")
	if sig1 != sig2 {
		t.Errorf("expected signature to be immune to -shm changes: sig1=%+v, sig2=%+v", sig1, sig2)
	}

	// 2. Modifying -wal MUST change the signature
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(walFile, []byte("fake wal log data appended with commit"), 0644); err != nil {
		t.Fatal(err)
	}
	sig3 := GetToolSourceSig(thermal.ToolDevin, info, "")
	if sig3 == sig1 || sig3.TotalSize == sig1.TotalSize || sig3.MaxNano == sig1.MaxNano {
		t.Errorf("expected signature to detect -wal changes: sig1=%+v, sig3=%+v", sig1, sig3)
	}

	// 3. Modifying main db MUST change the signature
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(dbFile, []byte("fake sqlite database checkpointed"), 0644); err != nil {
		t.Fatal(err)
	}
	sig4 := GetToolSourceSig(thermal.ToolDevin, info, "")
	if sig4 == sig3 || sig4.MaxNano == sig3.MaxNano {
		t.Errorf("expected signature to detect .db changes: sig3=%+v, sig4=%+v", sig3, sig4)
	}
}
