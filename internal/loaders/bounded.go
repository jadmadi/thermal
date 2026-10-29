// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// storeQueryBudget is the wall-clock ceiling for a single loader query.
const storeQueryBudget = 30 * time.Second

// storeValueLimit caps SQLITE_LIMIT_LENGTH so a single value can never exceed
// 16 MiB, independent of the engine's 1 GiB default.
const storeValueLimit = 16 << 20

// rowIterator covers the *sql.Rows surface the loaders consume, so bounded
// rows and plain rows are interchangeable.
type rowIterator interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
	Columns() ([]string, error)
}

// boundedRows releases its query context when the rows are closed.
type boundedRows struct {
	*sql.Rows
	cancel context.CancelFunc
}

func (r *boundedRows) Close() error {
	err := r.Rows.Close()
	r.cancel()
	return err
}

// openBoundedStore opens a read-only SQLite database and pins the per-value
// length limit on the single pooled connection. Every later query reuses that
// connection, so the limit and trusted-schema pragma persist.
func openBoundedStore(dbPath string, mmapSize int64) (*sql.DB, error) {
	db, err := sql.Open("sqlite", sqliteReadOnlyDSN(dbPath, mmapSize))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), storeQueryBudget)
	defer cancel()

	conn, err := db.Conn(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot acquire sqlite connection: %w", err)
	}
	if _, err := sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_LENGTH, storeValueLimit); err != nil {
		conn.Close()
		db.Close()
		return nil, fmt.Errorf("cannot set sqlite value limit: %w", err)
	}
	if err := conn.Close(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// queryStore runs a bounded query and returns rows that cancel their context
// on Close.
func queryStore(db *sql.DB, query string, args ...any) (*boundedRows, error) {
	return queryStoreWithBudget(db, storeQueryBudget, query, args...)
}

func queryStoreWithBudget(db *sql.DB, budget time.Duration, query string, args ...any) (*boundedRows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		cancel()
		return nil, err
	}
	return &boundedRows{Rows: rows, cancel: cancel}, nil
}

// queryRowStore runs a single-row bounded query into dest.
func queryRowStore(db *sql.DB, dest []any, query string, args ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), storeQueryBudget)
	defer cancel()
	return db.QueryRowContext(ctx, query, args...).Scan(dest...)
}
