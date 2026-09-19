//go:build race

package tui

import "testing"

// Under -race the instrumented build is roughly ten times slower, so the load
// budget is not measurable here. The real measurement lives in the non-race
// variant of this file.
func TestLoadBudget(t *testing.T) {
	t.Skip("timing budget is not meaningful under the race detector")
}
