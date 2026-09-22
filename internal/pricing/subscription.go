// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package pricing

import (
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

var standardPlans = []thermal.SubscriptionPlan{
	{
		ID:              "claude-pro",
		Name:            "Claude Pro ($20)",
		Type:            thermal.PlanTypeSubscription,
		MonthlyFee:      20.0,
		DailyTokenLimit: 15_000_000, // ~45 msgs/5h rolling window
		DefaultModel:    "claude-3-5-sonnet",
		Provider:        "Anthropic",
		Notes:           "Rolling 5-hour limit; throttles on multi-turn loops",
	},
	{
		ID:              "claude-max",
		Name:            "Claude Team/Max ($200)",
		Type:            thermal.PlanTypeSubscription,
		MonthlyFee:      200.0,
		DailyTokenLimit: 75_000_000, // ~5x pro
		DefaultModel:    "claude-3-5-sonnet",
		Provider:        "Anthropic",
		Notes:           "High rate limits for team/power users",
	},
	{
		ID:              "chatgpt-plus",
		Name:            "ChatGPT Plus ($20)",
		Type:            thermal.PlanTypeSubscription,
		MonthlyFee:      20.0,
		DailyTokenLimit: 20_000_000, // ~80 msgs/3h
		DefaultModel:    "gpt-4o",
		Provider:        "OpenAI",
		Notes:           "Standard 3-hour caps",
	},
	{
		ID:              "chatgpt-pro",
		Name:            "ChatGPT Pro ($200)",
		Type:            thermal.PlanTypeSubscription,
		MonthlyFee:      200.0,
		DailyTokenLimit: 150_000_000,
		DefaultModel:    "gpt-4o",
		Provider:        "OpenAI",
		Notes:           "High limit 4o + o1 reasoning access",
	},
	{
		ID:              "cursor-pro",
		Name:            "Cursor Pro ($20)",
		Type:            thermal.PlanTypeSubscription,
		MonthlyFee:      20.0,
		DailyTokenLimit: 25_000_000, // 500 fast req/mo (~16/day)
		DefaultModel:    "claude-3-5-sonnet",
		Provider:        "Cursor",
		Notes:           "500 fast requests/mo then slow pool",
	},
	{
		ID:              "deepseek-api",
		Name:            "DeepSeek V3 (API)",
		Type:            thermal.PlanTypePayAsYouGo,
		MonthlyFee:      0.0,
		DailyTokenLimit: 0,
		DefaultModel:    "deepseek-chat",
		Provider:        "DeepSeek",
		Notes:           "Pay-as-you-go with 90%+ prompt cache discount",
	},
	{
		ID:              "claude-api",
		Name:            "Claude 3.5 Sonnet (API)",
		Type:            thermal.PlanTypePayAsYouGo,
		MonthlyFee:      0.0,
		DailyTokenLimit: 0,
		DefaultModel:    "claude-3-5-sonnet",
		Provider:        "Anthropic",
		Notes:           "Pay-as-you-go with prompt cache discount",
	},
	{
		ID:              "openai-api",
		Name:            "GPT-4o (API)",
		Type:            thermal.PlanTypePayAsYouGo,
		MonthlyFee:      0.0,
		DailyTokenLimit: 0,
		DefaultModel:    "gpt-4o",
		Provider:        "OpenAI",
		Notes:           "Pay-as-you-go with cached input pricing",
	},
}

// StandardPlans returns a copy of all predefined subscription and API plans.
func StandardPlans() []thermal.SubscriptionPlan {
	return append([]thermal.SubscriptionPlan{}, standardPlans...)
}

// DefaultComparisonPlans returns the default plan IDs compared in replay mode.
func DefaultComparisonPlans() []string {
	return []string{"claude-pro", "claude-max", "cursor-pro", "deepseek-api", "claude-api"}
}

// LookupPlan resolves a plan by ID or prefix case-insensitively.
func LookupPlan(key string) (thermal.SubscriptionPlan, bool) {
	norm := strings.ToLower(strings.TrimSpace(key))
	if norm == "" {
		return thermal.SubscriptionPlan{}, false
	}
	for _, p := range standardPlans {
		if strings.ToLower(p.ID) == norm || strings.ToLower(p.Name) == norm {
			return p, true
		}
	}
	for _, p := range standardPlans {
		if strings.HasPrefix(strings.ToLower(p.ID), norm) {
			return p, true
		}
	}
	return thermal.SubscriptionPlan{}, false
}
