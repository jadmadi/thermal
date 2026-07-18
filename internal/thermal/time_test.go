package thermal

import (
	"testing"
	"time"
)

func TestUnixDay(t *testing.T) {
	// 1710504000 is 2024-03-15 12:00:00 UTC
	ts := int64(1710504000)
	got := UnixDay(ts)
	want := time.Unix(ts, 0).Format("2006-01-02")
	if got != want {
		t.Errorf("UnixDay(%d) = %q, want %q", ts, got, want)
	}
}

func TestStartOfToday(t *testing.T) {
	today := StartOfToday()
	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 || today.Nanosecond() != 0 {
		t.Errorf("StartOfToday not at midnight: %v", today)
	}
}
