// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"testing"
)

func TestModelName_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Canonical exact hits
		{"Canonical Claude 3.5 Sonnet", "claude-3.5-sonnet", "claude-3.5-sonnet"},
		{"Canonical DeepSeek V3", "deepseek-v3", "deepseek-v3"},
		{"Canonical DeepSeek R1", "deepseek-r1", "deepseek-r1"},
		{"Canonical GPT-4o", "gpt-4o", "gpt-4o"},
		{"Canonical Gemini 2.0 Flash", "gemini-2.0-flash", "gemini-2.0-flash"},
		{"Canonical GLM 5.3 Flash", "glm-5.3-flash", "glm-5.3-flash"},

		// Case-insensitive normalization
		{"Case upper GLM", "GLM-5.3-Flash", "glm-5.3-flash"},
		{"Case mixed DeepSeek", "DeepSeek-Chat", "deepseek-v3"},
		{"Case mixed Claude", "Claude-3.5-Sonnet", "claude-3.5-sonnet"},
		{"Case mixed GPT", "GPT-4O", "gpt-4o"},

		// Whitespace trimming
		{"Whitespace padding", "  gpt-4o  \n", "gpt-4o"},
		{"Empty string", "", ""},
		{"Whitespace only", "   \t  ", ""},

		// Router & gateway prefix stripping
		{"OpenRouter Anthropic prefix", "openrouter/anthropic/claude-3.5-sonnet", "claude-3.5-sonnet"},
		{"OpenRouter DeepSeek prefix", "openrouter/deepseek/deepseek-chat", "deepseek-v3"},
		{"OpenRouter dated suffix", "openrouter/anthropic/claude-3-5-sonnet-20241022", "claude-3.5-sonnet"},
		{"LiteLLM prefix", "litellm/openai/gpt-4o", "gpt-4o"},
		{"Proxy prefix", "proxy/google/gemini-2.0-flash", "gemini-2.0-flash"},
		{"Provider slash only", "deepseek/deepseek-chat", "deepseek-v3"},

		// Date-stamped variants
		{"Claude 3.7 Sonnet date", "claude-3-7-sonnet-20250219", "claude-3.7-sonnet"},
		{"Claude 3.5 Sonnet date", "claude-3-5-sonnet-20241022", "claude-3.5-sonnet"},
		{"Claude 3.5 Haiku date", "claude-3-5-haiku-20241022", "claude-3.5-haiku"},
		{"GPT-4o ISO date", "gpt-4o-2024-08-06", "gpt-4o"},
		{"GPT-4o-mini ISO date", "gpt-4o-mini-2024-07-18", "gpt-4o-mini"},
		{"o1 date", "o1-2024-12-17", "o1"},
		{"o3-mini date", "o3-mini-2025-01-31", "o3-mini"},

		// Latest & tier aliases
		{"ChatGPT latest", "chatgpt-4o-latest", "gpt-4o"},
		{"Claude Sonnet latest", "claude-3-7-sonnet-latest", "claude-3.7-sonnet"},
		{"DeepSeek Chat alias", "deepseek-chat", "deepseek-v3"},
		{"DeepSeek Reasoner alias", "deepseek-reasoner", "deepseek-r1"},
		{"Gemini 3 Pro High tier", "gemini-3-pro-high", "gemini-3-pro-preview"},
		{"Gemini 2 Flash 001", "gemini-2.0-flash-001", "gemini-2.0-flash"},
		{"Gemini 3.8 Flash High", "Gemini 3.8 Flash (High)", "gemini-3.8-flash"},
		{"Gemini 3.8 Flash Medium", "Gemini 3.8 Flash (Medium)", "gemini-3.8-flash"},
		{"Gemini 3.8 Flash Low", "Gemini 3.8 Flash (Low)", "gemini-3.8-flash"},
		{"Gemini 3.7 Flash High", "Gemini 3.7 Flash (High)", "gemini-3.7-flash"},
		{"Gemini 3.6 Flash High", "Gemini 3.6 Flash (High)", "gemini-3.6-flash"},
		{"Gemini 3.1 Pro High", "Gemini 3.1 Pro (High)", "gemini-3.1-pro"},
		{"Meta LLaMA 3.3 70B", "meta-llama/llama-3.3-70b-instruct", "llama-3.3-70b-instruct"},

		// Backward-compatible fallback for uncataloged custom models
		{"Uncataloged custom model", "my-custom-fine-tuned-model", "my-custom-fine-tuned-model"},
		{"Uncataloged router path", "openrouter/custom/finetune-v1", "finetune-v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := modelName(tt.input)
			if got != tt.expected {
				t.Errorf("modelName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
			if exported := CanonicalModelName(tt.input); exported != tt.expected {
				t.Errorf("CanonicalModelName(%q) = %q; want %q", tt.input, exported, tt.expected)
			}
		})
	}
}

func BenchmarkModelName_Canonical(b *testing.B) {
	input := "claude-3.5-sonnet"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := modelName(input)
		if res != "claude-3.5-sonnet" {
			b.Fatalf("unexpected %s", res)
		}
	}
}

func BenchmarkModelName_Alias(b *testing.B) {
	input := "deepseek-chat"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := modelName(input)
		if res != "deepseek-v3" {
			b.Fatalf("unexpected %s", res)
		}
	}
}

func BenchmarkModelName_Router(b *testing.B) {
	input := "openrouter/anthropic/claude-3-5-sonnet-20241022"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := modelName(input)
		if res != "claude-3.5-sonnet" {
			b.Fatalf("unexpected %s", res)
		}
	}
}
