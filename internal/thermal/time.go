package thermal

import (
	"os"
	"time"
)

func HomeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

func LocalDay(t time.Time) string {
	return t.Format("2006-01-02")
}

func StartOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
