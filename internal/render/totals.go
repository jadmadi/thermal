// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"strings"
)

func totalToolsCell(count, width int) string {
	if count <= 0 {
		return strings.Repeat(" ", width)
	}
	return padRight(fmt.Sprintf("%d tools", count), width)
}

func projectDaysCell(days, width int) string {
	if days <= 0 {
		return strings.Repeat(" ", width)
	}
	return padLeft(fmt.Sprintf("%d", days), width)
}

func totalCostText(text string) string {
	if text == "—" {
		return "$0.00"
	}
	return text
}

func modelDaysCell(days, width int) string {
	if days <= 0 {
		return strings.Repeat(" ", width)
	}
	return padLeft(fmt.Sprintf("%d", days), width)
}
