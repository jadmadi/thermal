// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

// DefaultConfigPaths returns the candidate paths for Thermal user configuration,
// prioritizing THERMAL_CONFIG env, then ~/.config/thermal/thermal.config,
// and ~/.config/thermal/config.json.
func DefaultConfigPaths(home string) []string {
	if env := strings.TrimSpace(os.Getenv("THERMAL_CONFIG")); env != "" {
		if strings.HasPrefix(env, "~/") {
			env = filepath.Join(home, env[2:])
		}
		return []string{filepath.Clean(env)}
	}
	return []string{
		filepath.Join(home, ".config", "thermal", "thermal.config"),
		filepath.Join(home, ".config", "thermal", "config.json"),
	}
}

// UserConfig represents custom tool path configurations.
type UserConfig struct {
	Paths map[string]string `json:"paths"`
}

// LoadUserConfig reads custom tool path overrides from ~/.config/thermal/thermal.config
// or ~/.config/thermal/config.json. Supports both JSON and simple INI/key-value formats.
func LoadUserConfig(home string) map[thermal.Tool]string {
	overrides := make(map[thermal.Tool]string)
	candidates := DefaultConfigPaths(home)

	for _, cand := range candidates {
		data, err := os.ReadFile(cand)
		if err != nil {
			continue
		}

		rawMap := parseConfigBytes(data)
		for rawKey, rawPath := range rawMap {
			rawPath = strings.TrimSpace(rawPath)
			if rawPath == "" {
				continue
			}
			// Expand tilde
			if strings.HasPrefix(rawPath, "~/") {
				rawPath = filepath.Join(home, rawPath[2:])
			}
			rawPath = filepath.Clean(rawPath)

			// Resolve tool alias
			if tool, ok := ResolveTool(rawKey); ok {
				overrides[tool] = rawPath
			}
		}
		if len(overrides) > 0 {
			break
		}
	}

	return overrides
}

// parseConfigBytes attempts JSON parsing first, falling back to simple INI/key-value parsing.
func parseConfigBytes(data []byte) map[string]string {
	res := make(map[string]string)
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return res
	}

	// 1. Try structured JSON with "paths": { ... }
	if trimmed[0] == '{' {
		var structured UserConfig
		if err := json.Unmarshal(trimmed, &structured); err == nil && len(structured.Paths) > 0 {
			return structured.Paths
		}
		// 2. Try flat JSON { "tool": "/path" }
		var flat map[string]string
		if err := json.Unmarshal(trimmed, &flat); err == nil && len(flat) > 0 {
			return flat
		}
	}

	// 3. Fallback: Line-based INI / key-value format
	// Supports:
	//   [paths]
	//   agy = ~/.gemini/antigravity
	//   opencode: ~/.local/share/opencode/opencode.db
	scanner := bufio.NewScanner(bytes.NewReader(trimmed))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		// Skip INI section headers like [paths] or [tools]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			continue
		}

		// Split on '=' or ':'
		sepIdx := strings.IndexAny(line, "=:")
		if sepIdx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:sepIdx])
		val := strings.TrimSpace(line[sepIdx+1:])

		// Strip optional surrounding quotes
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		val = strings.TrimSpace(val)

		if key != "" && val != "" {
			res[key] = val
		}
	}

	return res
}
