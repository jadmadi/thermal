// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import "fmt"

// Shared cost-marker definitions. The leaderboard and live overlay must use the
// same vocabulary so a reader never has to relearn the glyphs.
const (
	CostMarkerRecorded  = "$"
	CostMarkerEstimated = "~$"
	CostMarkerPartial   = "≥$"
	CostMarkerAbsent    = "—"
	CostMarkerApprox    = "≈"
)

// CostLegendText is the shared one-line cost legend.
func CostLegendText() string {
	return fmt.Sprintf("%s recorded · %s estimated · %s partial known · %s unpriced · %s approximate · $0.00 recorded zero",
		CostMarkerRecorded, CostMarkerEstimated, CostMarkerPartial, CostMarkerAbsent, CostMarkerApprox)
}
