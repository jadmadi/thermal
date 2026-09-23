// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import "strings"

// CanonicalModel represents a declared standard model identifier.
type CanonicalModel string

// canonicalModels holds known canonical model identifiers.
var canonicalModels = map[string]bool{
	// Anthropic
	"claude-3.7-sonnet": true,
	"claude-3.5-sonnet": true,
	"claude-3.5-haiku":  true,
	"claude-3-opus":     true,
	"claude-3-sonnet":   true,
	"claude-3-haiku":    true,
	"claude-opus":       true,
	"claude-sonnet":     true,
	"claude-haiku":      true,

	// OpenAI
	"gpt-4o":        true,
	"gpt-4o-mini":   true,
	"gpt-4-turbo":   true,
	"gpt-4":         true,
	"gpt-3.5-turbo": true,
	"o1":            true,
	"o1-preview":    true,
	"o1-mini":       true,
	"o3-mini":       true,
	"o3":            true,
	"o4-mini":       true,

	// DeepSeek
	"deepseek-v3":    true,
	"deepseek-r1":    true,
	"deepseek-coder": true,

	// Google
	"gemini-2.5-pro":            true,
	"gemini-2.5-flash":          true,
	"gemini-2.0-flash":          true,
	"gemini-2.0-flash-thinking": true,
	"gemini-2.0-pro":            true,
	"gemini-1.5-pro":            true,
	"gemini-1.5-flash":          true,
	"gemini-1.5-flash-8b":       true,
	"gemini-3-pro-preview":      true,

	// Zhipu / GLM
	"glm-5.3-flash": true,
	"glm-4-plus":    true,
	"glm-4-flash":   true,
	"glm-4-air":     true,

	// xAI
	"grok-3":      true,
	"grok-3-mini": true,
	"grok-2":      true,
	"grok-2-mini": true,

	// Qwen / Alibaba
	"qwen-2.5-coder-32b-instruct": true,
	"qwen-2.5-coder-7b-instruct":  true,
	"qwen-max":                    true,
	"qwen-plus":                   true,
	"qwen-turbo":                  true,

	// Meta
	"llama-3.3-70b-instruct":  true,
	"llama-3.1-405b-instruct": true,
	"llama-3.1-70b-instruct":  true,
	"llama-3.1-8b-instruct":   true,

	// Mistral
	"codestral-latest":     true,
	"mistral-large-latest": true,
	"mistral-small-latest": true,
}

// declaredAliases maps router aliases, dated tags, and tool-specific variants
// onto canonical model names.
var declaredAliases = map[string]string{
	// DeepSeek
	"deepseek-chat":              "deepseek-v3",
	"deepseek/deepseek-chat":     "deepseek-v3",
	"deepseek-reasoner":          "deepseek-r1",
	"deepseek/deepseek-reasoner": "deepseek-r1",
	"deepseek-coder-v2":          "deepseek-coder",
	"deepseek-coder-v2.5":        "deepseek-coder",

	// Anthropic
	"claude-3-7-sonnet":          "claude-3.7-sonnet",
	"claude-3-7-sonnet-20250219": "claude-3.7-sonnet",
	"claude-3.7-sonnet-20250219": "claude-3.7-sonnet",
	"claude-3-7-sonnet-latest":   "claude-3.7-sonnet",
	"claude-3-5-sonnet":          "claude-3.5-sonnet",
	"claude-3-5-sonnet-20241022": "claude-3.5-sonnet",
	"claude-3.5-sonnet-20241022": "claude-3.5-sonnet",
	"claude-3-5-sonnet-20240620": "claude-3.5-sonnet",
	"claude-3.5-sonnet-20240620": "claude-3.5-sonnet",
	"claude-3-5-sonnet-v2":       "claude-3.5-sonnet",
	"claude-3.5-sonnet-v2":       "claude-3.5-sonnet",
	"claude-3-5-haiku":           "claude-3.5-haiku",
	"claude-3-5-haiku-20241022":  "claude-3.5-haiku",
	"claude-3.5-haiku-20241022":  "claude-3.5-haiku",
	"claude-3-opus-20240229":     "claude-3-opus",
	"claude-3-sonnet-20240229":   "claude-3-sonnet",
	"claude-3-haiku-20240307":    "claude-3-haiku",

	// OpenAI
	"chatgpt-4o-latest":      "gpt-4o",
	"gpt-4o-latest":          "gpt-4o",
	"gpt-4o-2024-05-13":      "gpt-4o",
	"gpt-4o-2024-08-06":      "gpt-4o",
	"gpt-4o-2024-11-20":      "gpt-4o",
	"gpt-4o-mini-2024-07-18": "gpt-4o-mini",
	"o1-2024-12-17":          "o1",
	"o1-preview-2024-09-12":  "o1-preview",
	"o1-mini-2024-09-12":     "o1-mini",
	"o3-mini-2025-01-31":     "o3-mini",
	"o3-mini-high":           "o3-mini",
	"o3-mini-medium":         "o3-mini",
	"o3-mini-low":            "o3-mini",

	// Google
	"gemini-3-pro-high":    "gemini-3-pro-preview",
	"gemini-3-pro-low":     "gemini-3-pro-preview",
	"gemini-2.0-flash-001": "gemini-2.0-flash",
	"gemini-2.0-flash-exp": "gemini-2.0-flash",
	"gemini-1.5-pro-001":   "gemini-1.5-pro",
	"gemini-1.5-pro-002":   "gemini-1.5-pro",
	"gemini-1.5-flash-001": "gemini-1.5-flash",
	"gemini-1.5-flash-002": "gemini-1.5-flash",

	// Qwen
	"qwen-2.5-coder-32b":         "qwen-2.5-coder-32b-instruct",
	"qwen2.5-coder:32b":          "qwen-2.5-coder-32b-instruct",
	"qwen2.5-coder-32b-instruct": "qwen-2.5-coder-32b-instruct",

	// Meta
	"llama-3.3-70b":                     "llama-3.3-70b-instruct",
	"meta-llama/llama-3.3-70b-instruct": "llama-3.3-70b-instruct",

	// Mistral
	"codestral":     "codestral-latest",
	"mistral-large": "mistral-large-latest",
	"mistral-small": "mistral-small-latest",
}

// trimSpaceASCII removes ASCII whitespace without heap allocation if unchanged.
func trimSpaceASCII(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// isLowerASCII returns true if all characters are lowercase ASCII or non-letters.
func isLowerASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
}

// stripRouterPrefix removes common router/gateway prefixes.
func stripRouterPrefix(s string) string {
	for _, p := range [...]string{"openrouter/", "litellm/", "proxy/", "router/"} {
		if strings.HasPrefix(s, p) {
			return s[len(p):]
		}
	}
	return s
}

// trimDateSuffix removes a trailing -YYYY-MM-DD or -YYYYMMDD date stamp.
func trimDateSuffix(s string) string {
	// Match -YYYY-MM-DD (11 chars)
	if len(s) >= 11 && s[len(s)-11] == '-' && s[len(s)-6] == '-' && s[len(s)-3] == '-' {
		isDate := true
		for _, idx := range [...]int{10, 9, 8, 7, 5, 4, 2, 1} {
			c := s[len(s)-idx]
			if c < '0' || c > '9' {
				isDate = false
				break
			}
		}
		if isDate {
			return s[:len(s)-11]
		}
	}
	// Match -YYYYMMDD (9 chars)
	if len(s) >= 9 && s[len(s)-9] == '-' {
		isDate := true
		for _, c := range s[len(s)-8:] {
			if c < '0' || c > '9' {
				isDate = false
				break
			}
		}
		if isDate {
			return s[:len(s)-9]
		}
	}
	return s
}

// CanonicalModelName normalizes a model string to its canonical representation.
// It implements the 4-tier resolution priority:
// 1. Exact ID -> 2. Canonical Name -> 3. Declared Harness-Scoped Alias -> 4. Unresolved explicit fallback.
func CanonicalModelName(s string) string {
	return modelName(s)
}

// modelName canonicalizes a model id for grouping and display. Tools disagree
// on case (ZCode records GLM-5.3-Flash while models.dev lists glm-5.3-flash),
// routing prefixes, and date-stamped variants.
func modelName(s string) string {
	trimmed := trimSpaceASCII(s)
	if trimmed == "" {
		return ""
	}

	base := trimmed
	if !isLowerASCII(base) {
		base = strings.ToLower(base)
	}

	// 1. Exact canonical hit on raw string
	if canonicalModels[base] {
		return base
	}

	// 2. Exact alias hit on raw string
	if alias, ok := declaredAliases[base]; ok {
		return alias
	}

	// 3. Router and provider prefix normalization
	cleaned := stripRouterPrefix(base)
	leaf := cleaned
	if idx := strings.LastIndexByte(cleaned, '/'); idx >= 0 {
		leaf = cleaned[idx+1:]
	}

	if canonicalModels[cleaned] {
		return cleaned
	}
	if alias, ok := declaredAliases[cleaned]; ok {
		return alias
	}

	if canonicalModels[leaf] {
		return leaf
	}
	if alias, ok := declaredAliases[leaf]; ok {
		return alias
	}

	// 4. Date suffix stripping
	if trimmedDate := trimDateSuffix(leaf); trimmedDate != leaf {
		if canonicalModels[trimmedDate] {
			return trimmedDate
		}
		if alias, ok := declaredAliases[trimmedDate]; ok {
			return alias
		}
	}

	// 5. -latest suffix stripping
	if strings.HasSuffix(leaf, "-latest") && len(leaf) > 7 {
		withoutLatest := leaf[:len(leaf)-7]
		if canonicalModels[withoutLatest] {
			return withoutLatest
		}
		if alias, ok := declaredAliases[withoutLatest]; ok {
			return alias
		}
	}

	// 6. Explicit fallback (unresolved)
	if leaf != "" {
		return leaf
	}
	return base
}
