package thermal

import (
	"fmt"
	"strconv"
	"strings"
)

func CompactNumber(v int64) string {
	if v >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(v)/1_000_000_000)
	}
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(v)/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(v)/1_000)
	}
	return strconv.FormatInt(v, 10)
}

func PadRight(s string, width int) string {
	if n := width - len(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func PadLeft(s string, width int) string {
	if n := width - len(s); n > 0 {
		return strings.Repeat(" ", n) + s
	}
	return s
}

func FormatPath(p string) string {
	home := HomeDir()
	if home != "" && strings.HasPrefix(p, home+"/") {
		return "~" + p[len(home):]
	}
	return p
}
