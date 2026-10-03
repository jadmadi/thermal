package loaders

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestSqliteReadOnlyDSN(t *testing.T) {
	// 1. Basic path formatting
	dsn := sqliteReadOnlyDSN("/tmp/test.db", 268435456)
	expected := "file:/tmp/test.db?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=268435456"
	if dsn != expected {
		t.Fatalf("expected %q, got %q", expected, dsn)
	}

	// 2. Windows-style backslashes
	winDSN := sqliteReadOnlyDSN(`C:\Users\Developer\AppData\Local\data.db`, 0)
	expectedWin := "file:C:/Users/Developer/AppData/Local/data.db?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=268435456"
	if winDSN != expectedWin {
		t.Fatalf("expected %q, got %q", expectedWin, winDSN)
	}

	// 3. Real file opening with modernc.org/sqlite
	tmpDir := t.TempDir()
	dbFile := filepath.Join(tmpDir, "test.sqlite")
	createDB, err := sql.Open("sqlite", dbFile)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := createDB.Exec("CREATE TABLE test (id INT); INSERT INTO test VALUES (42);"); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	createDB.Close()

	roDSN := sqliteReadOnlyDSN(dbFile, 0)
	roDB, err := sql.Open("sqlite", roDSN)
	if err != nil {
		t.Fatalf("ro open failed: %v", err)
	}
	defer roDB.Close()

	var val int
	if err := roDB.QueryRow("SELECT id FROM test").Scan(&val); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}
