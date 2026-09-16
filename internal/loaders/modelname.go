package loaders

import "strings"

// modelName canonicalizes a model id for grouping and display. Tools disagree
// on case (ZCode records GLM-5.3-Flash while models.dev lists glm-5.3-flash),
// so names are lowercased to keep the same model in one row across tools.
// Pricing lookups are case insensitive, so this only affects our own grouping.
func modelName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
