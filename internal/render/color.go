// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"os"
)

var reset = "\033[0m"

func ColorCode(enabled bool, code string, value string) string {
	if enabled {
		return "\033[" + code + "m" + value + reset
	}
	return value
}

func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
