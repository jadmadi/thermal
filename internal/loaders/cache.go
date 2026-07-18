package loaders

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jadmadi/thermal/internal/thermal"
)

// DevinCache is a disk-backed snapshot of the expensive message_nodes
// aggregation. Invalidation is keyed on MAX(row_id) (covers new appends —
// message_nodes is append-only in practice) plus the visible session count
// (covers hidden/unhidden sessions). Both probes hit PK/stat indexes and
// complete in <2ms, so a warm `thermal` run skips the ~11s full scan.
type DevinCache struct {
	MaxRowID     int64             `json:"maxRowId"`
	SessionCount int               `json:"sessionCount"`
	Summary      thermal.Summary   `json:"summary"`
	Daily        []thermal.DailyRow `json:"daily"`
}

func devinCachePath() (string, error) {
	home := thermal.HomeDir()
	dir := filepath.Join(home, ".cache", "thermal")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "devin.json"), nil
}

func loadDevinCache() (DevinCache, bool) {
	p, err := devinCachePath()
	if err != nil {
		return DevinCache{}, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return DevinCache{}, false
	}
	var c DevinCache
	if err := json.Unmarshal(b, &c); err != nil {
		return DevinCache{}, false
	}
	return c, true
}

func saveDevinCache(c DevinCache) {
	p, err := devinCachePath()
	if err != nil {
		return
	}
	// Best-effort; cache misses just mean a slow path next time.
	b, err := json.Marshal(c)
	if err != nil {
		return
	}
	_ = os.WriteFile(p, b, 0o644)
}
