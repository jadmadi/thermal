// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package pricing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SourceURL is the models.dev catalog, the same registry OpenCode uses for
// provider metadata.
const SourceURL = "https://models.dev/api.json"

const fetchTimeout = 10 * time.Second

type modelsDevProvider struct {
	Models map[string]modelsDevModel `json:"models"`
}

type modelsDevModel struct {
	Cost *modelsDevCost `json:"cost"`
}

type modelsDevCost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cache_read"`
	CacheWrite float64 `json:"cache_write"`
}

// variantAliases maps names that tools record onto the catalog id they mirror.
// models.dev lists the preview id while some CLIs record a variant suffix.
var variantAliases = map[string]string{
	"gemini-3-pro-high": "gemini-3-pro-preview",
	"gemini-3-pro-low":  "gemini-3-pro-preview",
	"gemini-3.8-flash":  "gemini-2.5-flash",
	"gemini-3.7-flash":  "gemini-2.5-flash",
	"gemini-3.6-flash":  "gemini-2.5-flash",
	"gemini-3.1-pro":    "gemini-2.5-pro",
	"claude-3-5-sonnet": "claude-3.5-sonnet-v2",
	"claude-3.5-sonnet": "claude-3.5-sonnet-v2",
	"claude-3-7-sonnet": "claude-3.7-sonnet",
	"claude-sonnet":     "claude-3.7-sonnet",
	"deepseek-chat":     "deepseek-v3",
}

// preferredProviders orders first-party catalogs ahead of resellers. Prices
// differ between resellers for the same model, so estimates use list prices
// where a first-party entry exists.
var preferredProviders = []string{
	"openai", "anthropic", "google", "deepseek", "zhipuai", "zai", "xai",
	"mistralai", "moonshotai", "alibaba", "minimax", "meta", "cohere",
	"opencode", "openrouter", "github-copilot", "azure", "amazon-bedrock",
}

func isPlanProvider(id string) bool {
	return strings.Contains(id, "-plan")
}

// providerRank orders providers for price selection. Plan providers come last
// because they list subscription pricing as zero.
func providerRank(id string) int {
	if isPlanProvider(id) {
		return 100000
	}
	for i, p := range preferredProviders {
		if p == id {
			return i
		}
	}
	return len(preferredProviders) + 1
}

// fetchCatalog downloads and normalizes the models.dev catalog.
func fetchCatalog() (*Catalog, error) {
	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(SourceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pricing: %s returned %s", SourceURL, resp.Status)
	}

	var providers map[string]modelsDevProvider
	if err := json.NewDecoder(resp.Body).Decode(&providers); err != nil {
		return nil, err
	}
	return &Catalog{
		Version:   CacheVersion,
		FetchedAt: time.Now(),
		Source:    SourceURL,
		Models:    buildCatalog(providers),
	}, nil
}

// buildCatalog flattens the provider tree into one price per model id. When
// several providers list the same model, the better-ranked provider wins and
// ties break alphabetically so the result is deterministic.
func buildCatalog(providers map[string]modelsDevProvider) map[string]Price {
	out := make(map[string]Price)
	type chosen struct {
		rank     int
		provider string
	}
	picked := make(map[string]chosen)

	for pid, prov := range providers {
		rank := providerRank(pid)
		for mid, model := range prov.Models {
			if model.Cost == nil {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(mid))
			if key == "" {
				continue
			}
			// A subscription plan lists zero because usage is bundled, not
			// free. When a plan is the only source for a model, leave it
			// unpriceable so reports say so instead of showing $0.
			if isPlanProvider(pid) && model.Cost.Input == 0 && model.Cost.Output == 0 &&
				model.Cost.CacheRead == 0 && model.Cost.CacheWrite == 0 {
				continue
			}
			if cur, ok := picked[key]; ok && !betterChoice(rank, pid, cur.rank, cur.provider) {
				continue
			}
			picked[key] = chosen{rank: rank, provider: pid}
			out[key] = Price{
				Input:      model.Cost.Input,
				Output:     model.Cost.Output,
				CacheRead:  model.Cost.CacheRead,
				CacheWrite: model.Cost.CacheWrite,
			}.withFallbacks()
		}
	}

	for alias, target := range variantAliases {
		if p, ok := out[strings.ToLower(target)]; ok {
			out[alias] = p
		}
	}
	return out
}

func betterChoice(rank int, provider string, curRank int, curProvider string) bool {
	if rank != curRank {
		return rank < curRank
	}
	return provider < curProvider
}
