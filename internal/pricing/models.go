// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package pricing

import "strings"

// ModelIdentity defines canonical model metadata.
type ModelIdentity struct {
	CanonicalID string `json:"canonicalId"`
	DisplayName string `json:"displayName"`
	Provider    string `json:"provider"`
	Family      string `json:"family"`
}

// CanonicalIdentities lists recognized model identities and their vendor families.
var CanonicalIdentities = []ModelIdentity{
	// Anthropic
	{CanonicalID: "claude-3.7-sonnet", DisplayName: "Claude 3.7 Sonnet", Provider: "anthropic", Family: "claude"},
	{CanonicalID: "claude-3.5-sonnet", DisplayName: "Claude 3.5 Sonnet", Provider: "anthropic", Family: "claude"},
	{CanonicalID: "claude-3.5-sonnet-v2", DisplayName: "Claude 3.5 Sonnet v2", Provider: "anthropic", Family: "claude"},
	{CanonicalID: "claude-3.5-haiku", DisplayName: "Claude 3.5 Haiku", Provider: "anthropic", Family: "claude"},
	{CanonicalID: "claude-3-opus", DisplayName: "Claude 3 Opus", Provider: "anthropic", Family: "claude"},
	{CanonicalID: "claude-3-sonnet", DisplayName: "Claude 3 Sonnet", Provider: "anthropic", Family: "claude"},
	{CanonicalID: "claude-3-haiku", DisplayName: "Claude 3 Haiku", Provider: "anthropic", Family: "claude"},

	// OpenAI
	{CanonicalID: "gpt-4o", DisplayName: "GPT-4o", Provider: "openai", Family: "gpt"},
	{CanonicalID: "gpt-4o-mini", DisplayName: "GPT-4o mini", Provider: "openai", Family: "gpt"},
	{CanonicalID: "gpt-4-turbo", DisplayName: "GPT-4 Turbo", Provider: "openai", Family: "gpt"},
	{CanonicalID: "o1", DisplayName: "o1", Provider: "openai", Family: "o-series"},
	{CanonicalID: "o1-preview", DisplayName: "o1-preview", Provider: "openai", Family: "o-series"},
	{CanonicalID: "o1-mini", DisplayName: "o1-mini", Provider: "openai", Family: "o-series"},
	{CanonicalID: "o3-mini", DisplayName: "o3-mini", Provider: "openai", Family: "o-series"},

	// DeepSeek
	{CanonicalID: "deepseek-v3", DisplayName: "DeepSeek V3", Provider: "deepseek", Family: "deepseek"},
	{CanonicalID: "deepseek-r1", DisplayName: "DeepSeek R1", Provider: "deepseek", Family: "deepseek"},
	{CanonicalID: "deepseek-coder", DisplayName: "DeepSeek Coder", Provider: "deepseek", Family: "deepseek"},

	// Google
	{CanonicalID: "gemini-2.5-pro", DisplayName: "Gemini 2.5 Pro", Provider: "google", Family: "gemini"},
	{CanonicalID: "gemini-2.5-flash", DisplayName: "Gemini 2.5 Flash", Provider: "google", Family: "gemini"},
	{CanonicalID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash", Provider: "google", Family: "gemini"},
	{CanonicalID: "gemini-2.0-pro", DisplayName: "Gemini 2.0 Pro", Provider: "google", Family: "gemini"},
	{CanonicalID: "gemini-1.5-pro", DisplayName: "Gemini 1.5 Pro", Provider: "google", Family: "gemini"},
	{CanonicalID: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash", Provider: "google", Family: "gemini"},

	// Zhipu
	{CanonicalID: "glm-5.3-flash", DisplayName: "GLM 5.3 Flash", Provider: "zhipuai", Family: "glm"},
	{CanonicalID: "glm-4-plus", DisplayName: "GLM 4 Plus", Provider: "zhipuai", Family: "glm"},
	{CanonicalID: "glm-4-flash", DisplayName: "GLM 4 Flash", Provider: "zhipuai", Family: "glm"},

	// xAI
	{CanonicalID: "grok-3", DisplayName: "Grok 3", Provider: "xai", Family: "grok"},
	{CanonicalID: "grok-2", DisplayName: "Grok 2", Provider: "xai", Family: "grok"},

	// Qwen
	{CanonicalID: "qwen-2.5-coder-32b-instruct", DisplayName: "Qwen 2.5 Coder 32B", Provider: "alibaba", Family: "qwen"},

	// Meta
	{CanonicalID: "llama-3.3-70b-instruct", DisplayName: "Llama 3.3 70B", Provider: "meta", Family: "llama"},

	// Mistral
	{CanonicalID: "codestral-latest", DisplayName: "Codestral", Provider: "mistralai", Family: "mistral"},
}

var identityIndex = func() map[string]ModelIdentity {
	idx := make(map[string]ModelIdentity, len(CanonicalIdentities))
	for _, id := range CanonicalIdentities {
		idx[id.CanonicalID] = id
	}
	return idx
}()

// FindModelIdentity resolves a model name to its canonical ModelIdentity if known.
func FindModelIdentity(model string) (ModelIdentity, bool) {
	key := strings.ToLower(strings.TrimSpace(model))
	if id, ok := identityIndex[key]; ok {
		return id, true
	}
	if alias, ok := variantAliases[key]; ok {
		if id, ok := identityIndex[alias]; ok {
			return id, true
		}
	}
	return ModelIdentity{}, false
}
