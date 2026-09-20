// Package pricing estimates token cost from the models.dev catalog. Stored
// cost recorded by a source always wins; this package only fills the gaps and
// reports the models it cannot price.
package pricing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// CacheVersion bumps whenever the on-disk catalog shape changes.
const CacheVersion = 1

const cacheTTL = 24 * time.Hour

// Price is the USD cost per million tokens for one model.
type Price struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead,omitempty"`
	CacheWrite float64 `json:"cacheWrite,omitempty"`
}

// cacheReadRate falls back to a tenth of the input rate, the common cache
// discount, when the catalog omits the field.
func (p Price) cacheReadRate() float64 {
	if p.CacheRead > 0 {
		return p.CacheRead
	}
	return p.Input * 0.1
}

func (p Price) cacheWriteRate() float64 {
	if p.CacheWrite > 0 {
		return p.CacheWrite
	}
	return p.Input
}

// Catalog maps lowercase model ids to prices. It implements thermal.Pricer.
type Catalog struct {
	Version   int              `json:"version"`
	FetchedAt time.Time        `json:"fetchedAt"`
	Source    string           `json:"source"`
	Models    map[string]Price `json:"models"`
}

// Len returns the number of priced models.
func (c *Catalog) Len() int {
	if c == nil {
		return 0
	}
	return len(c.Models)
}

// DefaultCachePath returns ~/.cache/thermal/pricing.json.
func DefaultCachePath() string {
	return filepath.Join(thermal.HomeDir(), ".cache", "thermal", "pricing.json")
}

// DefaultOverridesPath returns ~/.config/thermal/pricing.json. The file holds
// a model id to Price map that wins over catalog entries.
func DefaultOverridesPath() string {
	return filepath.Join(thermal.HomeDir(), ".config", "thermal", "pricing.json")
}

// Load returns a usable catalog. It reads the on-disk cache, layers local
// overrides on top, and refreshes from models.dev when the cache is missing or
// stale. Offline skips the network. A failed refresh keeps the cached copy, so
// Load never blocks the caller and never returns nil.
func Load(cachePath string, offline bool) *Catalog {
	cat := readCache(cachePath)
	if cat == nil {
		cat = &Catalog{Version: CacheVersion, Models: map[string]Price{}}
	}
	applyOverrides(cat, DefaultOverridesPath())

	stale := cat.FetchedAt.IsZero() || time.Since(cat.FetchedAt) > cacheTTL
	if offline || !stale {
		return cat
	}

	fresh, err := fetchCatalog()
	if err != nil {
		// A failed refresh keeps whatever the cache and overrides provide.
		return cat
	}
	// Persist before layering overrides so a user price never becomes part of
	// the shared cache and outlives its own file.
	writeCache(cachePath, fresh)
	applyOverrides(fresh, DefaultOverridesPath())
	return fresh
}

// NewCatalog builds a catalog from an in-memory price map. Tests and the
// overrides file use it.
func NewCatalog(models map[string]Price) *Catalog {
	normalized := make(map[string]Price, len(models))
	for id, p := range models {
		normalized[strings.ToLower(strings.TrimSpace(id))] = p.withFallbacks()
	}
	return &Catalog{Version: CacheVersion, FetchedAt: time.Now(), Source: SourceURL, Models: normalized}
}

// Lookup resolves a model name to a price. Names are matched case
// insensitively, with provider path segments and date or variant suffixes
// stripped.
func (c *Catalog) Lookup(model string) (Price, bool) {
	if c == nil || len(c.Models) == 0 {
		return Price{}, false
	}
	for _, key := range lookupCandidates(model) {
		if p, ok := c.Models[key]; ok {
			return p, true
		}
	}
	return Price{}, false
}

// PriceDay implements thermal.Pricer. It prices each model from its token
// types and returns the models that have no usable price. A model with
// unclassified tokens is left unpriced rather than half-priced, because no
// rate matches those tokens.
func (c *Catalog) PriceDay(day thermal.DailyRow) (float64, []string) {
	if c == nil || len(day.Models) == 0 {
		return 0, nil
	}

	const million = 1_000_000.0
	var cost float64
	var missing []string

	for model, t := range day.Models {
		if t.Total() == 0 {
			continue
		}
		p, ok := c.Lookup(model)
		if !ok || t.Unclassified > 0 {
			missing = append(missing, model)
			continue
		}
		cost += float64(t.Input) * p.Input / million
		cost += float64(t.Output+t.Reasoning) * p.Output / million
		cost += float64(t.CacheRead) * p.cacheReadRate() / million
		cost += float64(t.CacheWrite) * p.cacheWriteRate() / million
	}
	sort.Strings(missing)
	return cost, missing
}

// lookupCandidates lists the model name shapes to try in order, most specific
// first.
func lookupCandidates(model string) []string {
	base := strings.ToLower(strings.TrimSpace(model))
	if base == "" {
		return nil
	}

	seen := make(map[string]bool)
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}

	add(base)
	if i := strings.LastIndex(base, "/"); i >= 0 {
		add(base[i+1:])
	}
	for _, key := range append([]string{}, out...) {
		if i := strings.Index(key, ":"); i >= 0 {
			add(key[:i])
		}
		add(trimDateSuffix(key))
		add(strings.TrimSuffix(key, "-latest"))
		add(trimTierSuffix(key))
	}
	return out
}

// tierSuffixes are the effort and tier words tools append to a model id. A tool
// records the variant it asked for while the catalog prices the base model, so
// dropping the suffix is what turns claude-sonnet-5-high into claude-sonnet-5.
//
// The list is deliberately short and specific. A general stem match would fold
// two different models onto one price, which is worse than naming a model as
// unpriced.
var tierSuffixes = []string{
	"-high", "-medium", "-low", "-max", "-mini", "-nano",
	"-thinking", "-reasoning", "-preview", "-fast", "-eco",
}

// trimTierSuffix removes one effort or tier word from the end of an id. It
// strips at most one, so a model genuinely called something-low-high would need
// a mapping rather than a guess.
func trimTierSuffix(s string) string {
	for _, suffix := range tierSuffixes {
		if strings.HasSuffix(s, suffix) && len(s) > len(suffix) {
			return s[:len(s)-len(suffix)]
		}
	}
	return s
}

// trimDateSuffix removes a trailing -YYYYMMDD build stamp.
func trimDateSuffix(s string) string {
	if len(s) < 9 || s[len(s)-9] != '-' {
		return s
	}
	for _, ch := range s[len(s)-8:] {
		if ch < '0' || ch > '9' {
			return s
		}
	}
	return s[:len(s)-9]
}

func readCache(path string) *Catalog {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cat Catalog
	if err := json.Unmarshal(b, &cat); err != nil {
		return nil
	}
	if cat.Version != CacheVersion || len(cat.Models) == 0 {
		return nil
	}
	for id, p := range cat.Models {
		cat.Models[id] = p.withFallbacks()
	}
	return &cat
}

func writeCache(path string, cat *Catalog) {
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	b, err := json.Marshal(cat)
	if err != nil {
		return
	}
	// Atomic replace so a torn write never leaves a half catalog behind.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

// withFallbacks fills the cache rates that models.dev sometimes omits.
func (p Price) withFallbacks() Price {
	if p.CacheRead == 0 {
		p.CacheRead = p.Input * 0.1
	}
	if p.CacheWrite == 0 {
		p.CacheWrite = p.Input
	}
	return p
}

// applyOverrides layers a user price map over the catalog. Overrides win.
func applyOverrides(cat *Catalog, path string) {
	if cat == nil {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var overrides map[string]Price
	if err := json.Unmarshal(b, &overrides); err != nil {
		return
	}
	if cat.Models == nil {
		cat.Models = make(map[string]Price)
	}
	for id, p := range overrides {
		cat.Models[strings.ToLower(strings.TrimSpace(id))] = p.withFallbacks()
	}
}
