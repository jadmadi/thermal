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
	Loader     func(string) (thermal.Summary, []thermal.DailyRow, error)
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
	}
}

func grokHomeDir(home string) string {
	if env := os.Getenv("GROK_HOME"); env != "" {
		return env
	}
	return filepath.Join(home, ".grok")
}

var toolAliases = map[string]thermal.Tool{
	"mimo":         thermal.ToolMiMoCode,
	"mimo-":        thermal.ToolMiMoCode,
	"mimocode":     thermal.ToolMiMoCode,
	"oc":           thermal.ToolOpenCode,
	"opencode":     thermal.ToolOpenCode,
	"codex":        thermal.ToolCodex,
	"devin":        thermal.ToolDevin,
	"agy":          thermal.ToolAgy,
	"cmd":          thermal.ToolCommandCode,
	"cc":           thermal.ToolCommandCode,
	"commandcode":  thermal.ToolCommandCode,
	"command-code": thermal.ToolCommandCode,
	"whale":        thermal.ToolCodewhale,
	"codewhale":    thermal.ToolCodewhale,
	"zc":           thermal.ToolZCode,
	"zcode":        thermal.ToolZCode,
	"grok":         thermal.ToolGrok,
	"muse":         thermal.ToolMuse,
	"all":          thermal.ToolAll,
	"auto":         thermal.ToolAuto,
}

func ResolveTool(name string) (thermal.Tool, bool) {
	if t, ok := toolAliases[strings.ToLower(name)]; ok {
		return t, true
	}
	return "", false
}

func LoadToolData(t thermal.Tool, info ToolInfo, dbPath string) (thermal.Summary, []thermal.DailyRow, string, error) {
	switch t {
	case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin, thermal.ToolZCode, thermal.ToolMuse:
		p := dbPath
		if p == "" {
			p = info.DBPath
		}
		if p == "" {
			return thermal.Summary{}, nil, "", fmt.Errorf("no database path configured for %s", info.Name)
		}
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return thermal.Summary{}, nil, "", fmt.Errorf("database not found: %s", p)
		}
		s, d, err := info.Loader(p)
		s.Tool = info.Name
		return s, d, info.DBPath, err
	default:
		dir := info.DataDir
		if dbPath != "" {
			dir = dbPath
		}
		if dir == "" {
			return thermal.Summary{}, nil, "", fmt.Errorf("no data directory configured for %s", info.Name)
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return thermal.Summary{}, nil, "", fmt.Errorf("data directory not found: %s", dir)
		}
		s, d, err := info.Loader(dir)
		s.Tool = info.Name
		dataPath := filepath.Join(dir, info.DataSubdir)
		if info.DataSubdir == "" {
			dataPath = filepath.Join(dir, "history.jsonl")
		}
		return s, d, dataPath, err
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
			fmt.Fprintln(os.Stderr, "  all                  Show leaderboard (default)")
			os.Exit(1)
		}
		return t
	}

	tools := AllTools()
	for _, t := range []thermal.Tool{thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolCodex, thermal.ToolAgy, thermal.ToolCommandCode, thermal.ToolCodewhale, thermal.ToolZCode, thermal.ToolGrok, thermal.ToolMuse} {
		info := tools[t]
		if info.DBPath != "" {
			if _, err := os.Stat(info.DBPath); err == nil {
				return t
			}
		}
		if info.DataDir != "" {
			if _, err := os.Stat(info.DataDir); err == nil {
				return t
			}
		}
	}

	fmt.Fprintf(os.Stderr, "thermal: no supported tool data found\n")
	os.Exit(1)
	return ""
}
