// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import "charm.land/bubbletea/v2"

func isPress(msg tea.Msg) (string, bool) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		return k.String(), true
	}
	return "", false
}
