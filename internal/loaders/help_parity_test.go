// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/render"
)

func TestHelpToolParity(t *testing.T) {
	help := render.RenderHelp(true)
	start := strings.Index(help, "Supported Tools:")
	end := strings.Index(help[start:], "Documentation & guides:")
	if start < 0 || end < 0 {
		t.Fatalf("help supported-tools section missing")
	}
	section := strings.ToLower(help[start : start+end])

	aliases := map[string]string{
		"mimocode":     "mimo",
		"command-code": "cmd",
	}
	tools := AllTools()
	if len(tools) != 15 {
		t.Fatalf("expected 15 registered tools, got %d", len(tools))
	}
	for tool := range tools {
		alias := string(tool)
		if mapped, ok := aliases[alias]; ok {
			alias = mapped
		}
		if !strings.Contains(section, alias) {
			t.Errorf("help is missing registered tool alias %q (tool %q)", alias, tool)
		}
	}
}
