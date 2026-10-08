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
	tools := map[thermal.Tool]ToolInfo{
		thermal.ToolMiMoCode: func() ToolInfo {
			db, dir := probePlatformDataPath(home, "mimocode", "mimocode.db")
			return ToolInfo{
				DBPath:  db,
				DataDir: dir,
				Name:    "MiMoCode",
				Loader:  LoadMiMoCodeData,
			}
		}(),
		thermal.ToolOpenCode: func() ToolInfo {
			db, dir := probePlatformDataPath(home, "opencode", "opencode.db")
			return ToolInfo{
				DBPath:  db,
				DataDir: dir,
				Name:    "OpenCode",
				Loader:  LoadOpenCodeData,
			}
		}(),
		thermal.ToolCodex: {
			DataDir:    filepath.Join(home, ".codex"),
			Name:       "Codex",
			DataSubdir: "sessions",
			Loader:     LoadCodexData,
		},
		thermal.ToolDevin: func() ToolInfo {
			db, dir := probePlatformDataPath(home, filepath.Join("devin", "cli"), "sessions.db")
			return ToolInfo{
				DBPath:  db,
				DataDir: dir,
				Name:    "Devin",
				Loader:  LoadDevinData,
			}
		}(),
		thermal.ToolAgy: {
			DataDir:    agyHomeDir(home),
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
		thermal.ToolMuse: func() ToolInfo {
			db, dir := probePlatformDataPath(home, "muse", "session-index.db")
			return ToolInfo{
				DBPath:  db,
				DataDir: dir,
				Name:    "Muse",
				Loader:  LoadMuseData,
			}
		}(),
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
		thermal.ToolZed: func() ToolInfo {
			db, dir := probePlatformDataPath(home, filepath.Join("zed", "threads"), "threads.db")
			return ToolInfo{
				DBPath:  db,
				DataDir: dir,
				Name:    "Zed",
				Loader:  LoadZedData,
			}
		}(),
	}

	// Apply user overrides from ~/.config/thermal/thermal.config or config.json
	for tool, customPath := range LoadUserConfig(home) {
		info, ok := tools[tool]
		if !ok || customPath == "" {
			continue
		}
		if fi, err := os.Stat(customPath); err == nil && !fi.IsDir() {
			// Direct file specified (e.g. SQLite database or single file)
			info.DBPath = customPath
			info.DataDir = filepath.Dir(customPath)
		} else {
			// Directory specified
			info.DataDir = customPath
			// For tools with standard DB filenames, probe inside the custom directory
			switch tool {
			case thermal.ToolOpenCode:
				cand := filepath.Join(customPath, "opencode.db")
				if _, err := os.Stat(cand); err == nil {
					info.DBPath = cand
				}
			case thermal.ToolMiMoCode:
				cand := filepath.Join(customPath, "mimocode.db")
				if _, err := os.Stat(cand); err == nil {
					info.DBPath = cand
				}
			case thermal.ToolDevin:
				cand := filepath.Join(customPath, "sessions.db")
				if _, err := os.Stat(cand); err == nil {
					info.DBPath = cand
				}
			case thermal.ToolHermes:
				cand := filepath.Join(customPath, "state.db")
				if _, err := os.Stat(cand); err == nil {
					info.DBPath = cand
				}
			case thermal.ToolZCode:
				cand := filepath.Join(customPath, "db.sqlite")
				if _, err := os.Stat(cand); err == nil {
					info.DBPath = cand
				}
			case thermal.ToolMuse:
				cand := filepath.Join(customPath, "session-index.db")
				if _, err := os.Stat(cand); err == nil {
					info.DBPath = cand
				}
			case thermal.ToolZed:
				cand := filepath.Join(customPath, "threads.db")
				if _, err := os.Stat(cand); err == nil {
					info.DBPath = cand
				}
			}
		}
		tools[tool] = info
	}

	return tools
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

func agyHomeDir(home string) string {
	for _, raw := range []string{os.Getenv("ANTIGRAVITY_APP_DATA_DIR"), os.Getenv("AGY_HOME")} {
		env := strings.TrimSpace(raw)
		if env == "" {
			continue
		}
		if strings.HasPrefix(env, "~/") {
			env = filepath.Join(home, env[2:])
		}
		if filepath.IsAbs(env) {
			if _, err := os.Stat(env); err == nil {
				return env
			}
			if _, err := os.Stat(filepath.Join(env, "brain")); err == nil {
				return env
			}
		}
		// Google Antigravity convention: relative app_data_dir names
		// represent subdirectories under ~/.gemini/ (e.g. "antigravity", "antigravity-cli", "antigravity-ide")
		geminiCand := filepath.Join(home, ".gemini", env)
		if _, err := os.Stat(geminiCand); err == nil {
			return geminiCand
		}
		if fi, err := os.Stat(env); err == nil && fi.IsDir() {
			if abs, err := filepath.Abs(env); err == nil {
				return abs
			}
			return env
		}
		if !strings.Contains(env, string(filepath.Separator)) {
			return geminiCand
		}
		return env
	}

	// Default auto-discovery across known Antigravity locations
	candidates := []string{
		filepath.Join(home, ".gemini", "antigravity"),
		filepath.Join(home, ".gemini", "antigravity-cli"),
		filepath.Join(home, ".gemini", "antigravity-ide"),
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		candidates = append(candidates, filepath.Join(appData, "Google", "Antigravity"))
	}
	for _, cand := range candidates {
		if _, err := os.Stat(filepath.Join(cand, "brain")); err == nil {
			return cand
		}
	}
	return filepath.Join(home, ".gemini", "antigravity")
}

// probePlatformDataPath probes candidate application data directories in order:
// 1. Unix XDG path: ~/.local/share/<appSubpath>
// 2. macOS Application Support: ~/Library/Application Support/<appSubpath>
// 3. Windows APPDATA / LOCALAPPDATA environment variables
// 4. Windows user profile AppData fallback: ~/AppData/Roaming and ~/AppData/Local
// It returns the first path whose dbName exists on disk, or the primary default.
func probePlatformDataPath(home string, appSubpath string, dbName string) (string, string) {
	primaryDir := filepath.Join(home, ".local", "share", appSubpath)
	primaryDB := filepath.Join(primaryDir, dbName)

	candidates := []string{
		primaryDir,
		filepath.Join(home, "Library", "Application Support", appSubpath),
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		candidates = append(candidates, filepath.Join(appData, appSubpath))
	}
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, appSubpath))
	}
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, "AppData", "Roaming", appSubpath),
			filepath.Join(home, "AppData", "Local", appSubpath),
		)
	}

	for _, dir := range candidates {
		candDB := filepath.Join(dir, dbName)
		if _, err := os.Stat(candDB); err == nil {
			return candDB, dir
		}
	}
	return primaryDB, primaryDir
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
	"antigravity":      thermal.ToolAgy,
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
	"zed":              thermal.ToolZed,
	"zed-editor":       thermal.ToolZed,
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
	case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin, thermal.ToolZCode, thermal.ToolMuse, thermal.ToolHermes, thermal.ToolZed:
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
			fmt.Fprintln(os.Stderr, "  agy, antigravity     Agy (Antigravity)")
			fmt.Fprintln(os.Stderr, "  command-code, cmd    command-code-ai")
			fmt.Fprintln(os.Stderr, "  codewhale, whale     codewhale")
			fmt.Fprintln(os.Stderr, "  zcode, zc            ZCode")
			fmt.Fprintln(os.Stderr, "  grok                 Grok CLI")
			fmt.Fprintln(os.Stderr, "  muse                 Muse")
			fmt.Fprintln(os.Stderr, "  claude               Claude Code")
			fmt.Fprintln(os.Stderr, "  droid                Droid (Factory)")
			fmt.Fprintln(os.Stderr, "  dsh, deepseek        DeepSeek (DSH)")
			fmt.Fprintln(os.Stderr, "  hermes, nous         Nous Hermes")
			fmt.Fprintln(os.Stderr, "  zed, zed-editor      Zed Editor")
			fmt.Fprintln(os.Stderr, "  all                  Show leaderboard (default)")
			os.Exit(1)
		}
		return t
	}

	tools := AllTools()
	for _, t := range []thermal.Tool{thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolCodex, thermal.ToolDevin, thermal.ToolAgy, thermal.ToolCommandCode, thermal.ToolCodewhale, thermal.ToolZCode, thermal.ToolGrok, thermal.ToolMuse, thermal.ToolClaude, thermal.ToolDroid, thermal.ToolDsh, thermal.ToolHermes, thermal.ToolZed} {
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
