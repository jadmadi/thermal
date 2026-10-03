// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"fmt"
	"strings"
)

// sqliteReadOnlyDSN formats a database file path for read-only SQLite access.
// It converts Windows backslashes to forward slashes with filepath.ToSlash and
// attaches read-only mode and memory mapping pragmas for high-throughput scanning.
func sqliteReadOnlyDSN(path string, mmapSize int64) string {
	clean := strings.ReplaceAll(path, "\\", "/")
	if !strings.HasPrefix(clean, "file:") {
		clean = "file:" + clean
	}
	if mmapSize <= 0 {
		mmapSize = 268435456 // 256MB default mmap window
	}
	return fmt.Sprintf("%s?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=%d", clean, mmapSize)
}
