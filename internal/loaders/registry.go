// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

type ToolInfo struct {
	DBPath     string
	DataDir    string
	Name       string
	DataSubdir string // e.g. "history.jsonl", "brain", "sessions", "projects"
	Loader     func(string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error)
}

// ToolData is the loader output for one tool: the lifetime summary, per-day
// rows, and per-project rows when the tool records where a session ran.
type ToolData struct {
	Summary  thermal.Summary
	Daily    []thermal.DailyRow
	Projects []thermal.ProjectDay
	Path     string
}

func AllTools() map[thermal.Tool]ToolInfo {
	home := thermal.HomeDir()
	return map[thermal.Tool]ToolInfo{
		thermal.ToolMiMoCode: {
			DBPath:  filepath.Join(home, ".local", "share", "mimocode", "mimocode.db"),
			DataDir: filepath.Join(home, ".local", "share", "mimocode"),
			Name:    "MiMoCode",
			Loader:  LoadMiMoCodeData,
		},
		thermal.ToolOpenCode: {
			DBPath:  filepath.Join(home, ".local", "share", "opencode", "opencode.db"),
			DataDir: filepath.Join(home, ".local", "share", "opencode"),
			Name:    "OpenCode",
			Loader:  LoadOpenCodeData,
		},
		thermal.ToolCodex: {
			DataDir:    filepath.Join(home, ".codex"),
			Name:       "Codex",
			DataSubdir: "sessions",
			Loader:     LoadCodexData,
		},
		thermal.ToolDevin: {
			DBPath:  filepath.Join(home, ".local", "share", "devin", "cli", "sessions.db"),
			DataDir: filepath.Join(home, ".local", "share", "devin", "cli"),
			Name:    "Devin",
			Loader:  LoadDevinData,
		},
		thermal.ToolAgy: {
			DataDir:    filepath.Join(home, ".gemini", "antigravity-cli"),
			Name:       "Agy",
			DataSubdir: "brain",
			Loader:     LoadAgyData,
		},
		thermal.ToolCommandCode: {
			DataDir:    filepath.Join(home, ".commandcode"),
			Name:       "command-code",
			DataSubdir: "projects",
			Loader:     LoadCommandCodeData,
		},
		thermal.ToolCodewhale: {
			DataDir:    filepath.Join(home, ".codewhale"),
			Name:       "codewhale",
			DataSubdir: "sessions",
			Loader:     LoadCodewhaleData,
		},
		thermal.ToolZCode: {
			DBPath:  filepath.Join(home, ".zcode", "cli", "db", "db.sqlite"),
			DataDir: filepath.Join(home, ".zcode"),
			Name:    "ZCode",
			Loader:  LoadZCodeData,
		},
		thermal.ToolGrok: {
			DataDir:    grokHomeDir(home),
			Name:       "Grok",
			DataSubdir: "sessions",
			Loader:     LoadGrokData,
		},
		thermal.ToolMuse: {
			DBPath:  filepath.Join(home, ".local", "share", "muse", "session-index.db"),
			DataDir: filepath.Join(home, ".local", "share", "muse"),
			Name:    "Muse",
			Loader:  LoadMuseData,
		},
		thermal.ToolClaude: {
			DataDir:    filepath.Join(home, ".claude"),
			Name:       "Claude",
			DataSubdir: "projects",
			Loader:     LoadClaudeData,
		},
		thermal.ToolDroid: {
			DataDir:    filepath.Join(home, ".factory"),
			Name:       "Droid",
			DataSubdir: "sessions",
			Loader:     LoadDroidData,
		},
		thermal.ToolDsh: {
			DataDir:    dshHomeDir(home),
			Name:       "DeepSeek (DSH)",
			DataSubdir: "storages",
			Loader:     LoadDshData,
		},
		thermal.ToolHermes: {
			DBPath:  filepath.Join(hermesHomeDir(home), "state.db"),
			DataDir: hermesHomeDir(home),
			Name:    "Nous Hermes",
			Loader:  LoadHermesData,
		},
	}
}

func hermesHomeDir(home string) string {
	if env := os.Getenv("HERMES_HOME"); env != "" {
		return env
	}
	return filepath.Join(home, ".hermes")
}

func dshHomeDir(home string) string {
	if env := os.Getenv("DSH_HOME"); env != "" {
		return env
	}
	return filepath.Join(home, ".dsh")
}

func grokHomeDir(home string) string {
	if env := os.Getenv("GROK_HOME"); env != "" {
		return env
	}
	return filepath.Join(home, ".grok")
}

var toolAliases = map[string]thermal.Tool{
	"mimo":             thermal.ToolMiMoCode,
	"mimo-":            thermal.ToolMiMoCode,
	"mimocode":         thermal.ToolMiMoCode,
	"oc":               thermal.ToolOpenCode,
	"opencode":         thermal.ToolOpenCode,
	"codex":            thermal.ToolCodex,
	"devin":            thermal.ToolDevin,
	"agy":              thermal.ToolAgy,
	"cmd":              thermal.ToolCommandCode,
	"commandcode":      thermal.ToolCommandCode,
	"command-code":     thermal.ToolCommandCode,
	"whale":            thermal.ToolCodewhale,
	"codewhale":        thermal.ToolCodewhale,
	"zc":               thermal.ToolZCode,
	"zcode":            thermal.ToolZCode,
	"grok":             thermal.ToolGrok,
	"muse":             thermal.ToolMuse,
	"claude":           thermal.ToolClaude,
	"ccode":            thermal.ToolClaude,
	"droid":            thermal.ToolDroid,
	"factory":          thermal.ToolDroid,
	"dsh":              thermal.ToolDsh,
	"deepseek":         thermal.ToolDsh,
	"deepseek-harness": thermal.ToolDsh,
	"hermes":           thermal.ToolHermes,
	"nous":             thermal.ToolHermes,
	"nous-hermes":      thermal.ToolHermes,
	"all":              thermal.ToolAll,
	"auto":             thermal.ToolAuto,
}

func ResolveTool(name string) (thermal.Tool, bool) {
	if t, ok := toolAliases[strings.ToLower(name)]; ok {
		return t, true
	}
	return "", false
}

func LoadToolData(t thermal.Tool, info ToolInfo, dbPath string) (ToolData, error) {
	switch t {
	case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin, thermal.ToolZCode, thermal.ToolMuse, thermal.ToolHermes:
		p := dbPath
		if p == "" {
			p = info.DBPath
		}
		if p == "" {
			return ToolData{}, fmt.Errorf("no database path configured for %s", info.Name)
		}
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return ToolData{}, fmt.Errorf("database not found: %s", p)
		}
		s, d, pr, err := info.Loader(p)
		s.Tool = info.Name
		return ToolData{Summary: s, Daily: d, Projects: pr, Path: info.DBPath}, err
	default:
		dir := info.DataDir
		if dbPath != "" {
			dir = dbPath
		}
		if dir == "" {
			return ToolData{}, fmt.Errorf("no data directory configured for %s", info.Name)
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return ToolData{}, fmt.Errorf("data directory not found: %s", dir)
		}
		s, d, pr, err := info.Loader(dir)
		s.Tool = info.Name
		dataPath := filepath.Join(dir, info.DataSubdir)
		if info.DataSubdir == "" {
			dataPath = filepath.Join(dir, "history.jsonl")
		}
		return ToolData{Summary: s, Daily: d, Projects: pr, Path: dataPath}, err
	}
}

func DetectTool(name string) thermal.Tool {
	if name != "auto" && name != "all" {
		t, ok := ResolveTool(name)
		if !ok {
			fmt.Fprintf(os.Stderr, "thermal: unknown tool: %s\n\n", name)
			fmt.Fprintln(os.Stderr, "Available tools:")
			fmt.Fprintln(os.Stderr, "  mimocode, mimo       MiMoCode")
			fmt.Fprintln(os.Stderr, "  opencode, oc         OpenCode")
			fmt.Fprintln(os.Stderr, "  codex                Codex CLI")
			fmt.Fprintln(os.Stderr, "  devin                Devin")
			fmt.Fprintln(os.Stderr, "  agy                  Agy (Antigravity)")
			fmt.Fprintln(os.Stderr, "  command-code, cmd    command-code-ai")
			fmt.Fprintln(os.Stderr, "  codewhale, whale     codewhale")
			fmt.Fprintln(os.Stderr, "  zcode, zc            ZCode")
			fmt.Fprintln(os.Stderr, "  grok                 Grok CLI")
			fmt.Fprintln(os.Stderr, "  muse                 Muse")
			fmt.Fprintln(os.Stderr, "  claude               Claude Code")
			fmt.Fprintln(os.Stderr, "  droid                Droid (Factory)")
			fmt.Fprintln(os.Stderr, "  dsh, deepseek        DeepSeek (DSH)")
			fmt.Fprintln(os.Stderr, "  hermes, nous         Nous Hermes")
			fmt.Fprintln(os.Stderr, "  all                  Show leaderboard (default)")
			os.Exit(1)
		}
		return t
	}

	tools := AllTools()
	for _, t := range []thermal.Tool{thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolCodex, thermal.ToolDevin, thermal.ToolAgy, thermal.ToolCommandCode, thermal.ToolCodewhale, thermal.ToolZCode, thermal.ToolGrok, thermal.ToolMuse, thermal.ToolClaude, thermal.ToolDroid, thermal.ToolDsh, thermal.ToolHermes} {
		info := tools[t]
		if info.DBPath != "" {
			if _, err := os.Stat(info.DBPath); err == nil {
				return t
			}
		} else if info.DataDir != "" {
			if _, err := os.Stat(info.DataDir); err == nil {
				return t
			}
		}
	}

	fmt.Fprintf(os.Stderr, "thermal: no supported tool data found\n")
	os.Exit(1)
	return ""
}
