package loaders

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var journalMagic = []byte{0xd9, 0xd5, 0x05, 0xf9, 0x20, 0xa1, 0x63, 0xd7}

func TestSqliteReadOnlyDSN(t *testing.T) {
	// 1. Basic path formatting
	dsn := sqliteReadOnlyDSN("/tmp/test.db", 268435456)
	expected := "file:/tmp/test.db?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=268435456&_pragma=trusted_schema=0"
	if dsn != expected {
		t.Fatalf("expected %q, got %q", expected, dsn)
	}

	// 2. Windows-style backslashes
	winDSN := sqliteReadOnlyDSN(`C:\Users\Developer\AppData\Local\data.db`, 0)
	expectedWin := "file:C:/Users/Developer/AppData/Local/data.db?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=268435456&_pragma=trusted_schema=0"
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

func TestRejectSubstitutableObjects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "objects.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE ordinary (id INTEGER);
		CREATE VIEW substitutable_view AS SELECT 1 AS id;
		CREATE VIEW broken_view AS SELECT missing_column FROM missing_table;
		CREATE VIRTUAL TABLE substitutable_fts USING fts5(body);
	`)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	db.Close()

	ro, err := openBoundedStore(path, 0)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer ro.Close()

	if err := rejectSubstitutableObjects(ro, "ordinary", "absent_table"); err != nil {
		t.Fatalf("ordinary or absent object rejected: %v", err)
	}
	for _, name := range []string{"substitutable_view", "broken_view", "substitutable_fts"} {
		err := rejectSubstitutableObjects(ro, name)
		if err == nil || !strings.Contains(err.Error(), "is not an ordinary table") {
			t.Fatalf("%s: expected ordinary-table rejection, got %v", name, err)
		}
	}

	if hasColumn(ro, "broken_view", "id") {
		t.Fatal("hasColumn returned true for a view")
	}
	if !hasColumn(ro, "ordinary", "id") {
		t.Fatal("hasColumn returned false for an ordinary table column")
	}
}

func TestLoadersRejectSubstitutedSchemaObjects(t *testing.T) {
	cases := []struct {
		name  string
		table string
		load  func(string) error
	}{
		{"opencode", "session_v2", func(p string) error {
			_, _, _, err := LoadOpenCodeData(p)
			return err
		}},
		{"mimocode", "message", func(p string) error {
			_, _, _, err := LoadMiMoCodeData(p)
			return err
		}},
		{"devin", "sessions", func(p string) error {
			_, _, _, err := LoadDevinData(p)
			return err
		}},
		{"muse", "sessions", func(p string) error {
			_, _, _, err := LoadMuseData(p)
			return err
		}},
		{"zcode", "model_usage", func(p string) error {
			_, _, _, err := LoadZCodeData(p)
			return err
		}},
		{"codex", "threads", func(p string) error {
			_, _, _, err := loadCodexFromStateDB(p)
			return err
		}},
		{"hermes", "sessions", func(p string) error {
			_, _, _, err := LoadHermesData(p)
			return err
		}},
		{"zed", "threads", func(p string) error {
			_, _, _, err := LoadZedData(p)
			return err
		}},
	}

	objects := []struct {
		name string
		sql  string
	}{
		{"view", `CREATE VIEW %s AS SELECT 1 AS id;`},
		{"virtual_table", `CREATE VIRTUAL TABLE %s USING fts5(body);`},
	}

	for _, tc := range cases {
		for _, obj := range objects {
			t.Run(tc.name+"/"+obj.name, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "store.sqlite")
				db, err := sql.Open("sqlite", path)
				if err != nil {
					t.Fatalf("create failed: %v", err)
				}
				if _, err := db.Exec(fmt.Sprintf(obj.sql, tc.table)); err != nil {
					db.Close()
					t.Fatalf("seed failed: %v", err)
				}
				db.Close()

				err = tc.load(path)
				if err == nil || !strings.Contains(err.Error(), "is not an ordinary table") {
					t.Fatalf("expected substitution rejection, got %v", err)
				}
			})
		}
	}

	t.Run("agy/conversation_summaries", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "conversation_summaries.db")
		db, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		if _, err := db.Exec(`CREATE VIEW conversation_summaries AS SELECT 1 AS conversation_id;`); err != nil {
			db.Close()
			t.Fatalf("seed failed: %v", err)
		}
		db.Close()

		if got := loadAgyWorkspaces(dir); len(got) != 0 {
			t.Fatalf("expected substituted agy store to be skipped, got %v", got)
		}
	})
}

func TestSQLiteEngineVersionAndCVEMitigations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "engine.db")
	rw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := rw.Exec(`CREATE VIRTUAL TABLE ft USING fts5(body);`); err != nil {
		rw.Close()
		t.Fatalf("fts5 create failed: %v", err)
	}
	if _, err := rw.Exec(`INSERT INTO ft(body) VALUES ('alpha beta gamma'), ('delta epsilon zeta');`); err != nil {
		rw.Close()
		t.Fatalf("fts5 seed failed: %v", err)
	}
	if _, err := rw.Exec(`INSERT INTO ft(ft) VALUES('optimize');`); err != nil {
		rw.Close()
		t.Fatalf("fts5 optimize failed: %v", err)
	}

	rows, err := rw.Query(`SELECT id, block FROM ft_data WHERE id > 10`)
	if err != nil {
		rw.Close()
		t.Fatalf("reading fts5 shadow table failed: %v", err)
	}
	type page struct {
		id    int64
		block []byte
	}
	var pages []page
	for rows.Next() {
		var p page
		if err := rows.Scan(&p.id, &p.block); err != nil {
			rows.Close()
			rw.Close()
			t.Fatalf("scan failed: %v", err)
		}
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		rw.Close()
		t.Fatalf("fts5 page iteration failed: %v", err)
	}
	rows.Close()
	if len(pages) == 0 {
		rw.Close()
		t.Fatal("expected at least one fts5 segment page")
	}

	// A leaf page stores szLeaf as a big-endian u16 at bytes 2..3. Force it
	// below the four-byte minimum that the CVE fix validates.
	for _, p := range pages {
		if len(p.block) < 4 {
			continue
		}
		p.block[2], p.block[3] = 0, 1
		if _, err := rw.Exec(`UPDATE ft_data SET block = ? WHERE id = ?`, p.block, p.id); err != nil {
			rw.Close()
			t.Fatalf("corrupting fts5 page failed: %v", err)
		}
	}
	rw.Close()

	ro, err := openBoundedStore(path, 0)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer ro.Close()

	var version string
	if err := queryRowStore(ro, []any{&version}, `SELECT sqlite_version()`); err != nil {
		t.Fatalf("sqlite_version failed: %v", err)
	}
	if !sqliteVersionAtLeast(version, 3, 53, 2) {
		t.Fatalf("embedded sqlite %s is older than 3.53.2", version)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("malformed FTS5 query panicked: %v", r)
		}
	}()
	queryErr := func() error {
		rows, err := queryStore(ro, `SELECT rowid FROM ft WHERE ft MATCH 'alpha'`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var rowid int64
			_ = rows.Scan(&rowid)
		}
		return rows.Err()
	}()
	if queryErr == nil {
		t.Fatal("expected a corruption error from the malformed FTS5 leaf, got nil")
	}
	if !strings.Contains(strings.ToLower(queryErr.Error()), "malformed") {
		t.Fatalf("expected a malformed-image error, got %v", queryErr)
	}
	t.Logf("malformed FTS5 leaf rejected with: %v", queryErr)
}

func TestReadOnlyHotJournalAndSuperJournalNoUnlink(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "store.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE sessions (id INTEGER PRIMARY KEY); INSERT INTO sessions(id) VALUES (1);`); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	db.Close()

	canary := filepath.Join(dir, "victim-mj000000900")
	if err := os.WriteFile(canary, []byte("canary"), 0o600); err != nil {
		t.Fatalf("canary write failed: %v", err)
	}
	forgeHotJournal(t, dbPath, canary)

	dbBefore := readFileBytes(t, dbPath)
	journalPath := dbPath + "-journal"
	journalBefore := readFileBytes(t, journalPath)
	canaryBefore := readFileBytes(t, canary)

	if _, _, _, err := LoadMuseData(dbPath); err == nil {
		t.Fatal("expected read-only loader to fail closed on a hot journal")
	}

	if got := readFileBytes(t, dbPath); !bytes.Equal(got, dbBefore) {
		t.Fatal("database file was rewritten during read-only recovery")
	}
	if got := readFileBytes(t, journalPath); !bytes.Equal(got, journalBefore) {
		t.Fatal("journal file was rewritten during read-only recovery")
	}
	if got := readFileBytes(t, canary); !bytes.Equal(got, canaryBefore) {
		t.Fatal("super-journal canary was modified during read-only recovery")
	}
}

func TestReadOnlyWALCloseNoCheckpoint(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "wal.db")
	writer, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	defer writer.Close()
	writer.SetMaxOpenConns(1)

	var mode string
	if err := writer.QueryRow(`PRAGMA journal_mode=WAL`).Scan(&mode); err != nil {
		t.Fatalf("enabling WAL failed: %v", err)
	}
	if !strings.EqualFold(mode, "wal") {
		t.Fatalf("expected WAL journal mode, got %q", mode)
	}
	if _, err := writer.Exec(`PRAGMA wal_autocheckpoint=0`); err != nil {
		t.Fatalf("disabling autocheckpoint failed: %v", err)
	}
	if _, err := writer.Exec(`CREATE TABLE sessions (id INTEGER PRIMARY KEY, payload BLOB)`); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
	payload := bytes.Repeat([]byte{0x5a}, 4096)
	for i := 0; i < 64; i++ {
		if _, err := writer.Exec(`INSERT INTO sessions(payload) VALUES (?)`, payload); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	walPath := dbPath + "-wal"
	shmPath := dbPath + "-shm"
	if info, err := os.Stat(walPath); err != nil || info.Size() == 0 {
		t.Fatalf("expected a populated WAL file, got %v", err)
	}

	dbBefore := readFileBytes(t, dbPath)
	walBefore := readFileBytes(t, walPath)

	ro, err := openBoundedStore(dbPath, 0)
	if err != nil {
		t.Fatalf("read-only open failed: %v", err)
	}
	var count int
	if err := queryRowStore(ro, []any{&count}, `SELECT COUNT(*) FROM sessions`); err != nil {
		t.Fatalf("read-only query failed: %v", err)
	}
	if count != 64 {
		t.Fatalf("expected 64 rows, got %d", count)
	}
	if err := ro.Close(); err != nil {
		t.Fatalf("read-only close failed: %v", err)
	}

	if got := readFileBytes(t, dbPath); !bytes.Equal(got, dbBefore) {
		t.Fatal("database file was rewritten by a read-only close")
	}
	if got := readFileBytes(t, walPath); !bytes.Equal(got, walBefore) {
		t.Fatal("WAL file was rewritten by a read-only close")
	}
	for _, path := range []string{walPath, shmPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s was unlinked: %v", filepath.Base(path), err)
		}
	}
}

func TestBoundedStoreValueLimitAndTimeout(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "bounded.db")
	rw, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := rw.Exec(`CREATE TABLE payloads (id INTEGER PRIMARY KEY, body BLOB)`); err != nil {
		rw.Close()
		t.Fatalf("schema failed: %v", err)
	}
	if _, err := rw.Exec(`INSERT INTO payloads(id, body) VALUES (1, zeroblob(?)), (2, zeroblob(?))`, storeValueLimit+1024, 1024); err != nil {
		rw.Close()
		t.Fatalf("insert failed: %v", err)
	}
	rw.Close()

	db, err := openBoundedStore(dbPath, 0)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer db.Close()

	var small []byte
	if err := queryRowStore(db, []any{&small}, `SELECT body FROM payloads WHERE id = 2`); err != nil {
		t.Fatalf("small value read failed: %v", err)
	}
	var large []byte
	if err := queryRowStore(db, []any{&large}, `SELECT body FROM payloads WHERE id = 1`); err == nil {
		t.Fatal("expected the 16 MiB value limit to reject a larger value")
	} else if !strings.Contains(strings.ToLower(err.Error()), "too big") {
		t.Fatalf("expected SQLITE_TOOBIG, got %v", err)
	}

	rows, err := queryStoreWithBudget(db, time.Millisecond, `
		WITH RECURSIVE counter(x) AS (
			SELECT 1 UNION ALL SELECT x + 1 FROM counter WHERE x < 1000000000
		)
		SELECT sum(x) FROM counter
	`)
	if err == nil {
		var sum int64
		if rows.Next() {
			_ = rows.Scan(&sum)
		}
		err = rows.Err()
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
	}
	if err == nil {
		t.Fatal("expected the query budget to stop a long-running query")
	}
	msg := strings.ToLower(err.Error())
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(msg, "interrupt") && !strings.Contains(msg, "deadline") {
		t.Fatalf("expected a deadline or interrupt error, got %v", err)
	}
}

func sqliteVersionAtLeast(version string, major, minor, patch int) bool {
	var ma, mi, pa int
	if _, err := fmt.Sscanf(version, "%d.%d.%d", &ma, &mi, &pa); err != nil {
		return false
	}
	if ma != major {
		return ma > major
	}
	if mi != minor {
		return mi > minor
	}
	return pa >= patch
}

func forgeHotJournal(t *testing.T, dbPath, superPath string) {
	t.Helper()
	name := []byte(superPath)
	var checksum uint32
	for _, b := range name {
		checksum += uint32(b)
	}

	var buf bytes.Buffer
	buf.Write(journalMagic)
	buf.Write(make([]byte, 20))
	buf.Write(name)

	var trailer [16]byte
	binary.BigEndian.PutUint32(trailer[0:4], uint32(len(name)))
	binary.BigEndian.PutUint32(trailer[4:8], checksum)
	copy(trailer[8:16], journalMagic)
	buf.Write(trailer[:])

	if err := os.WriteFile(dbPath+"-journal", buf.Bytes(), 0o600); err != nil {
		t.Fatalf("forging hot journal failed: %v", err)
	}
}

func readFileBytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", path, err)
	}
	return b
}
