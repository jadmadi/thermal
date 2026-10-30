// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

// PathShape describes what --db accepts for a tool.
type PathShape int

const (
	// ShapeSQLiteFile accepts a SQLite file, or a directory containing the
	// tool's expected database filename.
	ShapeSQLiteFile PathShape = iota
	// ShapeDirectory accepts a data directory.
	ShapeDirectory
	// ShapeEither accepts a SQLite file or a data directory.
	ShapeEither
)

// DataCondition is the user-facing category of a data-access failure.
type DataCondition string

const (
	ConditionMissing    DataCondition = "missing"
	ConditionUnreadable DataCondition = "unreadable"
	ConditionWrongShape DataCondition = "wrong-shape"
	ConditionNotDir     DataCondition = "not-dir"
	ConditionLoop       DataCondition = "loop"
	ConditionOther      DataCondition = "other"
)

// DBPathError is a user-facing data-access or --db validation failure. The
// raw OS/driver error stays in Err for --verbose.
type DBPathError struct {
	Tool        string
	ToolID      string
	Path        string
	DisplayPath string
	Condition   DataCondition
	Message     string
	Fix         string
	Next        []string
	Err         error
}

func (e *DBPathError) Error() string {
	var b strings.Builder
	b.WriteString(e.Message)
	for _, line := range e.Next {
		b.WriteString("\n  ")
		b.WriteString(line)
	}
	if e.Fix != "" {
		b.WriteString("\n  ")
		b.WriteString(e.Fix)
	}
	return b.String()
}

func (e *DBPathError) Unwrap() error { return e.Err }

// displayDataPath shortens a path under the user's home for human output.
func displayDataPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(os.PathSeparator)) {
		return "~" + path[len(home):]
	}
	return path
}

// classifyStatError maps an OS error to a condition, message, and remediation.
func classifyStatError(path string, err error) *DBPathError {
	display := displayDataPath(path)
	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "too many levels of symbolic links") ||
		strings.Contains(lower, "too many symbolic links"):
		return &DBPathError{
			Path:        path,
			DisplayPath: display,
			Condition:   ConditionLoop,
			Message:     fmt.Sprintf("too many levels of symbolic links at %s", display),
			Fix:         "resolve it with readlink -f " + display + " and pass the resolved path",
			Err:         err,
		}
	case os.IsPermission(err):
		mode := ""
		if fi, statErr := os.Lstat(path); statErr == nil {
			mode = fmt.Sprintf(" (mode %04o)", fi.Mode().Perm())
		}
		return &DBPathError{
			Path:        path,
			DisplayPath: display,
			Condition:   ConditionUnreadable,
			Message:     fmt.Sprintf("permission denied reading %s%s", display, mode),
			Fix:         "check the mode with ls -ld " + display + " and grant read access, or omit --db",
			Err:         err,
		}
	case strings.Contains(lower, "not a directory"):
		return &DBPathError{
			Path:        path,
			DisplayPath: display,
			Condition:   ConditionNotDir,
			Message:     fmt.Sprintf("%s is not a directory", display),
			Fix:         "check the path components or omit --db",
			Err:         err,
		}
	default:
		return &DBPathError{
			Path:        path,
			DisplayPath: display,
			Condition:   ConditionOther,
			Message:     fmt.Sprintf("cannot read %s: %v", display, err),
			Fix:         "run thermal --verbose for the driver error",
			Err:         err,
		}
	}
}

// MissingDefaultDataError names a tool whose default path does not exist and
// gives the inventory, all-tools report, and --db override.
func MissingDefaultDataError(t thermal.Tool, displayName, path string, err error) *DBPathError {
	display := displayDataPath(path)
	return &DBPathError{
		Tool:        displayName,
		ToolID:      string(t),
		Path:        path,
		DisplayPath: display,
		Condition:   ConditionMissing,
		Message:     fmt.Sprintf("%s has no data at %s", displayName, display),
		Next: []string{
			"inventory: thermal sources",
			"all tools: thermal",
			"override:  thermal --tool " + string(t) + " --db <path>",
		},
		Err: err,
	}
}

// EnrichDefaultDataError makes the all-tools line name the report the caller
// asked for, when one is known.
func EnrichDefaultDataError(err error, report string) error {
	var da *DBPathError
	if !errors.As(err, &da) || da.Condition != ConditionMissing || da.ToolID == "" {
		return err
	}
	all := "all tools: thermal"
	if report != "" && report != "all" {
		all += " " + report
	}
	da.Next = []string{
		"inventory: thermal sources",
		all,
		"override:  thermal --tool " + da.ToolID + " --db <path>",
	}
	return err
}

var toolDisplayNames = map[thermal.Tool]string{
	thermal.ToolMiMoCode:    "MiMoCode",
	thermal.ToolOpenCode:    "OpenCode",
	thermal.ToolDevin:       "Devin",
	thermal.ToolZCode:       "ZCode",
	thermal.ToolHermes:      "Nous Hermes",
	thermal.ToolZed:         "Zed",
	thermal.ToolCodex:       "Codex",
	thermal.ToolMuse:        "Muse",
	thermal.ToolClaude:      "Claude",
	thermal.ToolCommandCode: "CMD",
	thermal.ToolAgy:         "Agy",
	thermal.ToolDroid:       "Droid",
	thermal.ToolCodewhale:   "codewhale",
	thermal.ToolDsh:         "DeepSeek (DSH)",
	thermal.ToolGrok:        "Grok",
}

func toolDisplayName(t thermal.Tool) string {
	if name, ok := toolDisplayNames[t]; ok {
		return name
	}
	return string(t)
}

// ToolPathShape reports what --db accepts for a tool. Codex's loader consumes
// a data directory, so it is directory-shaped even though it also reads a
// state database internally. Muse accepts either a session-index file or a
// directory.
func ToolPathShape(t thermal.Tool) PathShape {
	switch t {
	case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin,
		thermal.ToolZCode, thermal.ToolHermes, thermal.ToolZed:
		return ShapeSQLiteFile
	case thermal.ToolMuse:
		return ShapeEither
	default:
		return ShapeDirectory
	}
}

// ExpectedDBFile returns the standard SQLite filename for a file-shaped tool.
func ExpectedDBFile(t thermal.Tool) string {
	switch t {
	case thermal.ToolOpenCode:
		return "opencode.db"
	case thermal.ToolMiMoCode:
		return "mimocode.db"
	case thermal.ToolDevin:
		return "sessions.db"
	case thermal.ToolZCode:
		return "db.sqlite"
	case thermal.ToolHermes:
		return "state.db"
	case thermal.ToolZed:
		return "threads.db"
	case thermal.ToolMuse:
		return "session-index.db"
	}
	return ""
}

// ValidateDBPath resolves and validates a user-supplied --db path. It returns
// the canonical path (a file or directory) and never falls back to a default.
func ValidateDBPath(t thermal.Tool, path string) (string, error) {
	display := toolDisplayName(t)
	path = strings.TrimSpace(path)
	if path == "" {
		return "", &DBPathError{
			Tool:    display,
			Message: "--db requires a non-empty path",
			Fix:     "example: thermal --tool " + string(t) + " --db <path> daily",
		}
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	fi, err := os.Stat(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return "", &DBPathError{
				Tool:        display,
				ToolID:      string(t),
				Path:        resolved,
				DisplayPath: displayDataPath(resolved),
				Condition:   ConditionMissing,
				Message:     fmt.Sprintf("--db path does not exist: %s", resolved),
				Fix:         "check the path or omit --db",
				Err:         err,
			}
		}
		classified := classifyStatError(resolved, err)
		classified.Tool = display
		classified.ToolID = string(t)
		return "", classified
	}

	shape := ToolPathShape(t)
	switch shape {
	case ShapeSQLiteFile:
		if fi.IsDir() {
			expected := ExpectedDBFile(t)
			candidate := filepath.Join(resolved, expected)
			cfi, cerr := os.Stat(candidate)
			if cerr != nil || cfi.IsDir() {
				return "", &DBPathError{
					Tool:    display,
					Path:    resolved,
					Message: fmt.Sprintf("%s is a directory; %s expects %s inside it", resolved, display, expected),
					Fix:     fmt.Sprintf("pass the database file directly: --db %s", candidate),
				}
			}
			resolved = candidate
			fi = cfi
		}
		if !fi.Mode().IsRegular() {
			return "", &DBPathError{
				Tool:    display,
				Path:    resolved,
				Message: fmt.Sprintf("%s is not a regular file", resolved),
			}
		}
		if fi.Size() == 0 {
			return "", &DBPathError{
				Tool:    display,
				Path:    resolved,
				Message: fmt.Sprintf("%s is empty; expected a SQLite database", resolved),
			}
		}
		if err := checkSQLiteHeader(resolved); err != nil {
			return "", &DBPathError{
				Tool:    display,
				Path:    resolved,
				Message: fmt.Sprintf("%s is not a SQLite database for %s", resolved, display),
				Fix:     "omit --db or point it at " + ExpectedDBFile(t),
				Err:     err,
			}
		}
	case ShapeDirectory:
		if !fi.IsDir() {
			return "", &DBPathError{
				Tool:    display,
				Path:    resolved,
				Message: fmt.Sprintf("%s is a file; %s expects a directory", resolved, display),
				Fix:     "pass the data directory, not a single file",
			}
		}
	case ShapeEither:
		if !fi.IsDir() {
			if !fi.Mode().IsRegular() {
				return "", &DBPathError{
					Tool:    display,
					Path:    resolved,
					Message: fmt.Sprintf("%s is not a regular file", resolved),
				}
			}
			if fi.Size() == 0 {
				return "", &DBPathError{
					Tool:    display,
					Path:    resolved,
					Message: fmt.Sprintf("%s is empty; expected a SQLite database", resolved),
				}
			}
			if err := checkSQLiteHeader(resolved); err != nil {
				return "", &DBPathError{
					Tool:    display,
					Path:    resolved,
					Message: fmt.Sprintf("%s is not a SQLite database for %s", resolved, display),
					Fix:     "omit --db or pass the data directory",
					Err:     err,
				}
			}
		}
	}

	f, err := os.Open(resolved)
	if err != nil {
		classified := classifyStatError(resolved, err)
		classified.Tool = display
		classified.ToolID = string(t)
		return "", classified
	}
	_ = f.Close()
	return resolved, nil
}

func checkSQLiteHeader(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, 16)
	if _, err := f.Read(buf); err != nil {
		return err
	}
	if !bytes.Equal(buf, []byte("SQLite format 3\x00")) {
		return fmt.Errorf("missing SQLite header")
	}
	return nil
}

// translateDBError turns a raw driver/loader failure on a --db path into a
// user-facing DBPathError. The raw error remains available through Unwrap.
func translateDBError(t thermal.Tool, path string, err error) error {
	if err == nil {
		return nil
	}
	var existing *DBPathError
	if errors.As(err, &existing) {
		return err
	}
	display := toolDisplayName(t)
	msg := err.Error()
	expected := ExpectedDBFile(t)
	fix := ""
	if expected != "" {
		fix = "omit --db or point it at " + expected
	}

	switch {
	case strings.Contains(msg, "file is not a database"):
		return &DBPathError{
			Tool:    display,
			Path:    path,
			Message: fmt.Sprintf("%s is not a SQLite database for %s", path, display),
			Fix:     fix,
			Err:     err,
		}
	case strings.Contains(msg, "disk I/O error"):
		if fi, statErr := os.Stat(path); statErr == nil && fi.IsDir() {
			return &DBPathError{
				Tool:    display,
				Path:    path,
				Message: fmt.Sprintf("%s is a directory; %s expects %s inside it", path, display, expected),
				Fix:     fmt.Sprintf("pass the database file directly: --db %s", filepath.Join(path, expected)),
				Err:     err,
			}
		}
		return &DBPathError{
			Tool:    display,
			Path:    path,
			Message: fmt.Sprintf("cannot read %s; check permissions", path),
			Err:     err,
		}
	case strings.Contains(msg, "unable to open database file"):
		return &DBPathError{
			Tool:    display,
			Path:    path,
			Message: fmt.Sprintf("cannot open %s for %s; check permissions and that the file exists", path, display),
			Err:     err,
		}
	case strings.Contains(msg, "database is locked"):
		return &DBPathError{
			Tool:    display,
			Path:    path,
			Message: fmt.Sprintf("%s is locked by another process; Thermal opens databases read-only, so retry when the writer is idle", path),
			Err:     err,
		}
	case strings.Contains(msg, "no such table"):
		return &DBPathError{
			Tool:    display,
			Path:    path,
			Message: fmt.Sprintf("%s has no expected table for %s; it may be from an unsupported version", path, display),
			Err:     err,
		}
	default:
		return &DBPathError{
			Tool:    display,
			Path:    path,
			Message: fmt.Sprintf("cannot open %s for %s", path, display),
			Err:     err,
		}
	}
}
